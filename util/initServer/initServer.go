package initserver

import (
	"github.com/chenpeicheng3804/go-utils/util/log"
	"net/http"
	"runtime"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

type server interface {
	ListenAndServe() error
}

func InitServer(address string, router *gin.Engine) server {
	log.Debug().Msg(runtime.GOOS)
	switch runtime.GOOS {
	case "windows":
		return &http.Server{
			Addr:           address,
			Handler:        router,
			ReadTimeout:    20 * time.Second,
			WriteTimeout:   20 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	default:
		s := endless.NewServer(address, router)
		s.ReadHeaderTimeout = 20 * time.Second
		s.WriteTimeout = 20 * time.Second
		s.MaxHeaderBytes = 1 << 20
		return s
	}

}
