package server

import (
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cRPC-common/tcp"
	"time"
)

type serviceConn struct {
	serviceName    string
	serviceVersion string
	conn           *tcp.Connection
	exit           bool
	gid            *tcp.GID
	heartbeatTime  time.Time
	ccs            *ControlCenterServer
}

func (sc *serviceConn) loop() {

	// todo: 上线记录
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("serviceConn loop panic serviceName: %v, serviceVersion: %v, err: %v", sc.serviceName, sc.serviceVersion, err)
		}
	}()
	for !sc.exit {
		msg, err := sc.conn.Receive(0 * time.Second)
		if err != nil {
			logrus.Errorf("receive msg in loop occurred err : %v", err)
			sc.exit = true
		}
		// 刷新心跳
		sc.handleHeartbeat(msg)
	}
	logrus.Infof("complete loop gid: %s", sc.gid.String())
	// 关闭 sc
	sc.close()
	// todo: 下线记录，如何通知其他 ccs 上 <订阅了本服务的服务>，本服务下线了？
}

/**
 * @Description: 节点下线
 * @receiver sc
 */
func (sc *serviceConn) offLine(cause string) {
	logrus.Infof("service conn offLine, cause:[%v]", cause)
	sc.exit = true
	sc.conn.Close() // 触发 sc.loop 内的 for 循环执行一次
}

func (sc *serviceConn) handleHeartbeat(msg *tcp.Message) {

	sc.heartbeatTime = time.Now()
	sc.ccs.refreshHeartbeat(sc)
	// todo: 刷新心跳逻辑是否写全
}

/**
 * @Description: sc.close 与 sc.offLine 语意的区别: close 是清理 ccs 内 sc 对应的信息，OffLine 是外部 让 sc 关闭
 * @receiver sc
 */
func (sc *serviceConn) close() {

	// 删除 ccs 内的 sc 信息
	sc.ccs.lock.Lock()
	defer sc.ccs.lock.Unlock()
	curS := sc.ccs.clientMap[sc.gid.NameAndVersion()]
	if curS == nil {
		logrus.Error("not find sc from ccs.clientMap")
	} else {
		index := 0
		for i := 0; i < len(curS); i++ {
			if curS[i] == sc {
				index = i
			}
		}
		curS = append(curS[:index], curS[index+1:]...)
	}

	sc.ccs.deleteHeartbeat(sc)                                                      // 清除 time wheel
	err := sc.ccs.redisOp(sc.gid.NameAndVersion(), sc.conn.IP, redisOpSRemServerIp) // 清除 redis 信息
	if err != nil {
		// 报警，人工介入
		logrus.Errorf("delete client ip from redis error: [%v]", err)
	}
}
