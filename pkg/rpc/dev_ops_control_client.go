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
	SendRetrySingle()
}

type controllerClient struct {
	conn         *grpc.ClientConn
	retryChannel chan int
	address      string
}

var Logger *log.Logger = nil

func NewControllerClient(address string) ControllerClient {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
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
			single := <-client.retryChannel
			if single == 1 {

				Logger.Println("begin retry connection for control rpc")
				if client.conn != nil {
					err := client.conn.Close()
					if err != nil {
						Logger.Println("close control connection conn error:" + err.Error())
					}
				}

				conn, err := grpc.Dial(client.address, grpc.WithInsecure())
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

func (client *controllerClient) SendRetrySingle() {
	client.retryChannel <- 1
}

func (client *controllerClient) Close() error {
	client.retryChannel <- 2
	return client.conn.Close()
}

func (client *controllerClient) getRpc() proto.DevOpsControllerClient {
	rpc := proto.NewDevOpsControllerClient(client.conn)
	return rpc
}
