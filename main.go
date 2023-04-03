package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
	"log"
	"wechatAPI/chatgpt"
	"wechatAPI/config"
	"wechatAPI/database"
	"wechatAPI/wechat"
)

var cfg *config.Config
var db *gorm.DB

func HandleMessageQueuing() {
	chatgptAPI := chatgpt.UnofficialProxyAPI{
		AccessToken:     cfg.UnofficialProxyAPI.AccessToken,
		Model:           cfg.UnofficialProxyAPI.Model,
		ReverseProxyAPI: cfg.UnofficialProxyAPI.ReverseProxyAPI,
	}

	var messageList []database.Message
	db.Where("answer = ?", "").Order("create_at ASC").Find(&messageList)

	for _, message := range messageList {
		session := &database.Session{}
		db.Where("id = ?", message.SessionID).First(session)

		if session.ParentMessageID != "" {
			options := chatgpt.SendMessageBrowserOptions{
				ParentMessageID: session.ParentMessageID,
			}

			if session.ConversationID != "" {
				options.ConversationID = &session.ConversationID
			}

			answer, err := chatgptAPI.SendMessage(message.Question, options)
			if err != nil {
				message.Answer = "出现错误: " + err.Error()
				db.Save(message)
				continue
			}

			message.Answer = answer[len(answer)-1].Message.Content.Parts[0]
			session.ParentMessageID = answer[len(answer)-1].Message.Id
			session.ConversationID = answer[len(answer)-1].ConversationId

			db.Save(message)
			db.Save(session)
		}
	}
}

func init() {
	// init config
	err := config.InitConfig()
	if err != nil {
		log.Fatalln(err)
	}
	cfg = config.GetConfig()

	// init redis
	err = database.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatalln(err)
	}
	db = database.GetDatabase()
}

func main() {
	go func() {
		for {
			HandleMessageQueuing()
		}
	}()

	app := fiber.New()

	// logger
	app.Use(logger.New(logger.Config{}))

	// auth signature
	app.Use(func(c *fiber.Ctx) error {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")

		if !wechat.CheckSignature(signature, timestamp, nonce) {
			return c.SendStatus(401)
		}

		return c.Next()
	})

	// routes
	app.Get("/", func(c *fiber.Ctx) error {
		echoStr := c.Query("echostr")
		return c.SendString(echoStr)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		//openid := c.Query("openid")

		requestBody := new(wechat.Request)
		err := c.BodyParser(requestBody)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}

		return wechat.HandleRequest(c, requestBody)
	})

	_ = app.Listen(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
}
