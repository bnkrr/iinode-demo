package main

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

type IOConnecter interface {
	InputChannel(context.Context) chan string
	OutputChannel(context.Context) chan string
}

type Runner struct {
	name          string
	port          int
	concurrency   int
	callType      int
	config        *ConfigRunner
	cancel        context.CancelFunc
	wgroup        *sync.WaitGroup
	serviceClient pb.ServiceClient
	ioconn        IOConnecter
}

func (r *Runner) Worker(ctx context.Context, wid int) error {
	defer r.wgroup.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case input := <-r.ioconn.InputChannel(ctx):
			r.CallGeneric(ctx, input)
		}
	}
}

func (r *Runner) CallGeneric(ctx context.Context, input string) {
	// call
	if r.callType == 2 {
		r.CallStream(ctx, input)
	} else {
		r.Call(ctx, input)
	}
}

func (r *Runner) Call(ctx context.Context, input string) {
	resp, err := r.serviceClient.Call(ctx, &pb.ServiceCallRequest{Input: input})
	if err != nil {
		log.Printf("call err, try again later: %v\n", err)
	} else {
		r.ioconn.OutputChannel(ctx) <- resp.Output
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
			r.ioconn.OutputChannel(ctx) <- resp.Output
		}
	}
}

func (r *Runner) Start() error {
	r.wgroup = &sync.WaitGroup{}
	r.wgroup.Add(r.concurrency) // add one here, in case waitgroup wait ends unexpectedly

	var ioconn IOConnecter
	var err error
	if r.config.IOType == "mq" {
		ioconn, err = NewIOConnectorMQ(r.config, r.wgroup)
	} else {
		ioconn, err = NewIOConnectorFile(r.config, r.wgroup)
	}
	if err != nil {
		return err
	}
	r.ioconn = ioconn

	ctx, cancel := context.WithCancel(context.Background())

	for i := 1; i <= r.concurrency; i++ {
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
	callType := 0
	if s.ReturnStream {
		callType = 2
	} else {
		callType = 1
	}

	rpcClient, err := NewServiceClient(fmt.Sprintf("localhost:%d", s.Port))
	if err != nil {
		return nil, err
	}
	r := &Runner{
		name:          s.Name,
		port:          s.Port,
		concurrency:   s.Concurrency,
		callType:      callType,
		config:        c,
		serviceClient: rpcClient,
	}
	return r, nil
}
