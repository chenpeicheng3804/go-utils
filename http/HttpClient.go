package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
)

func HttpClientGet(Uri, user, password string) (body []byte, err error) {
	client := &http.Client{}
	//r, _ := http.NewRequest("GET", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r, _ := http.NewRequest("GET", Uri, nil) // URL-encoded payload
	r.SetBasicAuth(user, password)
	resp, err := client.Do(r)
	if err != nil {
		//log.Println(err.Error())
		return body, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
func Api301(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "https://api.m.taobao.com/rest/api3.do?api=mtop.common.getTimestamp")
}
func HttpClientGitlabGet(Uri, Token string) (body []byte, err error) {
	client := &http.Client{}
	//r, _ := http.NewRequest("GET", urlStr, strings.NewReader(data.Encode())) // URL-encoded payload
	r, err := http.NewRequest("GET", Uri, nil) // URL-encoded payload
	if err != nil {
		return body, err
	}
	r.Header.Add("PRIVATE-TOKEN", Token)
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err.Error())
		//return Pages, body, err
		return body, err
	}
	defer resp.Body.Close()
	//获取Header头
	//v, ok := resp.Header["X-Total-Pages"]
	//if ok {
	//	Pages = v[0]
	//}
	//log.Println(io.ReadAll(resp.Body))
	//body, err = io.ReadAll(resp.Body)
	//return Pages, body, err
	return io.ReadAll(resp.Body)
}
func HttpClientGetSOCKS5(Uri, Socks5 string) (body []byte, err error) {
	// 创建一个 SOCKS5 代理拨号器
	dialer, err := proxy.SOCKS5("tcp", Socks5, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("无法连接代理: %v", err)
	}
	// // 设置一个 HTTP 客户端
	// httpClient := &http.Client{
	// 	Transport: &http.Transport{
	// 		Dial: dialer.Dial,
	// 	},
	// }
	// 创建一个带有 DialContext 的 HTTP 客户端
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		},
	}

	// 发起 HTTP GET 请求
	r, err := httpClient.Get(Uri)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer r.Body.Close()

	// 读取响应内容
	return io.ReadAll(r.Body)
}
func HttpClientPost(Uri string, reader *bytes.Reader) (body []byte, err error) {
	client := &http.Client{}
	//data := Tokens_Create{
	//	Name:   "api",
	//	Scopes: []string{"api"},
	//}
	//stu, err := json.Marshal(&data)
	//reader := bytes.NewReader(stu)
	r, err := http.NewRequest("POST", Uri, reader) // URL-encoded payload
	//增加header选项
	r.Header.Set("Content-Type", "application/json")
	//r.SetBasicAuth(user, password)
	if err != nil {
		log.Println("创建NewRequest客户端失败")
		return body, err
	}
	//处理返回结果
	resp, err := client.Do(r)
	if err != nil {
		log.Println("发起Http_Client_Post请求失败")
		return body, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
