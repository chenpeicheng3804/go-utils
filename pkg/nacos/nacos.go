package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Service
// nacos 参数结构体
type Service struct {
	NacosPort          uint64            `param:"port"`
	ServicePort        uint64            `param:"port"`
	ServiceMetadata    map[string]string `param:"metadata"`
	ServiceIp          string            `param:"ip"`
	ServiceClusterName string            `param:"clusterName"`
	ServiceName        string            `param:"serviceName"`
	ServiceGroupName   string            `param:"groupName"`
	NacosIp            string            `param:"ip"`
	NacosPath          string            `param:"clusterName"`
	NacosNamespaceId   string            `param:"NamespaceId"`
	ServiceReplicas    string            `param:"Replicas"`
	Client             naming_client.INamingClient
}

func NewService(Service Service) *Service {
	if len(Service.ServiceName) == 0 {
		panic("ServiceName Is empty!")
	}
	if len(Service.NacosIp) == 0 {
		panic("NacosIp Is empty!")
	}
	if Service.NacosPort == 0 {
		Service.NacosPort = 8848
	}
	if len(Service.NacosPath) == 0 {
		Service.NacosPath = "/nacos"
	}
	return &Service
}

// NewCreateNacosClient
// 创建nacos连接客户端
func (s *Service) NewCreateNacosClient() {

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(s.NacosIp, s.NacosPort, constant.WithContextPath(s.NacosPath)),
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(s.NacosNamespaceId),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		//constant.WithLogDir("/tmp/nacosson/log"),
		//constant.WithCacheDir("/tmp/nacosson/cache"),
		constant.WithLogLevel("error"),
	)

	// create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		panic(err)
	}
	//赋值nacos 客户端
	s.Client = client
}

// RegisterServiceInstance
// 注册nacos服务实例
func (s *Service) RegisterServiceInstance() {
	//提交注册信息
	success, err := s.Client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          s.ServiceIp,
		Port:        s.ServicePort,
		ServiceName: s.ServiceName,
		GroupName:   s.ServiceGroupName,
		ClusterName: s.ServiceClusterName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    s.ServiceMetadata,
	})
	if !success || err != nil {
		panic("RegisterServiceInstance failed!" + err.Error())
	}
	fmt.Printf("RegisterServiceInstance,param:%+v,result:%+v \n\n", s, success)

}

// DeRegisterServiceInstance
// 注销nacos服务实例
func (s *Service) DeRegisterServiceInstance() {
	//提交注销信息
	success, err := s.Client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          s.ServiceIp,
		Port:        s.ServicePort,
		ServiceName: s.ServiceName,
		GroupName:   s.ServiceGroupName,
		Cluster:     s.ServiceClusterName,
		Ephemeral:   true, //it must be true
	})

	if !success || err != nil {
		panic("DeRegisterServiceInstance failed!" + err.Error())
	}
	fmt.Printf("DeRegisterServiceInstance,param:%+v,result:%+v \n\n", s, success)
}

// GetService
// 获取服务信息
func (s *Service) GetService(service ...Service) (model.Service, error) {

	if service == nil {
		//fmt.Println(service)
		service = make([]Service, 1)
		service[0].ServiceName = s.ServiceName
		service[0].ServiceClusterName = s.ServiceClusterName
		service[0].ServiceGroupName = s.ServiceGroupName
	}
	if len(service[0].ServiceName) == 0 {
		panic("ServiceName Is empty!")
	}
	services, err := s.Client.GetService(vo.GetServiceParam{
		ServiceName: service[0].ServiceName,
		Clusters:    []string{service[0].ServiceClusterName}, // 默认值DEFAULT
		GroupName:   service[0].ServiceGroupName,             // 默认值DEFAULT_GROUP
	})
	return services, err
}

// SelectAllInstances
// 获取所有的实例列表
func (s *Service) SelectAllInstances(service ...Service) (instances []model.Instance, err error) {

	if service == nil {
		service = make([]Service, 1)
		service[0].ServiceName = s.ServiceName
		service[0].ServiceClusterName = s.ServiceClusterName
		service[0].ServiceGroupName = s.ServiceGroupName
	}
	if len(service[0].ServiceName) == 0 {
		panic("ServiceName Is empty!")
	}
	instances, err = s.Client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: service[0].ServiceName,
		GroupName:   service[0].ServiceGroupName,             // 默认值DEFAULT_GROUP
		Clusters:    []string{service[0].ServiceClusterName}, // 默认值DEFAULT
	})
	return instances, err
}

// SelectInstances
// 获取健康实例列表
func (s *Service) SelectInstances(service ...Service) (instances []model.Instance, err error) {

	if service == nil {
		service = make([]Service, 1)
		service[0].ServiceName = s.ServiceName
		service[0].ServiceClusterName = s.ServiceClusterName
		service[0].ServiceGroupName = s.ServiceGroupName
	}
	if len(service[0].ServiceName) == 0 {
		panic("ServiceName Is empty!")
	}

	// SelectInstances 只返回满足这些条件的实例列表：healthy=${HealthyOnly},enable=true 和weight>0
	instances, err = s.Client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: service[0].ServiceName,
		GroupName:   service[0].ServiceGroupName,             // 默认值DEFAULT_GROUP
		Clusters:    []string{service[0].ServiceClusterName}, // 默认值DEFAULT
		HealthyOnly: true,
	})
	return instances, err
}

// SelectOneHealthyInstance 将会按加权随机轮询的负载均衡策略返回一个健康的实例
// 实例必须满足的条件：health=true,enable=true and weight>0
func (s *Service) SelectOneHealthyInstance(service ...Service) (instance *model.Instance, err error) {

	if service == nil {
		service = make([]Service, 1)
		service[0].ServiceName = s.ServiceName
		service[0].ServiceClusterName = s.ServiceClusterName
		service[0].ServiceGroupName = s.ServiceGroupName
	}
	if len(service[0].ServiceName) == 0 {
		panic("ServiceName Is empty!")
	}

	// SelectOneHealthyInstance将会按加权随机轮询的负载均衡策略返回一个健康的实例
	// 实例必须满足的条件：health=true,enable=true and weight>0
	instance, err = s.Client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: service[0].ServiceName,
		GroupName:   service[0].ServiceGroupName,             // 默认值DEFAULT_GROUP
		Clusters:    []string{service[0].ServiceClusterName}, // 默认值DEFAULT
	})
	return instance, err
}
