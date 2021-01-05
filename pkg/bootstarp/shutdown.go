package bootstarp

import (
	"time"

	"github.com/jeevi-cao/lego/pkg/app"
)

var shutdownFunc = []func(){
	ShutdownHttpServer,
	ShutdownCrontab,
	ShutdownMongo,
	ShutdownZookeeper,
	ShutdownApp,
}

func Shutdown() {
	t1 := time.Now()
	for _, f := range shutdownFunc {
		f()
	}
	cost := time.Since(t1)
	app.App.GetLogger("").Info("[shutdown] app shutdown complete! time timeline:", cost)

}

func RegisterShutdown(f func()) {
	shutdownFunc = append(shutdownFunc, f)
}

func ShutdownCrontab() {
	cron, _ := app.App.GetCrontab()
	if cron != nil {
		cron.Clear()
		cron.Stop()
		app.App.GetLogger("").Info("[shutdown] shutdown crontab complete!")
	}
}

func ShutdownMongo() {
	mongos, _ := app.App.GetAllMongo()
	if mongos == nil {
		return
	}
	for instance, m := range mongos {
		m.Close()
		app.App.GetLogger("").Infof("[shutdown] shutdown mongo instance:%s complete!", instance)
	}
}

func ShutdownZookeeper() {
	zookeepers, _ := app.App.GetAllZookeeper()
	if zookeepers == nil {
		return
	}
	for instance, z := range zookeepers {
		z.Stop()
		app.App.GetLogger("").Infof("[shutdown] shutdown zookeeper instance:%s complete!", instance)
	}
}

func ShutdownHttpServer() {
	hs, _ := app.App.GetHttpServer()
	if hs != nil {
		hs.GracefulShutdown()
	}
}

func ShutdownApp() {
	app.App.Close()
}
