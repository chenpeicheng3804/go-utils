package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ExitWeb
// 优雅退出
func ExitWeb(server *http.Server, exitaction func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch

	//退出前动作
	exitaction()

	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.Shutdown(cxt)
	if err != nil {
		fmt.Println("err", err)
	}
	//
	//// 看看实际退出所耗费的时间
	//fmt.Println("------exited--------", time.Since(now))
}
