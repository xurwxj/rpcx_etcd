package serverplugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/rpcxio/libkv"
	"github.com/rpcxio/libkv/store"
	"github.com/smallnest/rpcx/log"
	estore "github.com/xurwxj/rpcx_etcd/store"
	etcd "github.com/xurwxj/rpcx_etcd/store/etcdv3"
)

func init() {
	etcd.Register()
}

// EtcdV3RegisterPlugin implements etcd registry.
type EtcdV3RegisterPlugin struct {
	// service address, for example, tcp@127.0.0.1:8972, quic@127.0.0.1:1234
	ServiceAddress string
	// etcd addresses
	EtcdServers []string
	// base path for rpcx server, for example com/example/rpcx
	BasePath string
	Metrics  metrics.Registry
	// Registered services
	Services       []string
	metasLock      sync.RWMutex
	metas          map[string]string
	UpdateInterval time.Duration
	Expired        time.Duration

	Options *store.Config
	kv      store.Store

	dying chan struct{}
	done  chan struct{}
}

// Start starts to connect etcd cluster
func (p *EtcdV3RegisterPlugin) Start() error {
	if p.Expired == 0 {
		p.Expired = p.UpdateInterval
	}
	if p.done == nil {
		p.done = make(chan struct{})
	}
	if p.dying == nil {
		p.dying = make(chan struct{})
	}

	if p.kv == nil {
		kv, err := libkv.NewStore(estore.ETCDV3, p.EtcdServers, p.Options)
		if err != nil {
			log.Errorf("cannot create etcd registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err := p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true, TTL: p.UpdateInterval + p.Expired})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", p.BasePath, err)
		return err
	}
	go func() {
		for {
			select {
			case <-etcd.ReConn:
				for _, name := range p.Services {
					nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
					p.metasLock.RLock()
					meta := p.metas[name]
					p.metasLock.RUnlock()

					err = p.kv.Put(nodePath, []byte(meta), &store.WriteOptions{TTL: p.UpdateInterval + p.Expired})
					if err != nil {
						log.Errorf("cannot re-create etcd path", nodePath, err)
						break
					}
				}
				log.Infof("etcd reconn........")
			}
		}
	}()
	return nil
}

// Stop unregister all services.
func (p *EtcdV3RegisterPlugin) Stop() error {
	if p.kv == nil {
		kv, err := libkv.NewStore(estore.ETCDV3, p.EtcdServers, p.Options)
		if err != nil {
			log.Errorf("cannot create etcd registry: %v", err)
			return err
		}
		p.kv = kv
	}

	for _, name := range p.Services {
		nodePath := fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
		exist, err := p.kv.Exists(nodePath)
		if err != nil {
			log.Errorf("cannot delete path %s: %v", nodePath, err)
			continue
		}
		if exist {
			p.kv.Delete(nodePath)
			log.Infof("delete path %s", nodePath)
		}
	}

	close(p.dying)
	<-p.done
	return nil
}

// HandleConnAccept handles connections from clients
func (p *EtcdV3RegisterPlugin) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("connections", p.Metrics).Mark(1)
	}
	return conn, true
}

// PreCall handles rpc call from clients
func (p *EtcdV3RegisterPlugin) PreCall(_ context.Context, _, _ string, args interface{}) (interface{}, error) {
	if p.Metrics != nil {
		metrics.GetOrRegisterMeter("calls", p.Metrics).Mark(1)
	}
	return args, nil
}

// Register handles registering event.
// this service is registered at BASE/serviceName/thisIpAddress node
func (p *EtcdV3RegisterPlugin) Register(name string, rcvr interface{}, metadata string) (err error) {
	if strings.TrimSpace(name) == "" {
		err = errors.New("Register service `name` can't be empty")
		return
	}

	if p.kv == nil {
		etcd.Register()
		kv, err := libkv.NewStore(estore.ETCDV3, p.EtcdServers, nil)
		if err != nil {
			log.Errorf("cannot create etcd registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)
	err = p.kv.Put(nodePath, []byte(metadata), &store.WriteOptions{TTL: p.UpdateInterval + p.Expired})
	if err != nil {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	p.Services = append(p.Services, name)

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	p.metas[name] = metadata
	p.metasLock.Unlock()
	return
}

func (p *EtcdV3RegisterPlugin) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	return p.Register(serviceName, fn, metadata)
}

func (p *EtcdV3RegisterPlugin) Unregister(name string) (err error) {
	if len(p.Services) == 0 {
		return nil
	}

	if strings.TrimSpace(name) == "" {
		err = errors.New("Register service `name` can't be empty")
		return
	}

	if p.kv == nil {
		etcd.Register()
		kv, err := libkv.NewStore(estore.ETCDV3, p.EtcdServers, nil)
		if err != nil {
			log.Errorf("cannot create etcd registry: %v", err)
			return err
		}
		p.kv = kv
	}

	err = p.kv.Put(p.BasePath, []byte("rpcx_path"), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", p.BasePath, err)
		return err
	}

	nodePath := fmt.Sprintf("%s/%s", p.BasePath, name)
	err = p.kv.Put(nodePath, []byte(name), &store.WriteOptions{IsDir: true})
	if err != nil && !strings.Contains(err.Error(), "Not a file") {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	nodePath = fmt.Sprintf("%s/%s/%s", p.BasePath, name, p.ServiceAddress)

	err = p.kv.Delete(nodePath)
	if err != nil {
		log.Errorf("cannot create etcd path %s: %v", nodePath, err)
		return err
	}

	if len(p.Services) > 0 {
		var services = make([]string, 0, len(p.Services)-1)
		for _, s := range p.Services {
			if s != name {
				services = append(services, s)
			}
		}
		p.Services = services
	}

	p.metasLock.Lock()
	if p.metas == nil {
		p.metas = make(map[string]string)
	}
	delete(p.metas, name)
	p.metasLock.Unlock()
	return
}
