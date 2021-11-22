package server

import (
	log "github.com/rs/zerolog"

	"github.com/smallnest/rpcx/server"
	"github.com/xurwxj/rpcx_etcd/registry"
)

type MicroServer struct {
	RpcxServer *server.Server
	Log        *log.Logger

	ServiceAddress string
}

type ServerPlugin interface {
	Start() error
}

// 添加server插件
func (ms *MicroServer) AddServerPlugin(serverPlugin ServerPlugin) {
	err := serverPlugin.Start()
	if err != nil {
		ms.Log.Err(err).Msg("addRegistryPlugin")
	}
	ms.RpcxServer.Plugins.Add(serverPlugin)
}

// 下架service 一分钟后停止服务
func (ms *MicroServer) UnRegistryService() {
	if err := ms.RpcxServer.UnregisterAll(); err != nil {
		ms.Log.Err(err).Msg("UnRegistryService")
		return
	}
	ms.Log.Info().Msg("UnRegistryService success !!!!")
}

// 注册service
func (ms *MicroServer) RegistryService(rs []registry.ServiceFuncItem) {
	for _, sf := range rs {
		switch sf.SFType {
		case "func":
			ms.RpcxServer.RegisterFunction(sf.SFName, sf.SFCall, sf.SFMeta)
		case "class":
			ms.RpcxServer.RegisterName(sf.SFName, sf.SFCall, sf.SFMeta)
		}
	}
}

// 启动一个微服务监听服务
func (ms *MicroServer) StartServer() bool {
	err := ms.RpcxServer.Serve("tcp", ms.ServiceAddress)
	if err != nil {
		panic(err)
	}
	ms.Log.Info().Msg("rpcx start success")
	return true
}
