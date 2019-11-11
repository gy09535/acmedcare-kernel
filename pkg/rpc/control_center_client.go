package rpc

import (
	"acmed.com/kernel/pkg/proto"
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type ControlCenterClient interface {
	RegisterService(service proto.Service) (result *proto.Result, err error)
	ReportHeartBeat(beat proto.HeartBeat) (result *proto.Result, err error)
	Close()
	SendRetrySingle()
}

type controlCenterClient struct {
	conn         *grpc.ClientConn
	retryChannel chan int
	address      string
}

func NewControlCenterClient(address string) ControlCenterClient {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect control center: %v", err)
		return nil
	}

	client := &controlCenterClient{conn, make(chan int), address}
	client.retry()
	return client
}

func (client *controlCenterClient) ReportHeartBeat(beat proto.HeartBeat) (result *proto.Result, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().ReportHeartBeat(ctx, &beat)
	return r, err
}

func (client *controlCenterClient) Close() {
	client.retryChannel <- 2
	err := client.conn.Close()
	if err != nil {
		Logger.Println("close control rpc error:" + err.Error())
	}
}

func (client *controlCenterClient) RegisterService(service proto.Service) (result *proto.Result, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().RegisterService(ctx, &service)
	return r, err
}

func (client *controlCenterClient) getRpc() proto.ControllerCenterClient {

	return proto.NewControllerCenterClient(client.conn)
}

func (client *controlCenterClient) SendRetrySingle() {

	client.retryChannel <- 1
}

func (client *controlCenterClient) retry() {

	go func() {
		for {
			single := <-client.retryChannel
			if single == 1 {
				Logger.Println("begin retry connection for control center")
				if client.conn != nil {
					err := client.conn.Close()
					if err != nil {
						Logger.Println("close connection control error:" + err.Error())
					}
				}

				conn, err := grpc.Dial(client.address, grpc.WithInsecure())
				if err != nil {
					Logger.Printf("did not connect control center: %v", err)
					continue
				}
				client.conn = conn
			} else {
				return
			}
		}
	}()
}
