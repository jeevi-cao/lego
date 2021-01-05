package app

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"lego/components/config"
	"lego/components/crontab"
	"lego/components/httpserver"
	"lego/components/log"
	"lego/components/mongo"
	"lego/components/zookeeper"
)

const (
	_ uint8 = iota
	DEVELOP
	TEST
	RELEASE
	PROD
)

//多实例默认标志
var defaultInstance = "app"
var multiInstanceSign = "multi"

//环境变量
var envName2Num = map[string]uint8{
	"develop": DEVELOP,
	"test":    TEST,
	"release": RELEASE,
	"prod":    PROD,
}

var envNum2Name = map[uint8]string{
	DEVELOP: "develop",
	TEST:    "test",
	RELEASE: "release",
	PROD:    "prod",
}

func EnvName2Num(name string) (uint8, error) {
	if num, ok := envName2Num[name]; ok {
		return num, nil
	}
	return 0, errors.New(fmt.Sprintf("name:%s not found num", name))
}

func EnvNum2Name(num uint8) (string, error) {
	if str, ok := envNum2Name[num]; ok {
		return str, nil
	}
	return "", errors.New(fmt.Sprintf("num:%d not found name", num))
}

func IsMultiInstance(sign string) bool {
	return sign == multiInstanceSign
}

//设置到App
var App *Application
var Once sync.Once

type Application struct {
	Name string
	//环境参数
	Env uint8
	//配置路径
	CfgFile string
	//
	RequestId string
	//组件配置
	Components *Components

	mutex *sync.Mutex
}

//支持的组件
type Components struct {
	//配置 - 核心级别
	config struct {
		handler *config.Config
		enable  bool
	}

	//日志组件 - 核心级别
	log struct {
		handler map[string]*log.Log
		enable  bool
	}

	//定时任务组件
	crontab struct {
		handler *crontab.Crontab
		enable  bool
	}
	//http server
	httpserver struct {
		handler *httpserver.HttpServer
		enable  bool
	}
	//mongo
	mongo struct {
		handler map[string]*mongo.Mongo
		enable  bool
	}
	//zookeeper
	zookeeper struct {
		handler map[string]*zookeeper.ZkBuilder
		enable  bool
	}
}

func init() {
	Once.Do(func() {
		App = &Application{
			Components: &Components{},
			mutex:      new(sync.Mutex),
		}
	})
}

func (a *Application) SetName(name string) {
	a.Name = name
}

func (a *Application) SetEnv(env string) error {

	if n, err := EnvName2Num(env); err == nil {
		a.Env = n
		return nil
	}
	return errors.New(fmt.Sprintf("application set env not exists env:%s", env))
}

func (a *Application) GetEnv() uint8 {
	return a.Env
}

func (a *Application) GetEnvName() string {
	str, _ := EnvNum2Name(a.Env)
	return str
}

//是否是开发环境
func (a *Application) IsDevelop() bool {
	return a.GetEnv() == DEVELOP
}

//是否是测试环境
func (a *Application) IsTest() bool {
	return a.GetEnv() == TEST
}

func (a *Application) IsRelease() bool {
	return a.GetEnv() == RELEASE
}

func (a *Application) IsProd() bool {
	return a.GetEnv() == PROD
}

//配置文件
func (a *Application) SetCfgFile(cfgFile string) {
	a.CfgFile = cfgFile
}

//配置文件获取
func (a *Application) GetCfgFile() (string, error) {
	if len(a.CfgFile) < 1 {
		return "", errors.New("no config file")
	}
	return a.CfgFile, nil
}

func (a *Application) SetRequestId(reqid string) {
	a.RequestId = reqid
}

func (a *Application) GetRequestId() string {
	return a.RequestId
}

//config
func (a *Application) SetConfig(cf *config.Config) {
	a.Components.config = struct {
		handler *config.Config
		enable  bool
	}{handler: cf, enable: true}
}

func (a *Application) GetConfig() (cf *config.Config, err error) {
	if a.Components.config.enable == false {
		return nil, errors.New("not init config")
	}
	return a.Components.config.handler, nil
}

func (a *Application) GetConfiger() *viper.Viper {
	cfg, _ := a.GetConfig()
	return cfg.Handler
}

