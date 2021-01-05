package bootstarp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jeevi-cao/lego/components/config"
	"github.com/jeevi-cao/lego/components/crontab"
	"github.com/jeevi-cao/lego/components/httpserver"
	"github.com/jeevi-cao/lego/components/httpserver/middleware"
	"github.com/jeevi-cao/lego/components/log"
	"github.com/jeevi-cao/lego/components/mongo"
	sig "github.com/jeevi-cao/lego/components/signal"
	"github.com/jeevi-cao/lego/components/zookeeper"
	"github.com/jeevi-cao/lego/pkg/app"
)

var initFunc = []func(){
	InitConfig,
	InitLog,
	InitApp,
	InitPid,
	InitCrontab,
	InitHttpServer,
	InitMongo,
	InitZookeeper,
}

func Init() error {
	t1 := time.Now()
	for _, f := range initFunc {
		f()
	}

	//注册信号函数
	sig.WatchSignal(Shutdown, nil, nil)

	cost := time.Since(t1)
	app.App.GetLogger("").Info("app init complete! time timeline:", cost)
	return nil
}

func RegisterInit(f func()) {
	initFunc = append(initFunc, f)
}

//注册route
func RegisterHttpRoutes(f func(engine *gin.Engine)) error {
	hs, _ := app.App.GetHttpServer()
	if hs == nil {
		return errors.New("http server not init")
	}
	f(hs.Engine)
	return nil
}

//注册定时任务
func RegisterCrontabTask(callbacks ...func(scheduler crontab.Scheduler)) error {
	cron, _ := app.App.GetCrontab()
	if cron == nil {
		return errors.New("crontab not init")
	}
	cron.AddTaskFunc(callbacks...)
	return nil
}

//注册信号监听函数
func RegisterSignalFunc(f sig.CallbackSignal) {
	sig.AddWatchFunc(f)
}

//初始化配置
func InitConfig() {
	cfg, err := app.App.GetCfgFile()
	if err != nil {
		panic(err.Error())
	}

	c, err := config.NewConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("[init] config error:%s", err.Error()))
	}
	//这是自动热加载文件
	c.WatchReConfig()
	app.App.SetConfig(c)
}

//初始化日志 -- 核心加载
//TODO 是否可以改成懒加载
func InitLog() {
	cfg := app.App.GetConfiger()
	//多实例
	if cfg.IsSet("log.type") && app.IsMultiInstance(cfg.GetString("log.type")) {
		instances := cfg.GetStringMap("log.instance")
		for instance := range instances {
			prefix := "log.instance." + instance + "."
			setting := log.Setting{
				Path:            cfg.GetString(prefix + "path"),
				FileName:        cfg.GetString(prefix + "filename"),
				ErrFileName:     cfg.GetString(prefix + "errfilename"),
				Level:           cfg.GetString(prefix + "level"),
				Format:          cfg.GetString(prefix + "format"),
				Split:           cfg.GetString(prefix + "split"),
				LifeTime:        cfg.GetDuration(prefix + "lifetime"),
				Rotation:        cfg.GetDuration(prefix + "rotation"),
				ReportCaller:    true,
				ReportHostIp:    true,
				ReportShortFile: true,
			}
			l, err := log.NewLog(setting)
			if err != nil {
				panic(fmt.Sprintf("[init] log instance: %s error:%s", instance, err.Error()))
			}
			app.App.SetLog(instance, l)
		}
	} else {
		setting := log.Setting{
			Path:            cfg.GetString("log.path"),
			FileName:        cfg.GetString("log.filename"),
			ErrFileName:     cfg.GetString("log.errfilename"),
			Level:           cfg.GetString("log.level"),
			Format:          cfg.GetString("log.format"),
			Split:           cfg.GetString("log.split"),
			LifeTime:        cfg.GetDuration("log.lifetime"),
			Rotation:        cfg.GetDuration("log.rotation"),
			ReportCaller:    true,
			ReportHostIp:    true,
			ReportShortFile: true,
		}
		l, err := log.NewLog(setting)
		if err != nil {
			panic(fmt.Sprintf("[init] log error:%s", err.Error()))
		}
		app.App.SetLog("", l)
	}
	app.App.GetLogger("").Info("[init] log component complete !")
}

//初始化app
func InitApp() {
	cfg := app.App.GetConfiger()
	name := cfg.GetString("app.name")
	app.App.SetName(name)

	if cfg.IsSet("app.request_id") {
		app.App.SetRequestId(cfg.GetString("app.request_id"))
	}

	app.App.GetLogger("").Info("[init] app component complete !")
}

//pid设置
func InitPid() {
	pid := os.Getpid()
	pidfile := app.App.GetConfiger().GetString("app.pidfile")
	if len(pidfile) < 1 {
		app.App.GetLogger("").Infof("[init] not need init pid file")
		return
	}
	//判断当前pid 是否存储
	file, err := os.OpenFile(pidfile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		app.App.GetLogger("").Warnf("[init] create pid file error:%s", err.Error())
		return
	}
	_, _ = file.WriteString(strconv.Itoa(pid))
	_ = file.Close()

	app.App.GetLogger("").Infof("[init] create pid file pid:%d", pid)
}

//定时任务初始化
func InitCrontab() {
	cfg := app.App.GetConfiger()
	enable := cfg.GetBool("crontab.enable")
	if enable {
		app.App.SetCrontab(crontab.New())
		app.App.GetLogger("").Infof("[init] crontab component complete!")
	}
}

