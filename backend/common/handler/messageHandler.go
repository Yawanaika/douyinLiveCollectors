package handler

import (
	"bytes"
	"compress/gzip"
	"douyinLiveCollectors/backend/common/enums"
	"douyinLiveCollectors/backend/common/log"
	"douyinLiveCollectors/backend/common/message"
	"douyinLiveCollectors/backend/library/time"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const (
	SendAckError                 = "SendAckError: %v"
	ParsePushFrameError          = "ParsePushFrameError: %v"
	DecompressPayloadError       = "DecompressPayloadError: %v"
	ParseResponseError           = "ParseResponseError: %v"
	ParseChatMessageError        = "ParseChatMessageError: %v"
	ParseGiftMessageError        = "ParseGiftMessageError: %v"
	ParseMemberMessageError      = "ParseMemberMessageError: %v"
	ParseLikeMessageError        = "ParseLikeMessageError: %v"
	ParseSocialMessageError      = "ParseSocialMessageError: %v"
	ParseRoomUserSeqMessageError = "ParseRoomUserSeqMessageError: %v"
	ParseFansclubMessageError    = "ParseFansclubMessageError: %v"
	ParseControlMessageError     = "ParseControlMessageError: %v"
	ParseEmojiChatMessageError   = "ParseEmojiChatMessageError: %v"
	ParseRoomStatsMessageError   = "ParseRoomStatsMessageError: %v"
	ParseRoomMessageError        = "ParseRoomMessageError: %v"
	ParseRoomRankMessageError    = "ParseRoomRankMessageError: %v"
	UnknownMessageError          = "UnknownMessageError: %v"
)

//const (
//	WebcastChatMessage        = "WebcastChatMessage"
//	WebcastGiftMessage        = "WebcastGiftMessage"
//	WebcastLikeMessage        = "WebcastLikeMessage"
//	WebcastMemberMessage      = "WebcastMemberMessage"
//	WebcastSocialMessage      = "WebcastSocialMessage"
//	WebcastRoomUserSeqMessage = "WebcastRoomUserSeqMessage"
//	WebcastFansclubMessage    = "WebcastFansclubMessage"
//	WebcastControlMessage     = "WebcastControlMessage"
//	WebcastEmojiChatMessage   = "WebcastEmojiChatMessage"
//	WebcastRoomStatsMessage   = "WebcastRoomStatsMessage"
//	WebcastRoomMessage        = "WebcastRoomMessage"
//	WebcastRoomRankMessage    = "WebcastRoomRankMessage"
//)

type Result struct {
	method string
	Result string
}

func Handler(ws *websocket.Conn, payload []byte, out chan<- Result) {
	resp := parseAndAck(ws, payload, out)
	go func() {
		for _, msg := range resp.GetMessagesList() {
			switch msg.GetMethod() {
			case enums.WebcastChatMessage:
				parseChatMessage(msg.GetPayload(), out)
			case enums.WebcastGiftMessage:
				parseGiftMessage(msg.GetPayload(), out)
			case enums.WebcastMemberMessage:
				parseMemberMessage(msg.GetPayload(), out)
			case enums.WebcastLikeMessage:
				parseLikeMessage(msg.GetPayload(), out)
			case enums.WebcastSocialMessage:
				parseSocialMessage(msg.GetPayload(), out)
			case enums.WebcastRoomUserSeqMessage:
				parseRoomUserSeqMessage(msg.GetPayload(), out)
			case enums.WebcastFansclubMessage:
				parseFansclubMessage(msg.GetPayload(), out)
			case enums.WebcastControlMessage:
				parseControlMessage(msg.GetPayload(), out)
				ws.Close()
			case enums.WebcastEmojiChatMessage:
				parseEmojiChatMessage(msg.GetPayload(), out)
			case enums.WebcastRoomStatsMessage:
				parseRoomStatsMessage(msg.GetPayload(), out)
			case enums.WebcastRoomMessage:
				parseRoomMessage(msg.GetPayload(), out)
			case enums.WebcastRoomRankMessage:
				parseRoomRankMessage(msg.GetPayload(), out)
			default:
				log.Info(UnknownMessageError, msg.String())
			}
		}
	}()
}

func parsePushFrame(payload []byte) (*message.PushFrame, error) {
	var frame message.PushFrame
	err := proto.Unmarshal(payload, &frame)
	if err != nil {
		return nil, err
	}
	return &frame, nil
}

func parseResponse(data []byte) (*message.Response, error) {
	var response message.Response
	err := proto.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func decompressGzip(payload []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return decompressed.Bytes(), nil
}

func sendAck(ws *websocket.Conn, LogId uint64, internalExt string) error {
	ack := &message.PushFrame{
		LogId:       LogId,
		PayloadType: "ack",
		Payload:     []byte(internalExt),
	}
	ackData, err := proto.Marshal(ack)
	if err != nil {
		return err
	}

	err = ws.WriteMessage(websocket.BinaryMessage, ackData)
	if err != nil {
		return fmt.Errorf(SendAckError, err)
	}
	return nil
}

func parseAndAck(ws *websocket.Conn, payload []byte, out chan<- Result) *message.Response {
	pkg, err := parsePushFrame(payload)
	if err != nil {
		log.Info(ParsePushFrameError, err.Error())
	}

	decompressedData, err := decompressGzip(pkg.Payload)
	if err != nil {
		log.Error(DecompressPayloadError, err.Error())
	}

	resp, err := parseResponse(decompressedData)
	if err != nil {
		log.Error(ParseResponseError, err.Error())
	}

	if resp.NeedAck {
		err = sendAck(ws, pkg.LogId, resp.InternalExt)
		if err != nil {
			log.Info(SendAckError, err.Error())
		}
		currentTime := time.Now()
		log.Info("Received ack message: ACK sent successfully")
		out <- Result{
			method: enums.WebcastRoomMessage,
			Result: fmt.Sprintf("%s  ACK sent successfully.", currentTime),
		}
	}
	return resp
}

func parseChatMessage(payload []byte, out chan<- Result) {
	var chat message.ChatMessage
	err := proto.Unmarshal(payload, &chat)
	if err != nil {
		log.Info(ParseChatMessageError, err.Error())
	}
	userName := chat.GetUser().GetNickName()
	userId := chat.GetUser().GetId()
	content := chat.GetContent()
	currentTime := time.ParseEventTime(chat.GetEventTime())
	log.Info("Received chat message : %s (ID: %d): %s", userName, userId, content)
	out <- Result{
		method: enums.WebcastChatMessage,
		Result: fmt.Sprintf("%s 【聊天消息】[ {%v} ] {%v} : {%v}", currentTime, userId, userName, content),
	}
}

func parseGiftMessage(payload []byte, out chan<- Result) {
	var gift message.GiftMessage
	err := proto.Unmarshal(payload, &gift)
	if err != nil {
		log.Info(ParseGiftMessageError, err.Error())
	}
	userName := gift.GetUser().GetNickName()
	toUser := gift.GetToUser().GetNickName()
	giftName := gift.GetGift().GetName()
	combo := gift.GetComboCount()
	currentTime := time.ParseEventTime(gift.GetCommon().GetCreateTime())
	log.Info("Received gift message : %s : to %s : %s X %v combo", userName, toUser, giftName, combo)
	out <- Result{
		method: enums.WebcastGiftMessage,
		Result: fmt.Sprintf("%s 【礼物消息】{%v} 给 {%s} 送出了 {%v} X {%v}连击", currentTime, userName, toUser, giftName, combo),
	}
}

func parseMemberMessage(payload []byte, out chan<- Result) {
	var member message.MemberMessage
	err := proto.Unmarshal(payload, &member)
	if err != nil {
		log.Info(ParseMemberMessageError, err.Error())
	}
	userId := member.GetUser().GetId()
	userName := member.GetUser().GetNickName()
	gender := []string{"女", "男", "unknown"}[member.GetUser().GetGender()]
	currentTime := time.ParseEventTime(member.GetCommon().GetCreateTime())
	log.Info("Received member message : %s (ID: %v, gender: %s) 进入了直播间", userName, userId, gender)
	out <- Result{
		method: enums.WebcastMemberMessage,
		Result: fmt.Sprintf("%s 【进场消息】[ {%v} ][ {%v} ] {%v} 进入了直播间", currentTime, userId, gender, userName),
	}
}

func parseRoomRankMessage(payload []byte, out chan<- Result) {
	var roomRank message.RoomRankMessage
	err := proto.Unmarshal(payload, &roomRank)
	if err != nil {
		log.Info(ParseRoomRankMessageError, err.Error())
	}
	ranksList := roomRank.GetRanksList()
	ranks := make(map[int]interface{}, 3)
	for i, rank := range ranksList {
		ranks[i] = map[string]interface{}{
			"Id": rank.GetUser().GetId(),
			"昵称": rank.GetUser().GetNickName(),
		}
	}
	currentTime := time.ParseEventTime(roomRank.GetCommon().GetCreateTime())
	log.Info("Received roomRank message : %v", ranks)
	out <- Result{
		method: enums.WebcastRoomRankMessage,
		Result: fmt.Sprintf("%s 【直播间排行榜消息】{%v}", currentTime, ranks),
	}
}

func parseRoomMessage(payload []byte, out chan<- Result) {
	var room message.RoomMessage
	err := proto.Unmarshal(payload, &room)
	if err != nil {
		log.Info(ParseRoomMessageError, err.Error())
	}
	roomId := room.GetCommon().GetRoomId()
	currentTime := time.ParseEventTime(room.GetCommon().GetCreateTime())
	log.Info("Received room message : 直播间id: %v", roomId)
	out <- Result{
		method: enums.WebcastRoomMessage,
		Result: fmt.Sprintf("%s 【直播间消息】直播间id: {%v}", currentTime, roomId),
	}
}

func parseRoomStatsMessage(payload []byte, out chan<- Result) {
	var roomStats message.RoomStatsMessage
	err := proto.Unmarshal(payload, &roomStats)
	if err != nil {
		log.Info(ParseRoomStatsMessageError, err.Error())
	}
	displayLong := roomStats.GetDisplayLong()
	currentTime := time.ParseEventTime(roomStats.GetCommon().GetCreateTime())
	log.Info("Received roomStates message : %v", displayLong)
	out <- Result{
		method: enums.WebcastRoomStatsMessage,
		Result: fmt.Sprintf("%s 【直播间统计消息】{%v}", currentTime, displayLong),
	}
}

func parseEmojiChatMessage(payload []byte, out chan<- Result) {
	var emoji message.EmojiChatMessage
	err := proto.Unmarshal(payload, &emoji)
	if err != nil {
		log.Info(ParseEmojiChatMessageError, err.Error())
	}
	emojiId := emoji.GetEmojiId()
	userName := emoji.GetUser().GetNickName()
	//common := emoji.GetCommon()
	defaultContent := emoji.GetDefaultContent()
	currentTime := time.ParseEventTime(emoji.GetCommon().GetCreateTime())
	log.Info("Received emojiChat message : %s : emojiId: %v,defaultContent: %s", userName, emojiId, defaultContent)
	out <- Result{
		method: enums.WebcastEmojiChatMessage,
		//Result: fmt.Sprintf("%s 【聊天表情包ID】 {%v},user：{%v},common:{%v},defaultContent:{%v}", currentTime, emojiId, userName, common, defaultContent),
		Result: fmt.Sprintf("%s 【聊天表情包ID】 {%v},user：{%v},defaultContent:{%v}", currentTime, emojiId, userName, defaultContent),
	}
}

func parseControlMessage(payload []byte, out chan<- Result) {
	var control message.ControlMessage
	err := proto.Unmarshal(payload, &control)
	if err != nil {
		log.Info(ParseControlMessageError, err.Error())
	}
	if control.GetStatus() == 3 {
		roomId := control.GetCommon().GetRoomId()
		currentTime := time.ParseEventTime(control.GetCommon().GetCreateTime())
		log.Info("Received control message : 直播间 %v 已结束", roomId)
		out <- Result{
			method: enums.WebcastControlMessage,
			Result: fmt.Sprintf("%s 【直播间消息】直播间 {%v} 已结束", currentTime, roomId),
		}
	}
}

func parseFansclubMessage(payload []byte, out chan<- Result) {
	var fansclub message.FansclubMessage
	err := proto.Unmarshal(payload, &fansclub)
	if err != nil {
		log.Info(ParseFansclubMessageError, err.Error())
	}
	content := fansclub.GetContent()
	currentTime := time.ParseEventTime(fansclub.GetCommonInfo().GetCreateTime())
	log.Info("Received fansclub message : 粉丝团消息: %s", content)
	out <- Result{
		method: enums.WebcastFansclubMessage,
		Result: fmt.Sprintf("%s 【粉丝团消息】 {%v}", currentTime, content),
	}
}

func parseRoomUserSeqMessage(payload []byte, out chan<- Result) {
	var roomUserSeq message.RoomUserSeqMessage
	err := proto.Unmarshal(payload, &roomUserSeq)
	if err != nil {
		log.Info(ParseRoomUserSeqMessageError, err.Error())
	}
	current := roomUserSeq.GetTotal()
	total := roomUserSeq.GetTotalPvForAnchor()
	currentTime := time.ParseEventTime(roomUserSeq.GetCommon().GetCreateTime())
	log.Info("Received roomUserSeq message : 当前观看人数: %v , 累计观看人数: %s", current, total)
	out <- Result{
		method: enums.WebcastRoomUserSeqMessage,
		Result: fmt.Sprintf("%s 【统计消息】当前观看人数: {%v} , 累计观看人数: {%s}", currentTime, current, total),
	}
}

func parseSocialMessage(payload []byte, out chan<- Result) {
	var social message.SocialMessage
	err := proto.Unmarshal(payload, &social)
	if err != nil {
		log.Info(ParseSocialMessageError, err.Error())
	}
	userName := social.GetUser().GetNickName()
	userId := social.GetUser().GetId()
	currentTime := time.ParseEventTime(social.GetCommon().GetCreateTime())
	log.Info("Received social message : %s (Id: %v) 关注了主播", userName, userId)
	out <- Result{
		method: enums.WebcastSocialMessage,
		Result: fmt.Sprintf("%s 【关注消息】[ {%v} ] {%v} 关注了主播", currentTime, userName, userId),
	}
}

func parseLikeMessage(payload []byte, out chan<- Result) {
	var like message.LikeMessage
	err := proto.Unmarshal(payload, &like)
	if err != nil {
		log.Info(ParseLikeMessageError, err.Error())
	}
	userName := like.GetUser().GetNickName()
	count := like.GetCount()
	currentTime := time.ParseEventTime(like.GetCommon().GetCreateTime())
	log.Info("Received like message : %s 点了 %v 个赞", userName, count)
	out <- Result{
		method: enums.WebcastLikeMessage,
		Result: fmt.Sprintf("%s 【点赞消息】【{%v}】 点了 {%v} 个赞", currentTime, userName, count),
	}
}
