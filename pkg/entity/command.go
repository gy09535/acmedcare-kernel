package entity

type ServiceCommandType int

const (
	Start = iota
	Stop
	Restart
	StartContext
	StopContext
)

type ServiceCommand struct {
	ServiceCommandType int    `json:"serviceCommandType"`
	ServiceName        string `json:"serviceName"`
}
