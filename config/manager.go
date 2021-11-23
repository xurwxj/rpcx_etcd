package config

import (
	"context"
	"sync"
	"time"

	"github.com/rpcxio/libkv/store"
	log "github.com/rs/zerolog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdKVWatchConfig struct {
	EtcdAddrss      []string
	DialTimeout     int
	Key             string // "/config/dev/eds"
	MergeConfigFunc func(data string)
	Log             *log.Logger
	Options         *store.Config
}

// 启动一个etcd kv watch 通过MergeConfigFunc回调全量更新数据 go StartKvWatch()
func StartKvWatch(watchConfig *EtcdKVWatchConfig, wg *sync.WaitGroup) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   watchConfig.EtcdAddrss, //etcd集群三个实例的端口
		DialTimeout: time.Duration(watchConfig.DialTimeout) * time.Second,
		Username:    watchConfig.Options.Username,
		Password:    watchConfig.Options.Password,
	})
	if err != nil {
		watchConfig.Log.Err(err).Msg("connect failed")
		return
	}
	watchConfig.Log.Debug().Msg("connect succ")
	defer cli.Close()

	writeConfig(cli, watchConfig)
	wg.Done()
	rch := cli.Watch(context.Background(), watchConfig.Key)
	for wresp := range rch { //阻塞在这里，如果没有key里没有变化，就一直停留在这里
		for _, ev := range wresp.Events {
			watchConfig.Log.Info().Str("key", string(ev.Kv.Key)).Str("value", string(ev.Kv.Value)).Msg("config update")
			watchConfig.MergeConfigFunc(string(ev.Kv.Value))
		}
	}
}

func writeConfig(cli *clientv3.Client, watchConfig *EtcdKVWatchConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, watchConfig.Key)
	cancel()
	if err != nil {
		watchConfig.Log.Err(err).Msg("get etcd config  errr")
		return
	}
	for _, v := range resp.Kvs {
		if string(v.Key) == watchConfig.Key {
			watchConfig.MergeConfigFunc(string(v.Value))
		}
	}
}
