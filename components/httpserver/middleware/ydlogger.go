package middleware

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

func YdLoggerMiddleWare(output io.Writer) gin.HandlerFunc {
	logCfg := gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			format := "%s requestId=%s client-ip=%s method=%s, path=%s, proto=%s, statusCode=%d, bodySize=%d latency=%s, user-agent=%s, error-message=%s \n"
			return fmt.Sprintf(format,
				param.TimeStamp.Format("2006-01-02 15:04:05,000"),
				param.Request.Header.Get("X-Request-Id"),
				param.ClientIP,
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.BodySize,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		Output: output,
	}
	return gin.LoggerWithConfig(logCfg)
}
