package main

import (
	"context"
	"fmt"
	"time"

	etcd3 "go.etcd.io/etcd/client/v3"
)

type EtcdManager struct {
	Id              string
	prefix          string
	client          *etcd3.Client
	defaultTimeout  time.Duration
	keepAliveTTL    int64
	localServiceTTL int64
	node            NodeMetadata
}

func (e *EtcdManager) metadataPath(name string) string {
	return fmt.Sprintf("%s/status/%s/metadata/%s", e.prefix, e.Id, name)
}

func (e *EtcdManager) onlinePath() string {
	return fmt.Sprintf("%s/status/%s/online", e.prefix, e.Id)
}

func (e *EtcdManager) localServicePath(serviceName string) string {
	return fmt.Sprintf("%s/status/%s/services/%s", e.prefix, e.Id, serviceName)
}

func (e *EtcdManager) configPath() string {
	return fmt.Sprintf("%s/config/%s", e.prefix, e.Id)
}

func (e *EtcdManager) UpdateMetadataItem(ctx context.Context, name string, value string) error {
	ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	_, err := e.client.Put(ctx0, e.metadataPath(name), value)
	cancel()
	return err
}

func (e *EtcdManager) RefreshMetadata(ctx context.Context) error {
	e.node.Refresh()
	e.Id = e.node.Id
	return nil
}

func (e *EtcdManager) SubmitMetadata(ctx context.Context) error {
	err := e.UpdateMetadataItem(ctx, "id", e.node.Id)
	if err != nil {
		return err
	}
	err = e.UpdateMetadataItem(ctx, "ip", e.node.PublicIp)
	return err
}

func (e *EtcdManager) GetConfig(ctx context.Context) ([]byte, error) {
	ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	resp, err := e.client.Get(ctx0, e.configPath())
	cancel()
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []byte("{}"), nil
	}
	return resp.Kvs[0].Value, nil
}

func (e *EtcdManager) GetConfigNoError(ctx context.Context) []byte {
	cfgbytes, err := e.GetConfig(ctx)
	if err != nil {
		return []byte("{}")
	}
	return cfgbytes
}

func (e *EtcdManager) WatchConfig(ctx context.Context) <-chan etcd3.WatchResponse {
	return e.client.Watch(ctx, e.configPath()) // <-chan WatchResponse
}

func (e *EtcdManager) Register(ctx context.Context) (<-chan *etcd3.LeaseKeepAliveResponse, error) {
	ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	leaseResp, err := e.client.Grant(ctx0, e.keepAliveTTL)
	cancel()
	if err != nil {
		return nil, err
	}

	ctx0, cancel = context.WithTimeout(ctx, e.defaultTimeout)
	_, err = e.client.Put(ctx0, e.onlinePath(), "true", etcd3.WithLease(leaseResp.ID))
	cancel()
	if err != nil {
		return nil, err
	}

	return e.client.KeepAlive(ctx, leaseResp.ID)
}

func (e *EtcdManager) RegisterLocalService(ctx context.Context, service *LocalService) (etcd3.LeaseID, error) {
	leaseId := service.LeaseId

	// 存在已有的lease, 需检查是否已经过期
	if leaseId != 0 {
		ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
		ttlResp, err := e.client.TimeToLive(ctx0, leaseId)
		cancel()
		if err != nil {
			return 0, err
		}
		if ttlResp.TTL < 0 {
			leaseId = 0
		}
	}

	// 没有当前lease，需新建lease
	if leaseId == 0 {
		ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
		leaseResp, err := e.client.Grant(ctx0, e.localServiceTTL)
		cancel()
		if err != nil {
			return 0, err
		}
		leaseId = leaseResp.ID
	} else {
		// 有lease id，则需要更新lease
		ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
		_, err := e.client.KeepAliveOnce(ctx0, leaseId)
		cancel()
		if err != nil {
			return 0, err
		}
	}

	ctx0, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	_, err := e.client.Put(ctx0, e.localServicePath(service.Name), service.Name, etcd3.WithLease(leaseId))
	cancel()
	if err != nil {
		return 0, err
	}

	return leaseId, nil
}

func (e *EtcdManager) Close() {
	e.client.Close()
}

func NewEtcdManager(endpoint string, username string, password string, prefix string) (*EtcdManager, error) {
	nodeMetadata := NodeMetadata{}

	cli, err := etcd3.New(etcd3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
		Username:    username,
		Password:    password,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdManager{
		Id:              "",
		prefix:          prefix,
		client:          cli,
		defaultTimeout:  2 * time.Second,
		keepAliveTTL:    10,
		localServiceTTL: 20,
		node:            nodeMetadata,
	}, nil

}
