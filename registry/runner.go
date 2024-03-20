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
	config        *ConfigRunner
	inputChannel  chan string
	outputChannel chan string
	cancel        context.CancelFunc
	wgroup        *sync.WaitGroup
	serviceClient pb.ServiceClient
	mqClient      *MQClient
}

func (r *Runner) Serve() error {
	log.Println("started")
	r.wgroup.Wait()
	log.Println("stopped")
	return nil
}

func (r *Runner) InputWorker() {
	// r.wgroup.Add(1)  // already added outside
	defer r.wgroup.Done()

	msgChannel, err := r.mqClient.GetInputQueue(r.config.InputQueue)
	if err != nil {
		log.Panicf("%v", err)
	}
	for msg := range msgChannel {
		input := string(msg.Body)
		r.inputChannel <- input
	}
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

	err := r.mqClient.CheckOutputQueue(r.config.OutputExchange, r.config.OutputRoutingKey)
	if err != nil {
		log.Fatalf("cannot connect to output queue, %v", err)
	}
	for output := range r.outputChannel {
		err = r.mqClient.Publish(r.config.OutputExchange, r.config.OutputRoutingKey, output)
		if err != nil {
			log.Printf("publish error, %v", err)
		}
		log.Printf("output: %s\n", output)
	}
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

func (r *Runner) Start() error {
	r.wgroup = &sync.WaitGroup{}
	r.wgroup.Add(1) // add one here, in case waitgroup wait ends unexpectedly

	ctx, cancel := context.WithCancel(context.Background())

	go r.InputWorker()
	for i := 1; i <= r.concurrency; i++ {
		go r.Worker(ctx, i)
	}
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

	mqClient, err := NewMQClient(c.MqUrl)
	if err != nil {
		return nil, err
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
		inputChannel:  make(chan string),
		outputChannel: make(chan string, 1000),
		mqClient:      mqClient,
		serviceClient: rpcClient,
	}
	return r, nil
}
