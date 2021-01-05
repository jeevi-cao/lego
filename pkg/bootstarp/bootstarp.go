package bootstarp

import "lego/pkg/app"

var StopChan = make(chan struct{})

func Start() {
	//启动httpserver
	hs, _ := app.App.GetHttpServer()
	if hs != nil {
		hs.ServerRun()
	}
	//crontab
	cron, _ := app.App.GetCrontab()
	if cron != nil {
		cron.StartAsync()
	}
}

//关闭服务
func Stop(stop bool) {
	Shutdown()
	if stop == true {
		StopChan <- struct{}{}
	}

}

func Restart() {
	Stop(false)
	Start()
}

func Run() {
	Start()
	<-StopChan
}
