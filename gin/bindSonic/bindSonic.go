package bindsonic

import (
	"bytes"
	"io"
	"net/http"
	"reflect"

	"github.com/bytedance/sonic"
)

type SonicJsonBinder struct{}

func (b *SonicJsonBinder) Name() string {
	return "SonicJsonBinder"
}

func (b *SonicJsonBinder) Bind(req *http.Request, obj interface{}) error {
	sonic.Pretouch(reflect.TypeOf(obj))
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	// 重置请求体，以便后续处理
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return sonic.Unmarshal(body, obj)
}
