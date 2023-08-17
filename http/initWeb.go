package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// InitServer
// 初始化web server
func InitServer(e *gin.Engine, Addr string, initaction func(), exitaction func()) {

	//配置优雅退出
	server := &http.Server{

		Addr:           Addr,
		Handler:        e,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go server.ListenAndServe()
	initaction()

	// 设置优雅退出
	ExitWeb(server, exitaction)
}
