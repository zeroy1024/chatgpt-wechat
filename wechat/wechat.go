package wechat

import (
	"crypto/sha1"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"sort"
	"strings"
	"time"
	"wechatAPI/config"
	"wechatAPI/database"
	"wechatAPI/utils"
)

type Request struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
	MsgId        int64  `xml:"MsgId"`
	MsgDataId    string `xml:"MsgDataId"`
	Idx          string `xml:"Idx"`
}

type xml struct {
	ToUserName   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Content      string `xml:"Content"`
}

func CheckSignature(signature, timestamp, nonce string) bool {
	cfg := config.GetConfig()

	tmpArr := []string{cfg.WechatToken, timestamp, nonce}
	sort.Strings(tmpArr)

	tmpStr := sha1.Sum([]byte(strings.Join(tmpArr, "")))

	return signature == fmt.Sprintf("%x", tmpStr)
}

func HandleRequest(ctx *fiber.Ctx, requestBody *Request) error {
	switch requestBody.MsgType {
	case "text":
		if requestBody.Content == "【收到不支持的消息类型，暂无法显示】" {
			return ctx.SendString("")
		}

		response := xml{
			ToUserName:   requestBody.FromUserName,
			FromUserName: requestBody.ToUserName,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
			Content:      requestBody.HandleTextMessage(),
		}

		return ctx.XML(response)
	default:
		return ctx.SendString("")
	}
}

func (r *Request) HandleTextMessage() string {
	switch {
	case strings.EqualFold(r.Content, "chatgpt"):
		formatString := `由于成本原因,采用了ProxyAPI的方案(使用access_token去请求官方api),这种方案导致了每次请求都需要一定时间，且不能多个人同时使用，所以本公众号设计了队列方案，每次发送消息会回给您一个UUID，您可以通过"查看 uuid“的方式查看chatgpt给您的答复，感谢理解`
		return formatString
	case strings.HasPrefix(r.Content, "查看"):
		messageSplit := strings.Split(r.Content, " ")

		if len(messageSplit) != 2 {
			return "格式错误，正确格式为: 查看 uuid"
		}

		uuidString := messageSplit[1]

		db := database.GetDatabase()
		message := &database.Message{}
		db.Where("uuid = ?", uuidString).First(message)

		if !utils.IsValidUUIDv4(uuidString) {
			return "格式错误，正确格式为: 查看 uuid"
		}

		if message.WeChatMsgID == 0 {
			return "没有找到该条消息"
		}

		if message.Answer == "" {
			return "chatgpt正在抓紧回复，请稍后尝试查看。"
		}

		answer := message.Answer

		db.Delete(message)

		return answer
	default:
		u := r.handleChatGPTMessage()
		return fmt.Sprintf("UUID: %s\n格式: 查看 UUID\n例如: 查看 %s", u, u)
	}

}

func (r *Request) handleChatGPTMessage() string {
	db := database.GetDatabase()

	user := &database.User{}
	db.Where("open_id = ?", r.FromUserName).Preload("Session").First(user)

	if user.ID == 0 {
		session := &database.Session{}
		session.ParentMessageID = uuid.New().String()

		user.OpenID = r.FromUserName
		user.Model = "text-davinci-002-render-sha"
		user.AccessToken = ""
		user.Session = *session

		db.Create(user)
	}

	if user.Session.ID == 0 {
		session := &database.Session{}
		session.ParentMessageID = uuid.New().String()
		session.UserID = user.ID
		db.Create(session)

		user.Session = *session
	}

	message := &database.Message{}
	db.Where("we_chat_msg_id = ?", r.MsgId).First(message)
	if message.UUID == "" {
		message.UUID = uuid.New().String()
		message.WeChatMsgID = r.MsgId
		message.SessionID = user.Session.ID
		message.CreateAt = time.Now()
		message.Question = r.Content
		db.Create(message)
	}

	return message.UUID
}
