package server

import (
	"fmt"

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
	ms.Log.Info().Msg("UnRegistryService success !!!!")
	fmt.Println("UnRegistryService success !!!!")
	if err := ms.RpcxServer.UnregisterAll(); err != nil {
		fmt.Println("UnRegistryService fail !!!!")
		return
	}
}

// 注册service
func (ms *MicroServer) RegistryService(rs []registry.ServiceFuncItem) {
	for _, sf := range rs {
		switch sf.SFType {
		case "func":
			if err := ms.RpcxServer.RegisterFunction(sf.SFName, sf.SFCall, sf.SFMeta); err != nil {
				ms.Log.Err(err).Str("sf.SFName", sf.SFName).Interface(" sf.SFCall", sf.SFCall).Str("sf.SFMeta", sf.SFMeta).Msg("RegistryService:RegisterFunction !!!!")
			}
		case "class":
			if err := ms.RpcxServer.RegisterName(sf.SFName, sf.SFCall, sf.SFMeta); err != nil {
				ms.Log.Err(err).Str("sf.SFName", sf.SFName).Interface(" sf.SFCall", sf.SFCall).Str("sf.SFMeta", sf.SFMeta).Msg("RegistryService:RegisterFunction !!!!")
			}
		}
	}
}

// 启动一个微服务监听服务
func (ms *MicroServer) StartServer() (errMsg error) {
	fmt.Println("rpcx start success", ms.ServiceAddress, ms.RpcxServer)
	errMsg = ms.RpcxServer.Serve("tcp", ms.ServiceAddress)
	if errMsg != nil {
		fmt.Println("rpcx start fail", errMsg, ms.ServiceAddress)
		return errMsg
	}
	return nil
}
