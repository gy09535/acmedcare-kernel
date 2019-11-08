package rpc

import (
	"acmed.com/kernel/pkg/proto"
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

type ControlCenterClient interface {
	RegisterService(service proto.Service) (result proto.Result, err error)
	ReportHeartBeat(beat proto.HeartBeat) (result proto.Result, err error)
	Close()
}

type controlCenterClient struct {
	conn *grpc.ClientConn
}

func NewControlCenterClient(address string) ControlCenterClient {

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Printf("did not connect control center: %v", err)
		return nil
	}

	return &controlCenterClient{conn}
}

func (client *controlCenterClient) ReportHeartBeat(beat proto.HeartBeat) (result proto.Result, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().ReportHeartBeat(ctx, &beat)
	return *r, err;
}

func (client *controlCenterClient) Close() {
	client.conn.Close()
}

func (client *controlCenterClient) RegisterService(service proto.Service) (result proto.Result, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := client.getRpc().RegisterService(ctx, &service)
	return *r, err;
}

func (client *controlCenterClient) getRpc() proto.ControllerCenterClient {

	return proto.NewControllerCenterClient(client.conn)
}
