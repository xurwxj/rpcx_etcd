package discovery

import (
	"fmt"
	"strings"

	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
)

type ServiceFunc struct {
	Address string
	SFMeta  string
}

var MicroservicesData map[string]map[string][]ServiceFunc

// 启动一个etcd 监听，服务发生变动都会触发通知 go StartWatchServices()
// basePath ="/services"
// mod 是环境 dev
// etcdAddrss etcd地址列表
func StartWatchServices(basePath, mod string, etcdAddrss []string) {
	mod = fmt.Sprintf("%s/", mod)
	d, _ := etcd_client.NewEtcdV3Discovery(basePath, mod, etcdAddrss, true, nil)
	watchCh := d.WatchService()
	for serviceData := range watchCh {
		collectServiceData(serviceData)
	}
}

// 将通知的数据做处理 过滤一些没带地址的数据
// eds/xxxxx.product.ed.status/tcp@localhost:8972 这些是需要的数据key
func collectServiceData(wresp []*client.KVPair) {
	MicroservicesData = make(map[string]map[string][]ServiceFunc)
	for _, v := range wresp {
		keys := strings.Split(v.Key, "/")
		if len(keys) != 3 {
			continue
		}
		microservicesName, serviceName, address := keys[0], keys[1], keys[2]
		newServiceFunc := ServiceFunc{
			Address: address,
			SFMeta:  v.Value,
		}
		serviceMap, ok := MicroservicesData[microservicesName]
		if !ok {
			serviceMap = make(map[string][]ServiceFunc)
			serviceMap[serviceName] = []ServiceFunc{newServiceFunc}
			MicroservicesData[microservicesName] = serviceMap
			continue
		}

		if serviceFuncs, ok := serviceMap[serviceName]; ok {
			serviceFuncs = append(serviceFuncs, newServiceFunc)
			serviceMap[serviceName] = serviceFuncs
			MicroservicesData[microservicesName] = serviceMap
			continue
		}
		serviceMap[serviceName] = []ServiceFunc{newServiceFunc}
		MicroservicesData[microservicesName] = serviceMap
	}
}
