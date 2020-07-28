package gin

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Tokumicn/go-frame/lego-lib/logs"
)

// Logger gin log插件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Now().Sub(start)
		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logs.Error(e)
			}
		} else {
			logs.Infof("%s status:%d method:%s query:%s ip:%s user-agent:%s latency:%d",
				c.Request.URL.Path, c.Writer.Status(), c.Request.Method, c.Request.URL.RawQuery,
				c.ClientIP(), c.Request.UserAgent(), latency)
		}
	}
}

// LogPostBody gin log插件 打印POST PUT请求json参数
func LogPostBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			body, _ := ioutil.ReadAll(c.Request.Body)

			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			str := string(body[:len(body)])

			logs.Infof("%s method:%s body:%s", c.Request.URL.Path, c.Request.Method, str)
		}
	}
}
