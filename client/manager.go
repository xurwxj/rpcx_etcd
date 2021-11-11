package client

import (
	"fmt"

	log "github.com/rs/zerolog"

	cmap "github.com/orcaman/concurrent-map"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
)

var ServiceClientPoolMap cmap.ConcurrentMap
var clientConfig *RpcxClientConfig

type RpcxClientConfig struct {
	BasePath   string // /services/dev
	EtcdAddrss []string
	PoolSize   int

	FailMode   client.FailMode
	SelectMode client.SelectMode
	Option     client.Option
	Log        *log.Logger
}

// 初始化微服务客户端参数
func InitClient(rpcxClientConfig *RpcxClientConfig) {
	ServiceClientPoolMap = cmap.New()
	clientConfig = rpcxClientConfig
}

// 通过微服务和接口服务名称，获取一个该微服务服务的连接的客户端
// microservices = "eds"
// service = "xxxxx.product.ed.status"
func GetXclient(microservices, service string) client.XClient {
	service = fmt.Sprintf("%s/%s", microservices, service)
	if ServiceClientPoolMap == nil {
		clientConfig.Log.Error().Msg("please init client ===")
		return nil
	}
	if xc, has := ServiceClientPoolMap.Get(service); has {
		x := xc.(*client.XClientPool).Get()
		if x != nil {
			return x
		}
	}

	d, err := etcd_client.NewEtcdV3Discovery(clientConfig.BasePath, service, clientConfig.EtcdAddrss, true, nil)
	if err != nil {
		clientConfig.Log.Err(err).Msg("GetXclient")
		return nil
	}
	xclientPool := client.NewXClientPool(clientConfig.PoolSize, service, clientConfig.FailMode, clientConfig.SelectMode, d, clientConfig.Option)
	ServiceClientPoolMap.Set(service, xclientPool)
	return xclientPool.Get()
}
