package nacos

import (
	"fmt"
	"testing"
)

var ServiceConfig = NewService(Service{
	NacosIp:   "192.168.3.30",
	NacosPort: 8848,     // 默认值8848
	NacosPath: "/nacos", // 默认值/nacos
	//NacosNamespaceId:   "demo",    	// 默认值public
	//ServiceClusterName: "cluster-a", 	// 默认值DEFAULT
	//ServiceGroupName:   "group-a", 	// 默认值DEFAULT_GROUP
	ServicePort:     8080,
	ServiceName:     "test",
	ServiceIp:       "192.168.3.30",
	ServiceMetadata: map[string]string{"idc": "shanghai"},
})

func TestRegisterServiceInstance(t *testing.T) {

	//创建nacos客户端
	ServiceConfig.NewCreateNacosClient()
	//注册客户端
	ServiceConfig.RegisterServiceInstance()
	//查询服务信息
	services, _ := ServiceConfig.GetService()
	fmt.Println("服务信息: ", services)

	//	获取注册实例信息
	for {
		instances, _ := ServiceConfig.SelectAllInstances()
		//for _,instance := range instances{
		//
		//}
		if len(instances) != 0 {
			for _, instance := range instances {
				fmt.Println(instance)
			}
			break
		}
	}
	//	获取健康实例列表
	for {
		instances, _ := ServiceConfig.SelectInstances()
		//for _,instance := range instances{
		//
		//}
		if len(instances) != 0 {
			fmt.Println("注册健康实例列表信息: ", instances)
			break
		}
	}

	// 获取加权随机轮询一个健康的实例
	for {
		instance, _ := ServiceConfig.SelectOneHealthyInstance()
		//for _,instance := range instances{
		//
		//}
		if instance != nil {
			fmt.Println("加权随机轮询一个健康的实例信息: ", instance)
			break
		}
	}
	//	注销实例
	ServiceConfig.DeRegisterServiceInstance()

}
