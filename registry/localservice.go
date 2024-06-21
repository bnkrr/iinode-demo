package main

import (
	"errors"
	"sync"
	"time"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	etcd3 "go.etcd.io/etcd/client/v3"
)

// @Title        localservice.go
// @Description  本地服务封装
// @Create       dlchang (2024/06/20 18:30)

// TODO: 更改Map底层实现，使用map+锁实现更原子化的register操作

type LocalServiceRegisterStatusType int

const (
	LocalServiceRegisterStatus_NEW    LocalServiceRegisterStatusType = iota // 新服务
	LocalServiceRegisterStatus_UPDATE                                       // 旧服务，但更新了内容
	LocalServiceRegisterStatus_RENEW                                        // 旧服务，未更新内容，但太久没更新
	LocalServiceRegisterStatus_EXIST                                        // 旧服务，无需操作
	LocalServiceRegisterStatus_ERROR                                        // 错误
)

func GetTimestamp() int64 {
	return time.Now().UnixMilli()
}

type LocalServiceMap struct {
	sync.Map
	minUpdateIntervalMilliSec int64
}

// 判断是否需要更新
// NEW: 新服务
// UPDATE: 旧服务，但更新了内容
// RENEW: 旧服务，未更新内容，但太久没更新了
// SAME: 旧服务，无需操作
func (m *LocalServiceMap) Register(req *pb.RegisterServiceRequest) (LocalServiceRegisterStatusType, *LocalService, error) {
	oldService, found, err := m.Get(req.Name)
	if err != nil {
		return LocalServiceRegisterStatus_ERROR, nil, err
	}
	now := GetTimestamp()
	newService := &LocalService{
		Name:        req.Name,
		Port:        int(req.Port),
		CallType:    req.CallType,
		Concurrency: int(req.Concurrency),
		TTL:         now + m.minUpdateIntervalMilliSec,
		LeaseId:     0,
	}

	// NEW: 新建service
	if !found {
		m.Store(req.Name, newService)
		return LocalServiceRegisterStatus_NEW, newService, nil
	}

	// UPDATE: service内容已更新
	if !oldService.SameMetadata(newService) {
		newService.LeaseId = oldService.LeaseId
		m.Store(req.Name, newService)
		return LocalServiceRegisterStatus_UPDATE, newService, nil
	}

	// RENEW: service太久未更新
	if oldService.TTL < now {
		newService.LeaseId = oldService.LeaseId
		m.Store(req.Name, newService)
		return LocalServiceRegisterStatus_RENEW, newService, nil
	}

	// EXIST: 已存在，无操作
	return LocalServiceRegisterStatus_EXIST, oldService, nil
}

func (m *LocalServiceMap) Get(serviceName string) (*LocalService, bool, error) {
	serviceAny, ok := m.Load(serviceName)
	if !ok {
		return nil, false, nil
	}
	service, ok := serviceAny.(*LocalService)
	if !ok {
		return nil, false, errors.New("load service err")
	}
	return service, true, nil
}

func (m *LocalServiceMap) RangeServices(f func(serviceName any, service *LocalService) bool) {
	m.Range(func(name any, value any) bool {
		service, ok := value.(*LocalService)
		if !ok {
			return true
		}
		return f(name, service)
	})
}

type LocalService struct {
	Name        string        // 用于存储service的名称
	Port        int           // 用于存储service的端口
	CallType    pb.CallType   // 用户存储service的调用类型
	Concurrency int           // 用户存储service的并发
	TTL         int64         // 下次需要注册的时间
	LeaseId     etcd3.LeaseID // lease 的id
}

func (s *LocalService) SameMetadata(other *LocalService) bool {
	return s.Name == other.Name &&
		s.Port == other.Port &&
		s.Concurrency == other.Concurrency &&
		s.CallType == other.CallType
}
