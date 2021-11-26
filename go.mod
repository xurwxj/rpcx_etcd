module github.com/xurwxj/rpcx_etcd

go 1.16

require (
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/rpcxio/libkv v0.5.1-0.20210420120011-1fceaedca8a5
	github.com/rs/zerolog v1.20.0
	github.com/smallnest/rpcx v1.6.11
	github.com/soheilhy/cmux v0.1.5-0.20210205191134-5ec6847320e5 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/xurwxj/gtils v1.0.1
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0

)

replace (
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.5.0-alpha.0
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.0-alpha.0
)
