package gin

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Tokumicn/go-frame/lego-lib/logs"
)

// Recover gin panic recover插件
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if val := recover(); val != nil {
				bytes, _ := httputil.DumpRequest(c.Request, false)

				logs.Errorf("[recover from panic] time:%d err:%v request:%s stack:%s",
					time.Now(), val, string(bytes), string(debug.Stack()))

				if checkNetError(val) {
					c.Error(val.(error))
					c.Abort()
					return
				}

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func checkNetError(val interface{}) bool {
	netErr, ok := val.(*net.OpError)
	if !ok {
		return false
	}

	sysErr, ok := netErr.Err.(*os.SyscallError)
	if !ok {
		return false
	}

	if sysErr.Err != syscall.EPIPE && sysErr.Err != syscall.ECONNRESET {
		return false
	}
	return true
}
