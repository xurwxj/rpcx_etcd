package main

import (
	"flag"
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rpcxio/libkv/store"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/server"
	services "github.com/xurwxj/rpcx_etcd/demoServices"
	"github.com/xurwxj/rpcx_etcd/registry"
	serverEtcd "github.com/xurwxj/rpcx_etcd/server"
	"github.com/xurwxj/rpcx_etcd/serverplugin"
)

var (
	addr     = flag.String("addr", "localhost:8973", "server address")
	etcdAddr = flag.String("etcdAddr", "etcd.xxining3d.io:2399", "etcd address")
	basePath = flag.String("base", "/services/devme", "prefix path")
	userName = flag.String("userName", "dev", "prefix path")
	etcdPw   = flag.String("etcdPw", "123456", "prefix path")
)

func main() {
	flag.Parse()

	go StartServer()
	// time.Sleep(10 * time.Second)
	// go startClient()

	select {}

}

var ServiceClientPoolMap cmap.ConcurrentMap

func StartServer() {
	server := &serverEtcd.MicroServer{
		RpcxServer:     server.NewServer(),
		Log:            &zerolog.Logger{},
		ServiceAddress: *addr,
	}
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		UpdateInterval: 30 * time.Second,
		Options:        new(store.Config),
	}
	r.Options.Username = *userName
	r.Options.Password = *etcdPw
	r.Options.PersistConnection = true
	server.AddServerPlugin(r)
	rs := []registry.ServiceFuncItem{
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxx.mul", SFCall: services.Mul},
			SFMeta:            registry.ServiceFuncMeta{AppName: "ppp", URLName: "mul", FuncName: "Mul", URLPath: "/mul", HTTPMethod: "POST", AuthLevel: "user", ProductLines: []string{""}},
		}),
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxx.add", SFCall: services.Add},
			SFMeta:            registry.ServiceFuncMeta{AppName: "ppp", URLName: "add", FuncName: "Add", URLPath: "/add", HTTPMethod: "POST", AuthLevel: "api"},
		}),
		// registry.GetServiceFunc(registry.ServiceFuncOBJ{
		// 	ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxx.add1", SFCall: services.Add1},
		// 	SFMeta:            registry.ServiceFuncMeta{AppName: "ppp", URLName: "add1", FuncName: "Add1", URLPath: "/add1", HTTPMethod: "POST", AuthLevel: "api"},
		// }),
	}

	server.RegistryService(rs)
	go server.StartServer()
}

func stop(server *serverEtcd.MicroServer) {
	time.Sleep(15 * time.Second)
	server.UnRegistryService()
	fmt.Println("..start UnRegistryService.................")

}

// func startClient() {
// 	param := &service_client.RpcxClientConfig{
// 		BasePath:   *basePath,
// 		EtcdAddrss: []string{*etcdAddr},
// 		PoolSize:   2,
// 		FailMode:   client.Failover,
// 		SelectMode: client.RoundRobin,
// 		Option:     client.DefaultOption,
// 		Log:        &zerolog.Logger{},
// 		Options:    &store.Config{},
// 	}
// 	param.Options.Username = *userName
// 	param.Options.Password = *etcdPw
// 	service_client.InitClient(param)
// 	args := &services.Args{
// 		A: 10,
// 		B: 20,
// 	}
// 	reply := &services.Reply{}
// 	service_client.CallService("hdsw", "Add", args, reply)

// 	fmt.Println("A+ B = C", args.A, args.B, reply.C)
// 	time.Sleep(10 * time.Second)

// 	service_client.CallService("hdsw", "Mul", args, reply)

// 	fmt.Println("A* B = C", args.A, args.B, reply.C)
// }
