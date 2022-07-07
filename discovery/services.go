package discovery

import (
	"fmt"
	"strings"

	json "github.com/json-iterator/go"

	"github.com/rpcxio/libkv/store"
	"github.com/smallnest/rpcx/client"
	etcd_client "github.com/xurwxj/rpcx_etcd/client"
)

type ServiceWactchParam struct {
	BasePath   string
	Mod        string
	EtcdAddrss []string
	Options    *store.Config
	CallBack   func(map[string]ServiceData)
}
type ServiceData struct {
	MethodType   string   `json:"methodType,omitempty" `
	Name         string   `json:"name,omitempty" `
	FuncName     string   `json:"funcName,omitempty" `
	Path         string   `json:"path,omitempty" `
	Method       string   `json:"method,omitempty" `
	Auth         string   `json:"auth,omitempty" `
	Perms        []string `json:"perms,omitempty"`
	ProductLines []string `json:"productLines,omitempty"`
	Funcs        []string `json:"funcs,omitempty" `
	Hosts        []string `json:"hosts,omitempty" `
	AppName      string   `json:"appName"`
}

// 启动一个etcd 监听，服务发生变动都会触发通知 go StartWatchServices()
// basePath ="/services"
// mod 是环境 dev
// etcdAddrss etcd地址列表
func StartWatchServices(param *ServiceWactchParam) {
	param.Mod = fmt.Sprintf("%s/", param.Mod)
	d, _ := etcd_client.NewEtcdV3Discovery(param.BasePath, param.Mod, param.EtcdAddrss, true, param.Options)
	watchCh := d.WatchService()
	for serviceData := range watchCh {
		collectServiceData(serviceData, param.CallBack)
	}
}

// 将通知的数据做处理 过滤一些没带地址的数据
// xxxxx.product.ed.status/tcp@localhost:8972 这些是需要的数据key
func collectServiceData(wresp []*client.KVPair, callBack func(map[string]ServiceData)) {
	tmpData := make(map[string]ServiceData, len(wresp))
	appData := make(map[string]ServiceData)
	for _, v := range wresp {
		keys := strings.Split(v.Key, "/")
		if len(keys) != 2 {
			continue
		}
		infos := strings.Split(v.Value, "=")
		if len(infos) != 2 {
			continue
		}
		serviceName := keys[0]
		address := keys[1]
		strs := strings.Split(address, "@")
		if len(strs) != 2 {
			continue
		}
		serviceInfo, ok := tmpData[serviceName]
		if !ok {
			meta := ServiceData{}
			if err := json.Unmarshal([]byte(infos[1]), &meta); err != nil {
				continue
			}
			meta.MethodType = infos[0]
			meta.Hosts = make([]string, 0)
			meta.Hosts = append(meta.Hosts, strs[1])
			tmpData[serviceName] = meta
			continue
		}
		serviceInfo.Hosts = append(serviceInfo.Hosts, strs[1])
		tmpData[serviceName] = serviceInfo
	}

	for k, v := range tmpData {
		if k == v.AppName {
			continue
		}
		if _, ok := appData[k]; !ok {
			appData[v.AppName] = v
		}
	}

	for k, v := range appData {
		tmpData[k] = v
	}
	if len(tmpData) > 0 && callBack != nil {
		callBack(tmpData)
	}
}
