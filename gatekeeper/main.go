package main

import (
	"flag"
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/the-rileyj/KeyMan/gatekeeper/gatekeeping"
)

func main() {
	debuggingFlag := flag.Bool("debug", false, "Use the HTTP router")
	forwardingFlag := flag.String("forward", "http://keymanager", "Set the host that the Gate Keeper will forward successful requests to")
	lockingFlag := flag.String("lock", "104.196.23.77", "Set the IP that will be able to access the key manager")

	flag.Parse()

	KeyManHost, err := url.ParseRequestURI(*forwardingFlag)

	if err != nil {
		panic(err)
	}

	KeyManReverseProxy := httputil.NewSingleHostReverseProxy(KeyManHost)

	router := gin.Default()

	router.Use(gatekeeping.MakeDomainLock(*lockingFlag))

	router.NoRoute(func(c *gin.Context) {
		KeyManReverseProxy.ServeHTTP(c.Writer, c.Request)
	})

	if *debuggingFlag {
		err = router.Run(":9901")
	} else {
		err = router.RunTLS(":9901", "./RJcert.crt", "./RJsecret.key")
	}

	if err != nil {
		fmt.Println(err)
	}
}
