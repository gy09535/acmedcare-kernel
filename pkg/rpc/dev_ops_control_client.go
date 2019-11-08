package rpc

import (
	"acmed.com/kernel/pkg/proto"
	"context"
	"errors"
	"google.golang.org/grpc"
	"log"
	"time"
)

type ControllerClient interface {
	StartContext() error
	StopContext() error
	GetServiceInfo() (service *proto.ServiceDto, err error)
	GetHeartBeat() (dto *proto.ResultDto, err error)
	Close() error
}

type controllerClient struct {
	conn         *grpc.ClientConn
	RetryChannel chan int
	address      string
}

var Logger *log.Logger = nil

func NewControllerClient(address string) ControllerClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect control client: %s", err.Error())
		return nil
	}

	client := &controllerClient{conn, make(chan int), address}
	client.retry()
	return client
}

func (client *controllerClient) StartContext() error {

	rpc := client.getRpc()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := rpc.Start(ctx, &proto.RequestDto{})
	if err != nil {
		return err
	}

	if !r.Success {
		return errors.New("start spring context error")
	}
	return nil
}

func (client *controllerClient) StopContext() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	rpc := client.getRpc()
	r, err := rpc.Stop(ctx, &proto.RequestDto{})
	if err != nil {
		return err
	}

	if !r.Success {
		return errors.New("stop spring context error")
	}

	return nil
}

func (client *controllerClient) GetServiceInfo() (service *proto.ServiceDto, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().GetInfo(ctx, &proto.RequestDto{})
	return r, err
}

func (client *controllerClient) GetHeartBeat() (dto *proto.ResultDto, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().Check(ctx, &proto.RequestDto{})
	return r, err
}

func (client *controllerClient) retry() {

	go func() {
		for {
			single := <-client.RetryChannel
			if single == 1 {

				if client.conn != nil {
					_ := client.conn.Close()
				}

				conn, err := grpc.Dial(client.address, grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					log.Printf("did not connect control client: %s", err.Error())
					continue
				}

				client.conn = conn
			} else {
				return
			}
		}
	}()
}

func (client *controllerClient) Close() error {
	client.RetryChannel <- 2
	return client.conn.Close()
}

func (client *controllerClient) getRpc() proto.DevOpsControllerClient {
	rpc := proto.NewDevOpsControllerClient(client.conn)
	return rpc
}
