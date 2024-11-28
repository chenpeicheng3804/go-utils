package bind

import (
	"bytes"
	"github.com/goccy/go-json"
	"io"
	"net/http"
)

type JSONBinder struct{}

func (b *JSONBinder) Name() string {
	return "JSONBinder"
}

func (b *JSONBinder) Bind(req *http.Request, obj interface{}) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	// 重置请求体，以便后续处理
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return json.Unmarshal(body, obj)
}
