package pkg

import (
	"acmed.com/kernel/pkg/entity"
	"acmed.com/kernel/pkg/handler"
	"acmed.com/kernel/pkg/rabbitmq"
	"acmed.com/kernel/pkg/rpc"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	ServiceRoot          = "ServiceRoot"
	ServicePort          = "ServicePort"
	TopicName            = "acmedcare.control.topic"
	ControlCenterAddress = "ControlCenterAddress"
)

var amqpClient *rabbitmq.AmqpClient = nil
var controlClient rpc.ControllerClient = nil
var isStart = false
var centerClient rpc.ControlCenterClient = nil
var heartBeatManager *handler.HeartBeatManager = nil

func Start() {

	f, err := getLogger();
	writers := []io.Writer{
		f,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	rpc.Logger = log.New(fileAndStdoutWriter, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	if err != nil {
		fmt.Printf("create log file error,%s\r\n", err.Error())
		os.Exit(1)
	}

	defer f.Close()
	StartService()
	defer centerClient.Close()
	if centerClient != nil {
		defer centerClient.Close()
	}

	startAmqp()
	defer amqpClient.Close()

	waitForStop()
}

func getLogger() (file *os.File, err error) {
	file, err = os.OpenFile("./kernel.log", os.O_RDWR|os.O_CREATE, 0666)
	return file, err
}

func startAmqp() {

	amqpClient = &rabbitmq.AmqpClient{}
	err := amqpClient.Init("amqp://admin:admin@rabbit.acmedcare.com:5670")
	if err != nil {
		rpc.Logger.Fatalf("create amqp connection error,%s \r\n", err.Error())
	}

	serviceInfo, err := controlClient.GetServiceInfo()
	if err != nil {
		rpc.Logger.Fatalf("get service error,%s\r\n", err.Error())
	}

	route := TopicName + "." + serviceInfo.GetName() + "." + serviceInfo.GetIp() + ":" + serviceInfo.GetPort()
	// receive command from server
	amqpClient.Receive(TopicName, route, route, receiveCommand)
	rpc.Logger.Printf("amqp is started")
}

func StartService() {

	if isStart {
		rpc.Logger.Println("service is already started")
		return
	}

	closeExistingService()
	root := os.Getenv(ServiceRoot)
	if strings.EqualFold(root, "") {
		root = "/service"
	}

	path := root + "/bin/startup.sh"
	execShell(path, true)
	isStart = true
	rpc.Logger.Println("service is started")
	startRpc()
}

func startRpc() {

	port := os.Getenv(ServicePort)
	address := "127.0.0.1:"
	if port != "" {
		address = address + port
	} else {
		address = address + "9999"
	}

	retry := 5
	for retry > 0 {
		tc := time.After(time.Second * 5)
		<-tc
		rpc.Logger.Println("begin open connection for service")
		controlClient = rpc.NewControllerClient(address)
		if controlClient != nil {
			break
		}
		retry--
	}

	if controlClient == nil {
		rpc.Logger.Println("start service rpc error, now begin shutdown service")
		rpc.Logger.Printf("control port is not connected:%s\r\n", ServicePort)
		ShutDownService()
		rpc.Logger.Fatalf("service start error\r\n")
	} else {

		rpc.Logger.Println("control rpc start success")
		centerClient = rpc.NewControlCenterClient(os.Getenv(ControlCenterAddress))
		if centerClient == nil {
			rpc.Logger.Println("control center is not connected")
		} else {
			heartBeatManager = handler.NewHeartBeatManager(controlClient, centerClient)
			heartBeatManager.Monitor()
			heartBeatManager.StartReport()
		}
	}
}

func ShutDownService() {

	//close heart beat check
	if heartBeatManager != nil {
		heartBeatManager.Close()
	}

	//close  service rpc
	if controlClient != nil {
		err := controlClient.Close()
		if err != nil {
			rpc.Logger.Printf("shutdown service error,%s", err.Error())
		}
	}

	// close control center rpc
	if centerClient != nil {
		centerClient.Close()
	}

	closeExistingService()
}

func closeExistingService() {

	rpc.Logger.Println("begin close existing service")
	root := os.Getenv(ServiceRoot)
	if strings.EqualFold(root, "") {
		root = "/service"
	}

	path := root + "/bin/shutdown.sh"
	execShell(path, false)
	isStart = false
}

func execShell(path string, failClose bool) {
	cmd := exec.Command("sh", path)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()

	rpc.Logger.Println("service script msg:", out.String())
	if err != nil {
		if failClose {
			rpc.Logger.Fatalf("exec %s shell error,%s\r\n", path, err.Error())
		} else {

			rpc.Logger.Printf("exec shell error,%s\r\n", err.Error())
		}
	}
}

func receiveCommand(msg *string) {
	var serviceCommand entity.ServiceCommand
	err := json.Unmarshal([]byte(*msg), &serviceCommand)
	if err != nil {

		rpc.Logger.Printf("json convert msg error,json:%s,error: %s\r\n", *msg, err.Error())
		return
	}

	switch serviceCommand.ServiceCommandType {

	case entity.StartContext:
		err = controlClient.StartContext()
		break
	case entity.StopContext:
		err = controlClient.StopContext()
		break
	case entity.Start:
		StartService()

	case entity.Stop:
		ShutDownService()

	case entity.Restart:
		ShutDownService()
		StartService()
	}

	if err != nil {
		rpc.Logger.Printf("control service error:%s", err.Error())
	}
}

func waitForStop() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sc
	rpc.Logger.Printf("exit: signal=<%d>.", sig)
	amqpClient.Close()
	ShutDownService()
	switch sig {
	case syscall.SIGTERM:
		rpc.Logger.Printf("exit: bye :-).")
		os.Exit(0)
	default:
		rpc.Logger.Fatalf("exit: bye :-(.")
	}
}
