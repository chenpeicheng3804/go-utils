package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// InitServer
// 初始化web server
// InitServer 函数用于初始化服务器，接受一个 gin.Engine 类型的参数 e，一个字符串类型的参数 Addr，以及两个函数类型的参数 initaction 和 exitaction
func InitServer(e *gin.Engine, Addr string, initaction func(), exitaction func()) {

	//配置优雅退出
	server := &http.Server{

		Addr:           Addr,             // 服务器监听的地址
		Handler:        e,                // 服务器使用的处理器
		ReadTimeout:    10 * time.Second, // 读取请求的超时时间
		WriteTimeout:   10 * time.Second, // 写入响应的超时时间
		MaxHeaderBytes: 1 << 20,          // 请求头的最大字节数
	}

	go server.ListenAndServe() // 在 goroutine 中启动服务器
	initaction()               // 执行初始化操作

	// 设置优雅退出
	ExitWeb(server, exitaction) // 调用 ExitWeb 函数设置优雅退出
}
