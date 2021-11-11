package server

import (
	"strings"

	log "github.com/rs/zerolog"

	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
	"github.com/xurwxj/rpcx_etcd/registry"
)

var s *server.Server

// 启动一个微服务监听服务，并将所具有的的微服务接口注册到etcd集群
func StartServer(serverPlugin *serverplugin.EtcdV3RegisterPlugin, rs []*registry.ServiceFuncItem, log *log.Logger) bool {
	s = server.NewServer()
	addRegistryPlugin(s, serverPlugin, log)
	for _, sf := range rs {
		switch sf.SFType {
		case "func":
			s.RegisterFunction(sf.SFName, sf.SFCall, sf.SFMeta)
		case "class":
			s.RegisterName(sf.SFName, sf.SFCall, sf.SFMeta)
		}
	}
	adderss := serverPlugin.ServiceAddress
	strs := strings.Split(adderss, "@")
	if len(strs) != 2 {
		return false
	}
	err := s.Serve("tcp", strs[1])
	if err != nil {
		panic(err)
	}
	log.Info().Msg("rpcx start success")
	return true
}

func addRegistryPlugin(s *server.Server, serverPlugin *serverplugin.EtcdV3RegisterPlugin, log *log.Logger) {
	err := serverPlugin.Start()
	if err != nil {
		log.Err(err).Msg("addRegistryPlugin")
	}
	s.Plugins.Add(serverPlugin)
}

func StopServer() {
	s.Close()
}
