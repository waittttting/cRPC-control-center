package server

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cRPC-common/tcp"
	"github.com/waittttting/cRPC-control-center/conf"
	"net"
	"strconv"
	"sync"
	"time"
)

// 注册锁 key
const lockRegister = "lockRegister"

type ControlCenterServer struct {
	config *conf.CCSConf
	receiveSocketChan chan *net.TCPConn
	clientMap map[string][]*ServiceClient
	lock sync.RWMutex
}

type ServiceClient struct {
	Conn *tcp.Connection
}

func NewControlCenterServer(config *conf.CCSConf) *ControlCenterServer {
	return &ControlCenterServer{
		config: config,
		receiveSocketChan: make(chan *net.TCPConn, config.Server.ReceiveSocketChanLen),
		clientMap: make(map[string][]*ServiceClient),
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
	// 处理接收到的 socket
	ccs.handleSocket()
	// 启动端口监听
	ccs.acceptSocket()
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

	go func() {
		for socket := range ccs.receiveSocketChan {
			logrus.Infof("socket addr = %v", socket.RemoteAddr())
			ccs.checkRegisterMsg(socket)
		}
	}()
}

func (ccs *ControlCenterServer)checkRegisterMsg(socket *net.TCPConn) {

	// 在注册的过程中，如果出现了 err 直接 调用 socket.Close(),
	// 注册到 redis 失败 结束本次注册，关闭 socket，等待 client 重连
	conn := tcp.NewConnection(socket)
	msg, err := conn.Receive(0)
	if err != nil {
		logrus.Errorf("register error address : [%v]", socket.RemoteAddr())
		socket.Close() // 调用 socket.Close(), Client 调用 read 时会直接报错 connection reset by peer
		return
	}

	// 注册 Service 的 IP 注册到 redis, key = serviceName + serviceVersion
	// Client 在注册与 Server 的 tcp 连接的时候，会根据 serviceName + serviceVersion 找到对应的IP 列表
	err = ccs.redisOp(msg, socket, redisOpSAddServerIp)
	if err != nil {
		logrus.Errorf("register to redis error: [%v]", err)
		socket.Close()
		return
	}

	if err = conn.Send(tcp.MsgRegisterPong()); err != nil {
		logrus.Errorf("send MsgRegisterPong error: [%v]", err)
		err = ccs.redisOp(msg, socket, redisOpSRemServerIp)
		if err != nil {
			// 报警，人工介入
			logrus.Errorf("delete client ip from redis error: [%v]", err)
		}
		socket.Close()
		return
	}

	if v, ok := ccs.clientMap[msg.Header.ServerName]; !ok {
		cs := []*ServiceClient{{Conn: conn}}
		ccs.clientMap[msg.Header.ServerName] = cs
	} else {
		// 虽然此处的 map 操作都是顺序的，但是可能会有 server 与 client 的 tcp 断开的情况，在这种情况下，也会写操作这个 map
		ccs.lock.Lock()
		defer ccs.lock.Unlock()
		v = append(v, &ServiceClient{Conn: conn})
	}
	// 连接建立完成


}

type redisOp int

const (
	redisOpSAddServerIp = 1
	redisOpSRemServerIp = 2
)

func (ccs *ControlCenterServer)redisOp(message *tcp.Message, socket *net.TCPConn, op redisOp) error {

	ip := socket.RemoteAddr().String() // todo: ip 转换
	client := redis.NewClient(&redis.Options{
		Addr:     ccs.config.Redis.Host,
		Password: ccs.config.Redis.Pwd,
		DB:       ccs.config.Redis.Index,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", message.Header.ServerName, message.Header.ServerVersion)

	switch op {
	case redisOpSAddServerIp:
		logrus.Infof("register info: key: %v, ip: %v", key, socket.RemoteAddr().String())
		_, err = client.SAdd(key, ip).Result()
	case redisOpSRemServerIp:
		logrus.Infof("delete info: key: %v, ip: %v", key, socket.RemoteAddr().String())
		_, err = client.SRem(key, ip).Result()
	}
	if err != nil {
		return err
	}
	logrus.Infof("redis client set, key: %v, value: %v", key, client.SMembers(key))
	client.Close()
	return nil
}