package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ExitWeb
// 优雅退出
func ExitWeb(server *http.Server, exitaction func()) {
	// 创建一个信号通道
	ch := make(chan os.Signal)
	// 监听系统信号
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	// 等待信号
	<-ch

	// 执行退出前动作
	exitaction()
	// 创建一个带有超时的上下文

	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器
	err := server.Shutdown(cxt)
	if err != nil {
		log.Println("服务器关闭出错:", err)
	}
	//
	//// 看看实际退出所耗费的时间
	//fmt.Println("------exited--------", time.Since(now))
}