//log 支持多实例
func (a *Application) SetLog(instance string, lg *log.Log) {
	defer a.mutex.Unlock()
	a.mutex.Lock()

	if a.Components.log.enable == false {
		a.Components.log = struct {
			handler map[string]*log.Log
			enable  bool
		}{handler: make(map[string]*log.Log), enable: false}
	}

	if instance == "" {
		instance = defaultInstance
	}

	a.Components.log.handler[instance] = lg
	a.Components.log.enable = true
}

func (a *Application) GetLog(instance string) (*log.Log, error) {
	if a.Components.log.enable == false {
		return nil, errors.New("not init log")
	}
	if instance == "" {
		instance = defaultInstance
	}
	hd, ok := a.Components.log.handler[instance]
	if !ok {
		return nil, errors.New("log not exists")
	}
	return hd, nil
}

func (a *Application) GetLogger(instance string) *logrus.Logger {
	l, _ := a.GetLog(instance)
	return l.Logger
}

//crontab
func (a *Application) SetCrontab(cron *crontab.Crontab) {
	a.Components.crontab = struct {
		handler *crontab.Crontab
		enable  bool
	}{handler: cron, enable: true}
}

func (a *Application) GetCrontab() (*crontab.Crontab, error) {
	if a.Components.crontab.enable == false {
		return nil, errors.New("not init crontab")
	}
	return a.Components.crontab.handler, nil
}

//httpserver
func (a *Application) SetHttpServer(hs *httpserver.HttpServer) {
	a.Components.httpserver = struct {
		handler *httpserver.HttpServer
		enable  bool
	}{handler: hs, enable: true}
}

func (a *Application) GetHttpServer() (*httpserver.HttpServer, error) {
	if a.Components.httpserver.enable == false {
		return nil, errors.New("not init httpserver")
	}
	return a.Components.httpserver.handler, nil
}

//mongo 支持多实例
func (a *Application) SetMongo(instance string, mg *mongo.Mongo) {
	defer a.mutex.Unlock()
	a.mutex.Lock()

	if a.Components.mongo.enable == false {
		a.Components.mongo = struct {
			handler map[string]*mongo.Mongo
			enable  bool
		}{handler: make(map[string]*mongo.Mongo), enable: false}
	}
	if instance == "" {
		instance = defaultInstance
	}
	a.Components.mongo.handler[instance] = mg
	a.Components.mongo.enable = true
}

func (a *Application) GetMongo(instance string) (*mongo.Mongo, error) {
	if a.Components.mongo.enable == false {
		return nil, errors.New("not init mongo")
	}
	if instance == "" {
		instance = defaultInstance
	}
	mg, ok := a.Components.mongo.handler[instance]
	if !ok {
		return nil, errors.New("mongo not exists")
	}
	return mg, nil
}

func (a *Application) GetAllMongo() (map[string]*mongo.Mongo, error) {
	if a.Components.mongo.enable == false {
		return nil, errors.New("not init mongo")
	}
	return a.Components.mongo.handler, nil
}

//zookeeper
func (a *Application) SetZookeeper(instance string, zk *zookeeper.ZkBuilder) {
	defer a.mutex.Unlock()
	a.mutex.Lock()

	if a.Components.zookeeper.enable == false {
		a.Components.zookeeper = struct {
			handler map[string]*zookeeper.ZkBuilder
			enable  bool
		}{handler: make(map[string]*zookeeper.ZkBuilder), enable: false}
	}

	if instance == "" {
		instance = defaultInstance
	}

	a.Components.zookeeper.handler[instance] = zk
	a.Components.zookeeper.enable = true
}

func (a *Application) GetZookeeper(instance string) (*zookeeper.ZkBuilder, error) {
	if a.Components.zookeeper.enable == false {
		return nil, errors.New("not init zookeeper")
	}
	if instance == "" {
		instance = defaultInstance
	}
	zk, ok := a.Components.zookeeper.handler[instance]
	if !ok {
		return nil, errors.New("mongo not exists")
	}
	return zk, nil
}

func (a *Application) GetAllZookeeper() (map[string]*zookeeper.ZkBuilder, error) {
	if a.Components.zookeeper.enable == false {
		return nil, errors.New("not init zookeeper")
	}
	return a.Components.zookeeper.handler, nil
}

func (a *Application) Close() {
	a.Components = &Components{}
}
