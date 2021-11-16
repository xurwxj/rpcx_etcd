package client

import (
	"context"

	"github.com/rpcxio/libkv/store"
	log "github.com/rs/zerolog"

	cmap "github.com/orcaman/concurrent-map"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
)

var serviceClientPoolMap cmap.ConcurrentMap
var clientConfig *RpcxClientConfig

type RpcxClientConfig struct {
	BasePath   string   // /services/dev
	EtcdAddrss []string // etcd地址
	PoolSize   int      // 连接池

	FailMode   client.FailMode
	SelectMode client.SelectMode
	Option     client.Option
	Log        *log.Logger
	Options    *store.Config
}

// 初始化微服务客户端参数
func InitClient(rpcxClientConfig *RpcxClientConfig) {
	serviceClientPoolMap = cmap.New()
	clientConfig = rpcxClientConfig
}

// 向微服务发送rpcx请求
func CallService(service, serviceMethod string, args, reply interface{}) bool {
	ctx := context.Background()
	xclient := getXclient(service)
	if xclient == nil {
		clientConfig.Log.Error().Msg("get service client nil ")
		return false
	}
	err := xclient.Call(ctx, serviceMethod, args, reply)
	if err != nil {
		clientConfig.Log.Err(err).Msg("get service client ")
		return false
	}
	return true
}

// 通过接口服务名称，获取一个该接口服务的连接的客户端
// service = "xxxxx.product.ed.status"
func getXclient(service string) client.XClient {
	if serviceClientPoolMap == nil {
		clientConfig.Log.Error().Msg("please init client ===")
		return nil
	}
	if xc, has := serviceClientPoolMap.Get(service); has {
		x := xc.(*client.XClientPool).Get()
		if x != nil {
			return x
		}
	}

	d, err := etcd_client.NewEtcdV3Discovery(clientConfig.BasePath, service, clientConfig.EtcdAddrss, true, clientConfig.Options)
	if err != nil {
		clientConfig.Log.Err(err).Msg("GetXclient")
		return nil
	}
	xclientPool := client.NewXClientPool(clientConfig.PoolSize, service, clientConfig.FailMode, clientConfig.SelectMode, d, clientConfig.Option)
	serviceClientPoolMap.Set(service, xclientPool)
	return xclientPool.Get()
}
