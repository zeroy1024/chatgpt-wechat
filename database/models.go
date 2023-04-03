package database

import (
	"time"
)

type User struct {
	ID          uint `gorm:"primarykey"`
	OpenID      string
	AccessToken string
	Model       string
	Session     Session `gorm:"foreignKey:UserID"`
}

type Session struct {
	ID              uint `gorm:"primarykey"`
	UserID          uint
	ConversationID  string
	ParentMessageID string
	Message         []Message `gorm:"foreignKey:SessionID"`
}

type Message struct {
	UUID        string `gorm:"primarykey"`
	WeChatMsgID int64
	SessionID   uint
	CreateAt    time.Time
	Question    string
	Answer      string
}
