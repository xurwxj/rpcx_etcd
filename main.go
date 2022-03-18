package main

import (
	"flag"
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rpcxio/libkv/store"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	service_client "github.com/xurwxj/rpcx_etcd/client"
	services "github.com/xurwxj/rpcx_etcd/demoServices"
	"github.com/xurwxj/rpcx_etcd/registry"
	serverEtcd "github.com/xurwxj/rpcx_etcd/server"
	"github.com/xurwxj/rpcx_etcd/serverplugin"
)

var (
	addr     = flag.String("addr", "localhost:8973", "server address")
	etcdAddr = flag.String("etcdAddr", "127.0.0.1:2379", "etcd address")
	basePath = flag.String("base", "/services/dev111", "prefix path")
)

func main() {
	flag.Parse()

	go StartServer()
	time.Sleep(20 * time.Second)
	go startClient()

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
	r.Options.PersistConnection = true
	server.AddServerPlugin(r)
	rs := []registry.ServiceFuncItem{
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxxxx.Mul", SFCall: services.Mul},
			SFMeta:            registry.ServiceFuncMeta{URLName: "mul", FuncName: "Mul", URLPath: "/mul", HTTPMethod: "POST", AuthLevel: "user", ProductLines: []string{""}},
		}),
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxxxx.Add", SFCall: services.Add},
			SFMeta:            registry.ServiceFuncMeta{URLName: "add", FuncName: "Add", URLPath: "/add", HTTPMethod: "POST", AuthLevel: "api"},
		}),
	}

	server.RegistryService(rs)
	go server.StartServer()
}

func startClient() {
	param := &service_client.RpcxClientConfig{
		BasePath:   *basePath,
		EtcdAddrss: []string{*etcdAddr},
		PoolSize:   2,
		FailMode:   client.Failover,
		SelectMode: client.RoundRobin,
		Option:     client.DefaultOption,
		Log:        &zerolog.Logger{},
		Options:    &store.Config{},
	}
	service_client.InitClient(param)
	args := &services.Args{
		A: 10,
		B: 20,
	}
	reply := &services.Reply{}
	service_client.CallService("xxxxxx.Add", "Add", args, reply)

	fmt.Println("A+ B = C", args.A, args.B, reply.C)
	time.Sleep(10 * time.Second)

	service_client.CallService("xxxxxx.Mul", "Mul", args, reply)

	fmt.Println("A* B = C", args.A, args.B, reply.C)
}
