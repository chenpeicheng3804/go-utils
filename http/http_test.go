package http

import (
	"github.com/chenpeicheng3804/go-utils/nacos"
	"github.com/chenpeicheng3804/go-utils/util"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"testing"
)

var ServiceConfig = nacos.NewService(nacos.Service{
	NacosIp:            "192.168.3.30",
	NacosPort:          8848,
	NacosPath:          "/nacos",
	NacosNamespaceId:   "demo",
	ServiceClusterName: "cluster-a",
	ServiceGroupName:   "group-a",
	ServicePort:        8080,
	ServiceName:        "test",
	ServiceIp:          util.GetIps(),
	ServiceMetadata:    map[string]string{"idc": "shanghai"},
})

// Elegant online and offline
func TestElegantOnlineAndOffline(t *testing.T) {
	//创建nacos客户端
	ServiceConfig.NewCreateNacosClient()

	//创建gin服务
	r := gin.Default()

	//创建路由组
	ServerGroup := r.Group(ServiceConfig.ServiceName)
	//配置路由
	{
		ServerGroup.GET("/", Default)
	}
	log.Println("Listen: ", "http://"+ServiceConfig.ServiceIp+":"+strconv.FormatUint(ServiceConfig.ServicePort, 10)+"/"+ServiceConfig.ServiceName+"/")
	//将nacos 注册 注销函数传入
	InitServer(r, ServiceConfig.ServiceIp+":"+strconv.FormatUint(ServiceConfig.ServicePort, 10), ServiceConfig.RegisterServiceInstance, ServiceConfig.DeRegisterServiceInstance)
}

// Default
// web前端页面
func Default(c *gin.Context) {
	c.Writer.Write([]byte(`<html>
<head><title>demo</title></head>
<body>
<h1>` + ServiceConfig.ServiceIp + ":" + strconv.FormatUint(ServiceConfig.ServicePort, 10) + `</h1>
</body>
</html>
`))

}
