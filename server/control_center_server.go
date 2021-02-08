package server

import (
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cRPC-common/tcp"
	"github.com/waittttting/cRPC-control-center/conf"
	"net"
	"strconv"
	"time"
)

type ControlCenterServer struct {
	config *conf.CCSConf
	receiveSocketChan chan *net.TCPConn

}

func NewControlCenterServer(config *conf.CCSConf) *ControlCenterServer {
	return &ControlCenterServer{
		config: config,
		receiveSocketChan: make(chan *net.TCPConn, config.Server.ReceiveSocketChanLen),
	}
}

// 控制中心集群中 各个节点如何进行通信 ？ Gossip 协议？
// https://github.com/hashicorp/memberlist Gossip 的一个实现
// 各个服务节点 通过 dns 获取 控制中心的 IP
// dns 系统是怎么工作的，一个网址有多个可提供服务的节点，是随机提供其中一个吗？新增一个服务节点，会立刻生效吗

// 每个 goroutine 内都要使用 recover 吗, go 的 recover

func (ccs *ControlCenterServer) Start() {

	// ccs 集群通信
	ccs.clusterCommunication()
	// 启动端口监听
	ccs.acceptSocket()
	// 处理接收到的 socket
	logrus.Info("control server started")
}

// 集群通信
func (ccs *ControlCenterServer) clusterCommunication() {
	// todo: 集群内部依赖 Gossip 协议 进行通信
}

func (ccs *ControlCenterServer) acceptSocket() {

	tcp.AcceptSocket(strconv.Itoa(ccs.config.Server.Port), ccs.receiveSocketChan, 3 * time.Second)
}

func (ccs *ControlCenterServer)handleSocket() {

	for socket := range ccs.receiveSocketChan {
		logrus.Infof("socket addr = %v", socket.RemoteAddr())
	}
}
