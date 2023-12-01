package http

import (
	"fmt"
	"testing"
)

func TestHttpClient(t *testing.T) {
	body, _ := HttpClientGetSOCKS5("https://cip.cc", "127.0.0.1:3001")
	fmt.Println(string(body))
	body, _ = HttpClientGet("https://cip.cc", "", "")
	fmt.Println(string(body))

	body, _ = HttpClientGitlabGet("https://gitlab.xxx/api/v4/projects?per_page=50000&search=xxx", "xxx")
	fmt.Println(string(body))
}
