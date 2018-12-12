package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
)

func getIP(request *http.Request) string { return request.Header.Get("Cf-Connecting-Ip") }

func makeDomainLock(lockIP string) func(*gin.Context) {
	return func(c *gin.Context) {
		bytes, _ := httputil.DumpRequest(c.Request, true)

		fmt.Println(string(bytes))
		if lockIP == getIP(c.Request) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(
			400,
			gin.H{
				"error": "sorry, you are not authorized for this information",
			},
		)
	}
}

func main() {
	httpsRouter := gin.Default()

	httpsRouter.Use(makeDomainLock("104.196.23.77"))


	httpsRouter.GET("/", func(c *gin.Context) {
		fmt.Println("wow")
		c.Writer.Write([]byte("WOW"))
	})

	httpsRouter.RunTLS(":443", "./RJcert.crt", "./RJsecret.key")
}
