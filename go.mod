module github.com/xurwxj/rpcx_etcd

go 1.16

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.15.1 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/orcaman/concurrent-map v1.0.0
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rpcxio/libkv v0.5.1-0.20210420120011-1fceaedca8a5
	github.com/rs/zerolog v1.26.1
	github.com/smallnest/rpcx v1.7.3
	github.com/stretchr/testify v1.7.1
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xurwxj/gtils v1.0.21-0.20220323191446-a588f6533624
	go.etcd.io/etcd/client/v3 v3.5.2
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064 // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect

)

replace (
	go.etcd.io/etcd/api/v3 => go.etcd.io/etcd/api/v3 v3.5.2
	go.etcd.io/etcd/client/v3 => go.etcd.io/etcd/client/v3 v3.5.2
)
