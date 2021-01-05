package pprof

import (
	"github.com/DeanThompson/ginpprof"

	"github.com/jeevi-cao/lego/components/httpserver"
)

func UseHttpPprof(server *httpserver.HttpServer) {
	//http server 设置
	ginpprof.Wrap(server.Engine)
}
