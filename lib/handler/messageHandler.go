package handler

import (
	"bytes"
	"compress/gzip"
	"douyinLiveCollectors/lib/message"
	"fmt"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"log"
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
const (
	WebcastChatMessage        = "WebcastChatMessage"
	WebcastGiftMessage        = "WebcastGiftMessage"
	WebcastLikeMessage        = "WebcastLikeMessage"
	WebcastMemberMessage      = "WebcastMemberMessage"
	WebcastSocialMessage      = "WebcastSocialMessage"
	WebcastRoomUserSeqMessage = "WebcastRoomUserSeqMessage"
	WebcastFansclubMessage    = "WebcastFansclubMessage"
	WebcastControlMessage     = "WebcastControlMessage"
	WebcastEmojiChatMessage   = "WebcastEmojiChatMessage"
	WebcastRoomStatsMessage   = "WebcastRoomStatsMessage"
	WebcastRoomMessage        = "WebcastRoomMessage"
	WebcastRoomRankMessage    = "WebcastRoomRankMessage"
)

func Handler(ws *websocket.Conn, payload []byte) {
	resp := parseAndAck(ws, payload)
	for _, msg := range resp.GetMessagesList() {
		switch msg.GetMethod() {
		case WebcastChatMessage:
			parseChatMessage(msg.GetPayload())
		case WebcastGiftMessage:
			parseGiftMessage(msg.GetPayload())
		case WebcastMemberMessage:
			parseMemberMessage(msg.GetPayload())
		case WebcastLikeMessage:
			parseLikeMessage(msg.GetPayload())
		case WebcastSocialMessage:
			parseSocialMessage(msg.GetPayload())
		case WebcastRoomUserSeqMessage:
			parseRoomUserSeqMessage(msg.GetPayload())
		case WebcastFansclubMessage:
			parseFansclubMessage(msg.GetPayload())
		case WebcastControlMessage:
			parseControlMessage(msg.GetPayload())
			ws.Close()
		case WebcastEmojiChatMessage:
			parseEmojiChatMessage(msg.GetPayload())
		case WebcastRoomStatsMessage:
			parseRoomStatsMessage(msg.GetPayload())
		case WebcastRoomMessage:
			parseRoomMessage(msg.GetPayload())
		case WebcastRoomRankMessage:
			parseRoomRankMessage(msg.GetPayload())
		default:
			log.Printf(UnknownMessageError, msg.String())
		}
	}
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

func parseAndAck(ws *websocket.Conn, payload []byte) *message.Response {
	pkg, err := parsePushFrame(payload)
	if err != nil {
		log.Fatalf(ParsePushFrameError, err)
	}

	decompressedData, err := decompressGzip(pkg.Payload)
	if err != nil {
		log.Fatalf(DecompressPayloadError, err)
	}

	resp, err := parseResponse(decompressedData)
	if err != nil {
		log.Fatalf(ParseResponseError, err)
	}

	if resp.NeedAck {
		err = sendAck(ws, pkg.LogId, resp.InternalExt)
		if err != nil {
			log.Fatalf(SendAckError, err)
		}
		log.Println("ACK sent successfully.")
	}
	return resp
}

func parseChatMessage(payload []byte) {
	var chat message.ChatMessage
	err := proto.Unmarshal(payload, &chat)
	if err != nil {
		log.Printf(ParseChatMessageError, err)
	}
	userName := chat.GetUser().GetNickName()
	userId := chat.GetUser().GetId()
	content := chat.GetContent()
	log.Printf("【聊天消息】[ {%v} ] {%v} : {%v}", userId, userName, content)
}

func parseGiftMessage(payload []byte) {
	var gift message.GiftMessage
	err := proto.Unmarshal(payload, &gift)
	if err != nil {
		log.Printf(ParseGiftMessageError, err)
	}
	userName := gift.GetUser().GetNickName()
	giftName := gift.GetGift().GetName()
	combo := gift.GetComboCount()
	log.Printf("【礼物消息】{%v} 送出了 {%v} X {%v}连击", userName, giftName, combo)
}

func parseMemberMessage(payload []byte) {
	var member message.MemberMessage
	err := proto.Unmarshal(payload, &member)
	if err != nil {
		log.Printf(ParseMemberMessageError, err)
	}
	userId := member.GetUser().GetId()
	userName := member.GetUser().GetNickName()
	gender := []string{"女", "男", "unknown"}[member.GetUser().GetGender()]
	log.Printf("【进场消息】[ {%v} ][ {%v} ] {%v} 进入了直播间", userId, gender, userName)
}

func parseRoomRankMessage(payload []byte) {
	var roomRank message.RoomRankMessage
	err := proto.Unmarshal(payload, &roomRank)
	if err != nil {
		log.Printf(ParseRoomRankMessageError, err)
	}
	ranksList := roomRank.GetRanksList()
	ranks := make(map[int]interface{}, 3)
	for i, rank := range ranksList {
		ranks[i] = map[string]interface{}{
			"Id": rank.GetUser().GetId(),
			"昵称": rank.GetUser().GetNickName(),
		}
	}
	log.Printf("【直播间排行榜消息】{%v}", ranks)
}

func parseRoomMessage(payload []byte) {
	var room message.RoomMessage
	err := proto.Unmarshal(payload, &room)
	if err != nil {
		log.Printf(ParseRoomMessageError, err)
	}
	roomId := room.GetCommon().GetRoomId()
	log.Printf("【直播间消息】直播间id: {%v}", roomId)
}

func parseRoomStatsMessage(payload []byte) {
	var roomStats message.RoomStatsMessage
	err := proto.Unmarshal(payload, &roomStats)
	if err != nil {
		log.Printf(ParseRoomStatsMessageError, err)
	}
	displayLong := roomStats.GetDisplayLong()
	log.Printf("【直播间统计消息】{%v}", displayLong)
}

func parseEmojiChatMessage(payload []byte) {
	var emoji message.EmojiChatMessage
	err := proto.Unmarshal(payload, &emoji)
	if err != nil {
		log.Printf(ParseEmojiChatMessageError, err)
	}
	emojiId := emoji.GetEmojiId()
	user := emoji.GetUser()
	common := emoji.GetCommon()
	defaultContent := emoji.GetDefaultContent()
	log.Printf("【聊天表情包ID】 {%v},user：{%v},common:{%v},defaultContent:{%v}", emojiId, user, common, defaultContent)
}

func parseControlMessage(payload []byte) {
	var control message.ControlMessage
	err := proto.Unmarshal(payload, &control)
	if err != nil {
		log.Printf(ParseControlMessageError, err)
	}
	if control.GetStatus() == 3 {
		roomId := control.GetCommon().GetRoomId()
		log.Printf("【直播间消息】直播间 {%v} 已结束", roomId)
	}
}

func parseFansclubMessage(payload []byte) {
	var fansclub message.FansclubMessage
	err := proto.Unmarshal(payload, &fansclub)
	if err != nil {
		log.Printf(ParseFansclubMessageError, err)
	}
	content := fansclub.GetContent()
	log.Printf("【粉丝团消息】 {%v}", content)
}

func parseRoomUserSeqMessage(payload []byte) {
	var roomUserSeq message.RoomUserSeqMessage
	err := proto.Unmarshal(payload, &roomUserSeq)
	if err != nil {
		log.Printf(ParseRoomUserSeqMessageError, err)
	}
	current := roomUserSeq.GetTotal()
	total := roomUserSeq.GetTotalPvForAnchor()
	log.Printf("【统计消息】当前观看人数: {%v} , 累计观看人数: {%v}", current, total)
}

func parseSocialMessage(payload []byte) {
	var social message.SocialMessage
	err := proto.Unmarshal(payload, &social)
	if err != nil {
		log.Printf(ParseSocialMessageError, err)
	}
	userName := social.GetUser().GetNickName()
	userId := social.GetUser().GetId()
	log.Printf("【关注消息】[ {%v} ] {%v} 关注了主播", userName, userId)
}

func parseLikeMessage(payload []byte) {
	var like message.LikeMessage
	err := proto.Unmarshal(payload, &like)
	if err != nil {
		log.Printf(ParseLikeMessageError, err)
	}
	userName := like.GetUser().GetNickName()
	count := like.GetCount()
	log.Printf("【点赞消息】{%v} 点了 {%v} 个赞", userName, count)

}