//初始化server
func InitHttpServer() {
	cfg := app.App.GetConfiger()
	if !cfg.IsSet("httpserver.http_host") {
		return
	}
	host := cfg.GetString("httpserver.http_host")
	port := cfg.GetInt("httpserver.http_port")
	isHttps := cfg.GetBool("httpserver.enable_https")
	middlewares := cfg.GetStringSlice("httpserver.middleware")

	//日志输出, 测试环境 双写
	l, _ := app.App.GetLog("")
	outWriter := l.Writer
	if app.App.IsDevelop() {
		outWriter = io.MultiWriter(os.Stdout, outWriter)
	}
	//改写gin日志数据地址
	gin.DefaultErrorWriter = outWriter
	gin.DefaultWriter = outWriter

	hs := httpserver.NewHttpServer(host, port, isHttps)

	//非测试环境 打开
	if !app.App.IsDevelop() {
		hs.SetServerModeRelease()
	}

	//TODO 这段代码逻辑不太好
	if len(middlewares) > 0 {
		for _, mw := range middlewares {
			switch mw {
			case "cors":
				hs.SetMiddleware(middleware.CorsMiddleWare())
			case "requestid":
				hs.SetMiddleware(middleware.RequestIdMiddleware(app.App.GetRequestId()))
			case "ydlogger":
				hs.SetMiddleware(middleware.YdLoggerMiddleWare(outWriter))
			}
		}
	}
	app.App.SetHttpServer(hs)
	app.App.GetLogger("").Info("[init] http server complete!")
}

//初始化mongo
func InitMongo() {
	cfg := app.App.GetConfiger()

	//判断是否有配置
	if !cfg.IsSet("mongo") {
		return
	}
	var instances map[string]interface{}
	var prefix string
	var multi bool
	//判断是否多实例
	if cfg.IsSet("mongo.type") && app.IsMultiInstance(cfg.GetString("mongo.type")) {
		instances = cfg.GetStringMap("mongo.instance")
		prefix = "mongo.instance."
		multi = true
	} else {
		instances = map[string]interface{}{"mongo": ""}
		prefix = ""
		multi = false
	}
	for instance := range instances {
		pre := prefix + instance + "."
		setting := &mongo.Setting{}
		if cfg.IsSet(pre + "uri") {
			setting.Uri = cfg.GetString(pre + "uri")
		}
		if cfg.IsSet(pre + "hosts") {
			setting.Hosts = cfg.GetString(pre + "hosts")
		}
		if cfg.IsSet(pre + "replset") {
			setting.ReplSet = cfg.GetString(pre + "replset")
		}
		if cfg.IsSet(prefix + "username") {
			setting.Username = cfg.GetString(pre + "username")
		}
		if cfg.IsSet(prefix + "password") {
			setting.Password = cfg.GetString(pre + "password")
		}
		if cfg.IsSet(prefix + "max_pool_size") {
			setting.MaxPoolSize = cfg.GetUint64(pre + "max_pool_size")
		}
		if cfg.IsSet(prefix + "min_pool_size") {
			setting.MinPoolSize = cfg.GetUint64(pre + "min_pool_size")
		}
		if cfg.IsSet(prefix + "max_idle_time") {
			setting.MaxIdleTime = cfg.GetInt(prefix + "max_idle_time")
		}
		if cfg.IsSet(prefix + "read_preference") {
			setting.ReadPreference = cfg.GetString(pre + "read_preference")
		}
		mg, err := mongo.NewMongo(setting)
		if err != nil {
			app.App.GetLogger("").Fatalf("[init] mongo instance:%s  error:%s", instance, err.Error())
			continue
		}
		if !multi {
			instance = ""
		}
		app.App.SetMongo(instance, mg)
		app.App.GetLogger("").Infof("[init] mongo instance:%s set !", instance)
	}
	app.App.GetLogger("").Info("[init] mongo component complete !")
}

func InitZookeeper() {
	cfg := app.App.GetConfiger()
	if !cfg.IsSet("zookeeper") {
		return
	}
	//判断是否多实例
	var instances map[string]interface{}
	var prefix string
	var multi bool

	if cfg.IsSet("zookeeper.type") && app.IsMultiInstance(cfg.GetString("zookeeper.type")) {
		instances = cfg.GetStringMap("zookeeper.instance")
		prefix = "zookeeper.instance."
		multi = true
	} else {
		instances = map[string]interface{}{"zookeeper": ""}
		prefix = ""
		multi = false
	}

	for instance := range instances {
		pre := prefix + instance + "."

		hosts := cfg.GetStringSlice(pre + "hosts")
		t := cfg.GetInt(pre + "session_timeout")
		sessionTimeout := 5 * time.Second
		if t > 0 {
			sessionTimeout = time.Duration(t) * time.Second
		}
		zb, err := zookeeper.NewZkBuilder(hosts, sessionTimeout)
		if err != nil {
			app.App.GetLogger("").Fatalf("[init] zookeeper instance:%s error:%s", instance, err.Error())
			continue
		}
		if multi == false {
			instance = ""
		}
		app.App.SetZookeeper(instance, zb)
		app.App.GetLogger("").Info("[init] mongo instance:%s set !", instance)
	}
	app.App.GetLogger("").Info("[init] zookeeper component complete !")
}
