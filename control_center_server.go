package main

import (
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cx-rpc/control_center/conf"
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
}

// 集群通信
func (ccs *ControlCenterServer) clusterCommunication() {

}

func (ccs *ControlCenterServer) acceptSocket() {
	// 获取监听地址
	addr, err := net.ResolveTCPAddr("tcp4", ":" + strconv.Itoa(ccs.config.Server.Port))
	if err != nil {
		logrus.Fatalf("resolve TCP addr error [%v]", err)
	}
	// 获取监听器
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logrus.Fatalf("listen TCP error [%v]", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("access tcp_server occur panic: %v", r)
			}
		}()
		for {
			// 接收长连接
			socket, err := ln.AcceptTCP()
			if err != nil {
				logrus.Errorf("accept TCP error [%v]", err)
			}
			timer := time.NewTimer( 200 * time.Millisecond)
			select {
			case ccs.receiveSocketChan <- socket:
				timer.Reset(0)
				logrus.Infof("socket accepted [%v]", socket.RemoteAddr())
			case <-timer.C:
				timer.Reset(0)
				socket.Close()
				logrus.Errorf("receive socket timeout")
			}
		}
	}()
}
