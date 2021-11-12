package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	service_client "github.com/xurwxj/rpcx_etcd/client"
	services "github.com/xurwxj/rpcx_etcd/demoServices"
	"github.com/xurwxj/rpcx_etcd/discovery"
	"github.com/xurwxj/rpcx_etcd/registry"
	serverEtcd "github.com/xurwxj/rpcx_etcd/server"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	addr     = flag.String("addr", "localhost:8973", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:22379", "etcd address")
	basePath = flag.String("base", "/services/dev", "prefix path")
)

func main() {

	flag.Parse()

	// go StartServer()
	// time.Sleep(2 * time.Second)
	// go startClient()
	go watchServices()
	// time.Sleep(1 * time.Second)

	// go test()

	select {}

}

var ServiceClientPoolMap cmap.ConcurrentMap

func test() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:22379"}, //etcd集群三个实例的端口
		DialTimeout: 2 * time.Second,
	})

	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}

	fmt.Println("connect succ")

	defer cli.Close()

	rch := cli.Watch(context.Background(), "/config/dev/eds") //阻塞在这里，如果没有key里没有变化，就一直停留在这里
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

func watchServices() {
	discovery.StartWatchServices("/services", "dev", []string{*etcdAddr})
}

func StartServer() {
	// MicroServer
	server := &serverEtcd.MicroServer{
		RpcxServer:     server.NewServer(),
		Log:            &zerolog.Logger{},
		ServiceAddress: *addr,
	}
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		UpdateInterval: time.Minute,
	}
	r.UpdateInterval = -1
	server.AddServerPlugin(r)
	rs := []registry.ServiceFuncItem{
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxxxx.sm.use", SFCall: services.Mul},
			SFMeta:            registry.ServiceFuncMeta{URLName: "softModularUse", FuncName: "SoftModularUse", URLPath: "/sm/use", HTTPMethod: "post", AuthLevel: "user", ProductLines: []string{"dental"}},
		}),
		registry.GetServiceFunc(registry.ServiceFuncOBJ{
			ServiceFuncCommon: registry.ServiceFuncCommon{SFType: "func", SFName: "xxxxxx.product.ed.status", SFCall: services.Add},
			SFMeta:            registry.ServiceFuncMeta{URLName: "productEDStatus", FuncName: "ProductEDStatus", URLPath: "/product/ed/status", HTTPMethod: "POST", AuthLevel: "api"},
		}),
	}
	server.RegistryService(rs)
	go server.StartServer()
	go stop(server)
}

func stop(server *serverEtcd.MicroServer) {
	time.Sleep(15 * time.Second)
	server.UnRegistryService()
	fmt.Println("..start UnRegistryService...............")

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
	}
	service_client.InitClient(param)
	for {
		args := &services.Args{
			A: 10,
			B: 20,
		}
		reply := &services.Reply{}
		service_client.CallService(context.Background(), "xxxxxx.product.ed.status", "Add", args, reply)

		fmt.Println("A* B = C", args.A, args.B, reply.C)
		time.Sleep(2 * time.Second)
	}

}
