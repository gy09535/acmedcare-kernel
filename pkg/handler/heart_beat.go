package handler

import (
	"acmed.com/kernel/pkg/proto"
	"acmed.com/kernel/pkg/rpc"
	"time"
)

type HeartBeatManager struct {
	controlClient       rpc.ControllerClient
	healthCheckChannel  chan int
	heartBeat           proto.HeartBeat
	controlCenterClient rpc.ControlCenterClient
	reportChannel       chan int
}

func NewHeartBeatManager(controlClient rpc.ControllerClient, client rpc.ControlCenterClient) *HeartBeatManager {

	return &HeartBeatManager{controlClient: controlClient, controlCenterClient: client, heartBeat: proto.HeartBeat{}, healthCheckChannel: make(chan int), reportChannel: make(chan int)}
}

func (manager *HeartBeatManager) Monitor() {

	serviceInfo, err := manager.controlClient.GetServiceInfo()
	if err != nil {
		rpc.Logger.Fatalf("get service info error:" + err.Error())
	}

	if serviceInfo != nil {
		manager.heartBeat.ServiceName = serviceInfo.Name
		manager.heartBeat.Ip = serviceInfo.Ip
		manager.heartBeat.Port = serviceInfo.Port
		manager.heartBeat.CurrentTime = 0
		service := proto.Service{Name: serviceInfo.Name, Ip: serviceInfo.Ip, Port: serviceInfo.Port}
		_, err = manager.controlCenterClient.RegisterService(service)
	}

	if err != nil {
		rpc.Logger.Fatalf("register service error:" + err.Error())
	}

	go func() {

		sendFailCount := 0
		for {
			select {
			case <-time.After(time.Second * 10):

				rpc.Logger.Println("begin get heartbeat")
				serviceBeat, err := manager.controlClient.GetHeartBeat()
				if err != nil {
					rpc.Logger.Println("get heartbeat error:" + err.Error())
					manager.heartBeat.FailCount++
					sendFailCount++
					if sendFailCount > 10 {
						manager.controlClient.SendRetrySingle()
						sendFailCount = 0
					}
					break
				}

				sendFailCount = 0
				if serviceBeat.Success {
					manager.heartBeat.SuccessCount++
				} else {
					manager.heartBeat.FailCount++
				}

				break
			case <-manager.healthCheckChannel:
				rpc.Logger.Println("closed health check channel")
				return
			}
		}
	}()
}

func (manager *HeartBeatManager) StartReport() {
	go func() {
		sendFailCount := 0
		for {
			select {
			case <-time.After(time.Minute):
				rpc.Logger.Println("report heartbeat from service:" + manager.heartBeat.ServiceName + " to control center")
				_, err := manager.controlCenterClient.ReportHeartBeat(manager.heartBeat)

				if err != nil {
					sendFailCount++
					rpc.Logger.Printf("report heartbeat error:%s", err.Error())

					if sendFailCount > 2 {
						manager.controlCenterClient.SendRetrySingle()
						sendFailCount = 0
					}
					break
				}
				sendFailCount = 0
				manager.heartBeat.SuccessCount = 0
				manager.heartBeat.FailCount = 0
				manager.heartBeat.CurrentTime = 0
				break
			case <-manager.reportChannel:
				return
			}
		}
	}()
}

func (manager *HeartBeatManager) Close() {

	manager.healthCheckChannel <- 1
	manager.reportChannel <- 1
}
