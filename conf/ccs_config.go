package conf

type CCSConf struct {
	Server Server
	Redis Redis
}

type Server struct {
	Port int
	// 接收 socket 队列的长度
	ReceiveSocketChanLen int
}

type Redis struct {
	Host string
	Pwd string
	Index int
}
