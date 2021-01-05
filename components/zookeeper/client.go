package zookeeper

import (
	"time"

	"github.com/go-zookeeper/zk"
)

//ZooKeeper builder
type ZkBuilder struct {
	Conn    *zk.Conn
	Setting *Setting
}

//config
type Setting struct {
	Hosts          []string
	SessionTimeout time.Duration
}

//实例化 zookeeper
func NewZkBuilder(hosts []string, t time.Duration) (*ZkBuilder, error) {
	zc := Setting{
		Hosts:          hosts,
		SessionTimeout: t,
	}
	zb := &ZkBuilder{
		Setting: &zc,
	}
	err := zb.Start()
	return zb, err
}

//start
func (zb *ZkBuilder) Start() error {
	conn, _, err := zk.Connect(zb.Setting.Hosts, zb.Setting.SessionTimeout)
	if err != nil {
		return err
	}
	zb.Conn = conn
	return nil
}

func (zb *ZkBuilder) Restart() error {
	zb.Conn.Close()
	return zb.Start()
}

//停止
func (zb *ZkBuilder) Stop() {
	zb.Conn.Close()
	zb.Conn = nil
	zb.Setting = nil
}
