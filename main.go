package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/xurwxj/rpcx_etcd/discovery"

	"github.com/rpcxio/rpcx-etcd/serverplugin"
	example "github.com/rpcxio/rpcx-examples"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	services "github.com/xurwxj/rpcx_etcd/demoServices"
	"github.com/xurwxj/rpcx_etcd/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	addr     = flag.String("addr", "localhost:8973", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:22379", "etcd address")
	basePath = flag.String("base", "/services/dev/dental", "prefix path")
)

func main() {

	flag.Parse()
	go watchServices()
	time.Sleep(1 * time.Second)

	// go test()
	// go discovery11()
	go startServer()
	// time.Sleep(5 * time.Second)

	// go startClient()
	select {}

}

func startServer() {
	s := server.NewServer()
	addRegistryPlugin(s)
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
	for _, r := range rs {
		s.RegisterFunction(r.SFName, r.SFCall, r.SFMeta)
	}
	err := s.Serve("tcp", *addr)
	if err != nil {
		panic(err)
	}
}

func startClient() {
	flag.Parse()
	ServiceClientPoolMap = cmap.New()

	for {
		args := &example.Args{
			A: 10,
			B: 20,
		}
		reply := &example.Reply{}
		xclient := getXclient("xxxxxx.product.ed.status")

		err := xclient.Call(context.Background(), "Add", args, reply)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}

		log.Printf("%d * %d = %d", args.A, args.B, reply.C)
		time.Sleep(time.Second)
	}
}

var ServiceClientPoolMap cmap.ConcurrentMap

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       *basePath,
		UpdateInterval: time.Minute,
	}
	r.UpdateInterval = -1
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}

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

func getXclient(name string) client.XClient {
	if xc, has := ServiceClientPoolMap.Get(name); has {
		x := xc.(*client.XClientPool).Get()
		if x != nil {
			return x
		}
	}
	fmt.Println("getXclient:New")

	d, err := etcd_client.NewEtcdV3Discovery(*basePath, name, []string{*etcdAddr}, true, nil)
	a := d.GetServices()

	fmt.Println("....", err, a)

	xclientOptions := client.DefaultOption

	xclientPool := client.NewXClientPool(2, name, client.Failover, client.RoundRobin, d, xclientOptions)
	ServiceClientPoolMap.Set(name, xclientPool)
	return xclientPool.Get()
}
func watchServices() {
	discovery.StartWatchServices("/services", "dev", []string{*etcdAddr})

}
