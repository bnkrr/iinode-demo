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

type Runner struct {
	name          string
	port          int
	concurrency   int
	callType      int
	inputChannel  chan string
	outputChannel chan string
	cancel        context.CancelFunc
	wgroup        *sync.WaitGroup
	serviceClient pb.ServiceClient
}

func (r *Runner) Serve() error {
	r.wgroup.Wait()
	log.Println("stopped.")
	return nil
}

func (r *Runner) Worker(ctx context.Context, wid int) error {
	r.wgroup.Add(1)
	defer r.wgroup.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case input := <-r.inputChannel:
			r.CallGeneric(ctx, input)
		}
	}
}

func (r *Runner) OutputWorker() error {
	r.wgroup.Add(1)
	defer r.wgroup.Done()

	for output := range r.outputChannel {
		log.Printf("output: %s\n", output)
	}
	return nil
}

func (r *Runner) NewServiceClient() error {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", r.port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	r.serviceClient = pb.NewServiceClient(conn)
	return nil
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
		r.outputChannel <- resp.Output
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
			r.outputChannel <- resp.Output
		}
	}
}

func (r *Runner) TestInputWorker(cnt int) error {
	r.wgroup.Add(1)
	defer r.wgroup.Done()

	for i := 1; i < cnt; i++ {
		r.inputChannel <- fmt.Sprintf("task %v", i)
	}
	return nil
}

func (r *Runner) Start() error {
	r.wgroup = &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	for i := 1; i <= r.concurrency; i++ {
		go r.Worker(ctx, i)
	}
	go r.TestInputWorker(100)
	go r.OutputWorker()

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

func NewRunner(s *LocalService, c *ConfigRunner) *Runner {
	callType := 0
	if s.ReturnStream {
		callType = 2
	} else {
		callType = 1
	}
	r := &Runner{
		name:          s.Name,
		port:          s.Port,
		concurrency:   s.Concurrency,
		callType:      callType,
		inputChannel:  make(chan string, 1000),
		outputChannel: make(chan string, 1000),
	}
	r.NewServiceClient()
	return r
}
