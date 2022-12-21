package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func SendWeChatMessage(url, content string) error {
	var contentMap map[string]interface{}
	contentMap = make(map[string]interface{})
	contentMap["content"] = content
	var body map[string]interface{}
	body = make(map[string]interface{})
	body["msgtype"] = "markdown"
	body["markdown"] = contentMap
	bodyByte, err := json.Marshal(body)
	if err != nil {
		return err
	}
	response, err := client.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		return nil
	}
	return errors.New(response.Status)
}
