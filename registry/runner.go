package main

// @Title        runner.go
// @Description  在本地处理输入输出，同时按照配置调用local service
// @Create       dlchang (2024/04/03 15:00)

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	pb "github.com/bnkrr/iinode-demo/pb_autogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InputMessage interface {
	Msg() string
	Ack()
}

type InputConnector interface {
	SetInputChannel(context.Context, chan InputMessage)
}

type OutputConnector interface {
	SetOutputChannel(context.Context, chan string)
}

type Runner struct {
	name          string
	port          int
	concurrency   int
	callType      pb.CallType
	config        *ConfigRunner
	cancel        context.CancelFunc
	wgroup        *sync.WaitGroup
	serviceClient pb.ServiceClient
	inconn        InputConnector
	outconn       OutputConnector
	inputCh       chan InputMessage
	outputCh      chan string
	callEndChn    map[int](chan struct{})
}

func (r *Runner) Worker(ctx context.Context, wid int) error {
	defer r.wgroup.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case input := <-r.inputCh:
			r.CallGeneric(ctx, wid, input)
		}
	}
}

func (r *Runner) CallGeneric(ctx context.Context, wid int, input InputMessage) {
	// call
	if r.callType == pb.CallType_STREAM {
		r.CallStream(ctx, input.Msg())
	} else if r.callType == pb.CallType_ASYNC {
		r.CallAsync(ctx, wid, input.Msg())
	} else {
		r.Call(ctx, input.Msg())
	}
	input.Ack()
}

func (r *Runner) Call(ctx context.Context, input string) {
	resp, err := r.serviceClient.Call(ctx, &pb.ServiceCallRequest{Input: input})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
	} else {
		r.outputCh <- resp.Output
	}
}

func (r *Runner) CallStream(ctx context.Context, input string) {
	stream, err := r.serviceClient.CallStream(ctx, &pb.ServiceCallRequest{Input: input})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
	} else {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%v.CallStream, %v", r.serviceClient, err)
			}
			r.outputCh <- resp.Output
		}
	}
}

func (r *Runner) CallAsync(ctx context.Context, wid int, input string) {
	resp, err := r.serviceClient.CallAsync(ctx, &pb.ServiceCallRequest{RequestId: int32(wid), Input: input})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
		return
	} else {
		log.Printf("async call (id:%v), %v\n", wid, resp.Output)
	}
	ch, ok := r.callEndChn[wid]
	if !ok {
		log.Panicf("err find channel, %v\n", wid)
	}
	<-ch
}

func (r *Runner) ReceiveResult(req *pb.SubmitResultRequest) {
	// 无法获取初始化时的context，所以要提前初始化output channel
	// r.ioconn.OutputChannel(context.Background()) <- req.GetOutput()
	r.outputCh <- req.GetOutput()
	if req.GetEnd() {
		wid := int(req.RequestId)
		ch, ok := r.callEndChn[wid]
		if !ok {
			log.Panicf("err find channel %v", wid)
		}
		ch <- struct{}{}
	}
}

func (r *Runner) StartWithIOManager(ioManager *IOManager) error {
	r.wgroup = &sync.WaitGroup{}
	r.wgroup.Add(r.concurrency) // add outside of go func (before everything), in case waitgroup wait ends unexpectedly

	// set up io connector
	inconn, err := ioManager.GetInputConnector(r.config, r.wgroup)
	if err != nil {
		return err
	}
	r.inconn = inconn
	outconn, err := ioManager.GetOutputConnector(r.config, r.wgroup)
	if err != nil {
		return err
	}
	r.outconn = outconn

	ctx, cancel := context.WithCancel(context.Background())

	r.inputCh = make(chan InputMessage)
	r.inconn.SetInputChannel(ctx, r.inputCh)
	r.outputCh = make(chan string)
	r.outconn.SetOutputChannel(ctx, r.outputCh)

	for i := 1; i <= r.concurrency; i++ {
		r.callEndChn[i] = make(chan struct{})
		go r.Worker(ctx, i)
	}

	r.cancel = cancel
	return nil
}

func (r *Runner) Stop() error {
	if r.cancel == nil {
		return errors.New("no cancel function")
	}
	r.cancel()
	return nil
}

func (r *Runner) Serve() error {
	log.Println("started")
	r.wgroup.Wait()
	log.Println("stopped")
	return nil
}

func NewServiceClient(url string) (pb.ServiceClient, error) {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return pb.NewServiceClient(conn), nil
}

func NewRunner(s *LocalService, c *ConfigRunner) (*Runner, error) {
	rpcClient, err := NewServiceClient(fmt.Sprintf("localhost:%d", s.Port))
	if err != nil {
		return nil, err
	}
	r := &Runner{
		name:          s.Name,
		port:          s.Port,
		concurrency:   s.Concurrency,
		callType:      s.CallType,
		config:        c,
		serviceClient: rpcClient,
		callEndChn:    make(map[int]chan struct{}),
	}
	return r, nil
}
