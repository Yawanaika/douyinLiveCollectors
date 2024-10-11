package collectors

import (
	"douyinLiveCollectors/lib/handler"
	"douyinLiveCollectors/lib/sign"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	WebSocketConnectError          = "WebSocketConnectError: %v"
	WebSocketError                 = "WebSocketError: %v"
	FailedToGetTtwIdError          = "FailedToGetTtwIdError: %v"
	FailedToRequestLiveRoomError   = "FailedToRequestLiveRoomError: %v"
	FailedToReadResponseBodyError  = "FailedToReadResponseBodyError: %v"
	FailedToGenerateSignatureError = "FailedToGenerateSignatureError: %v"
	FailedToSendRequestError       = "FailedToSendRequestError: %v"
)

var (
	RoomIdNotFound = errors.New("roomId not found in response")
)

type LiveViewer struct {
	liveId    int
	ttwId     string
	roomId    string
	liveUrl   string
	userAgent string
	ws        *websocket.Conn
}

func NewLiveViewer(liveId int) *LiveViewer {

	return &LiveViewer{
		liveId:    liveId,
		liveUrl:   sign.Url,
		userAgent: sign.UserAgent,
	}
}

func (v *LiveViewer) Start() {
	v.setTtwId().setRoomID()
	wss := fmt.Sprintf("wss://webcast5-ws-web-hl.douyin.com/webcast/im/push/v2/?app_name=douyin_web"+
		"&version_code=180800&webcast_sdk_version=1.0.14-beta.0"+
		"&update_version_code=1.0.14-beta.0&compress=gzip&device_platform=web&cookie_enabled=true"+
		"&screen_width=1536&screen_height=864&browser_language=zh-CN&browser_platform=Win32"+
		"&browser_name=Mozilla"+
		"&browser_version=5.0%%20(Windows%%20NT%%2010.0;%%20Win64;%%20x64)%%20AppleWebKit/537.36%%20(KHTML,"+
		"%%20like%%20Gecko)%%20Chrome/126.0.0.0%%20Safari/537.36"+
		"&browser_online=true&tz_name=Asia/Shanghai"+
		"&cursor=d-1_u-1_fh-7392091211001140287_t-1721106114633_r-1"+
		"&internal_ext=internal_src:dim|wss_push_room_id:%s|wss_push_did:7319483754668557238"+
		"|first_req_ms:1721106114541|fetch_time:1721106114633|seq:1|wss_info:0-1721106114633-0-0|"+
		"wrds_v:7392094459690748497"+
		"&host=https://live.douyin.com&aid=6383&live_id=1&did_rule=3&endpoint=live_pc&support_wrds=1"+
		"&user_unique_id=7319483754668557238&im_path=/webcast/im/fetch/&identity=audience"+
		"&need_persist_msg_count=15&insert_task_id=&live_reason=&room_id=%s&heartbeatDuration=0", v.roomId, v.roomId)

	_, err := sign.GenerateSignature(wss)
	if err != nil {
		log.Fatalf(FailedToGenerateSignatureError, err)
	}
	wss += fmt.Sprintf("&signature=%v", "60REnIok7wNZFPjJ")
	//wss += fmt.Sprintf("&signature=%v", signature)

	headers := http.Header{
		"Cookie":     []string{fmt.Sprintf("ttwid=%s", v.ttwId)},
		"User-Agent": []string{v.userAgent},
	}

	var dialer websocket.Dialer
	v.ws, _, err = dialer.Dial(wss, headers)
	if err != nil {
		v.stop()
		log.Printf(WebSocketConnectError, err)
	} else {
		log.Println("Websocket connected.")
		go v.listen()
	}
}

func (v *LiveViewer) listen() {
	defer v.ws.Close()
	for {
		_, messages, err := v.ws.ReadMessage()
		if err != nil {
			log.Printf(WebSocketError, err)
			v.stop()
			break
		}
		handler.Handler(v.ws, messages)
	}
}

func (v *LiveViewer) stop() {
	if v.ws != nil {
		v.ws.Close()
	}
	log.Println("WebSocket connection closed.")
}

func (v *LiveViewer) setRoomID() *LiveViewer {
	url := v.liveUrl + strconv.Itoa(v.liveId)

	headers := map[string]string{
		"User-Agent": v.userAgent,
		"Cookie":     fmt.Sprintf("ttwid=%v; msToken=%v; __ac_nonce=0123407cc00a9e438deb4", v.ttwId, sign.GenerateMsToken()),
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf(sign.FailedToCreateRequestError, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf(FailedToRequestLiveRoomError, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(FailedToReadResponseBodyError, err)
	}

	// 使用正则表达式查找 roomID
	roomIdrRe := regexp.MustCompile(`roomId\\":\\"(\d+)\\"`)
	matches := roomIdrRe.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		log.Println(RoomIdNotFound)
	}
	v.roomId = matches[1]
	return v
}

func (v *LiveViewer) setTtwId() *LiveViewer {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", v.liveUrl, nil)

	req.Header.Set("User-Agent", "v.userAgent")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf(FailedToSendRequestError, err)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "ttwid" {
			v.ttwId = cookie.Value
		} else {
			log.Printf(FailedToGetTtwIdError, err)
		}
	}
	return v
}
