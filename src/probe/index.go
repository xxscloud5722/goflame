package probe

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var client = http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func Http(requestUrl, method string, headers map[string]string, body string, form map[string]string) error {
	var request *http.Request
	var err error
	if form != nil && len(form) > 0 {
		formData := url.Values{}
		if form != nil {
			for key, value := range form {
				formData.Set(key, value)
			}
		}
		request, err = http.NewRequest(method, requestUrl, strings.NewReader(formData.Encode()))
	} else if body != "" {
		request, err = http.NewRequest(method, requestUrl, bytes.NewBuffer([]byte(body)))
	} else {
		request, err = http.NewRequest(method, requestUrl, nil)
	}
	if headers != nil {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		return nil
	}
	return errors.New("StatusCode Not 200")
}
