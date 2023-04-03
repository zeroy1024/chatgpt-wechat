package chatgpt

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Author struct {
	Role Role `json:"role"`
}

type Content struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

type Message struct {
	Id      string  `json:"id"`
	Author  Author  `json:"author"`
	Content Content `json:"content"`
}

type MessageActionType string

const (
	MessageActionTypeNext    MessageActionType = "next"
	MessageActionTypeVariant MessageActionType = "variant"
)

type RequestBody struct {
	Action          MessageActionType `json:"action"`
	Messages        []Message         `json:"messages"`
	ConversationId  *string           `json:"conversation_id"`
	ParentMessageId string            `json:"parent_message_id"`
	Model           string            `json:"model"`
}
type ResponseBody struct {
	Message        Message `json:"message"`
	ConversationId string  `json:"conversation_id"`
	Error          string  `json:"error"`
}
