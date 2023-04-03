package chatgpt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

type UnofficialProxyAPI struct {
	AccessToken     string `json:"access_token"`
	Model           string `json:"model"`
	ReverseProxyAPI string `json:"reverse_proxy_api"`
}

type SendMessageBrowserOptions struct {
	ConversationID  *string
	ParentMessageID string
	MessageID       string
	Action          MessageActionType
}

// 生成body
func (api *UnofficialProxyAPI) generateRequestBody(message string, options SendMessageBrowserOptions) RequestBody {
	requestBody := RequestBody{
		Model:           api.Model,
		ParentMessageId: options.ParentMessageID,
		Action:          options.Action,
		Messages: []Message{
			{
				Id: uuid.New().String(),
				Author: Author{
					Role: RoleUser,
				},
				Content: Content{
					ContentType: "text",
					Parts:       []string{message},
				},
			},
		},
	}

	if options.ConversationID != nil {
		requestBody.ConversationId = options.ConversationID
	}

	if requestBody.ParentMessageId == "" {
		requestBody.ParentMessageId = uuid.New().String()
	}

	if requestBody.Action == "" {
		requestBody.Action = MessageActionTypeNext
	}

	return requestBody
}

// 发送请求
func (api *UnofficialProxyAPI) sendRequest(body RequestBody) (*http.Response, error) {
	requestBodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", api.ReverseProxyAPI, bytes.NewBuffer(requestBodyBytes))

	// set header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.AccessToken))
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}

	return client.Do(req)
}

// 处理响应
func (api *UnofficialProxyAPI) handleResponse(response *http.Response) ([]ResponseBody, error) {
	if !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("handleResponse body read error: %s", err)
		}

		return nil, fmt.Errorf("handleResponse Content-Type error: %s", string(bodyBytes))
	}

	scanner := bufio.NewScanner(response.Body)

	var responseList []ResponseBody
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 6 && line[:5] == "data:" {
			data := line[6:]

			if data == "[DONE]" {
				break
			}

			if !json.Valid([]byte(data)) && len(responseList) > 0 {
				continue
			}

			var responseBody ResponseBody
			err := json.Unmarshal([]byte(data), &responseBody)
			if err != nil {
				return nil, fmt.Errorf("handleResponse json.Unmarshal error: %s", err)
			}

			responseList = append(responseList, responseBody)
		}
	}

	return responseList, nil
}

func (api *UnofficialProxyAPI) SendMessage(message string, options SendMessageBrowserOptions) ([]ResponseBody, error) {
	requestBody := api.generateRequestBody(message, options)

	response, err := api.sendRequest(requestBody)
	if err != nil {
		return nil, fmt.Errorf("sendRequest error: %s", err)
	}

	return api.handleResponse(response)
}
