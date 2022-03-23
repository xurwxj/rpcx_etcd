module github.com/xurwxj/rpcx_etcd

go 1.16

require (
	github.com/google/btree v1.0.1 // indirect
	github.com/orcaman/concurrent-map v1.0.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rpcxio/libkv v0.5.1-0.20210420120011-1fceaedca8a5
	github.com/rs/zerolog v1.26.1
	github.com/smallnest/rpcx v1.7.3
	github.com/stretchr/testify v1.7.1
	github.com/xurwxj/gtils v1.0.20
	go.etcd.io/etcd/client/v3 v3.5.2

)

replace (
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.5.2
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.2
)
