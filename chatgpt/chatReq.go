package chatgpt

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
)

type ChatReq struct {
	Action          string           `json:"action"`
	Messages        []ChatReqMessage `json:"messages"`
	ConversationId  interface{}      `json:"conversation_id"`
	ParentMessageId string           `json:"parent_message_id"`
	Model           string           `json:"model"`
}

type ChatReqMessage struct {
	Id      string            `json:"id"`
	Role    string            `json:"role"`
	Content ChatReqMsgContent `json:"content"`
}

type ChatReqMsgContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

func (msg *ChatReq) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

func CreateChatReqBody(message, parentID string, conversationId interface{}) *bytes.Buffer {
	req := &ChatReq{
		Action: "next",
		Messages: []ChatReqMessage{
			{
				Id:   uuid.New().String(),
				Role: "user",
				Content: ChatReqMsgContent{
					ContentType: "text",
					Parts:       []string{message},
				},
			},
		},
		ConversationId:  conversationId,
		ParentMessageId: parentID,
		Model:           "text-davinci-002-render",
	}
	return bytes.NewBuffer(req.ToJson())
}
