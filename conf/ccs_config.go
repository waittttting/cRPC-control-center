package conf

type CCSConf struct {
	Server Server
}

type Server struct {
	Port int
	// 接收 socket 队列的长度
	ReceiveSocketChanLen int
}
