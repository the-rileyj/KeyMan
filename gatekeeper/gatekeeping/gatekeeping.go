package gatekeeping

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	cloudflareIPV4s = "https://www.cloudflare.com/ips-v4"
	cloudflareIPV6s = "https://www.cloudflare.com/ips-v6"
)

var rangesOfCloudFlareIPs []*net.IPNet

func getRangesOfCloudFlareIPs() ([]*net.IPNet, error) {
	parsedRanges := make([]*net.IPNet, 0)

	for _, listOfRanges := range []string{cloudflareIPV4s, cloudflareIPV6s} {
		response, err := http.Get(listOfRanges)

		if err != nil {
			return []*net.IPNet{}, err
		}

		bodyBytes, err := ioutil.ReadAll(response.Body)

		response.Body.Close()

		if err != nil {
			return []*net.IPNet{}, err
		}

		responseBody := string(bodyBytes)

		for _, ipRangeString := range strings.Split(responseBody, "\n") {
			if ipRangeString != "" {
				_, ipRange, err := net.ParseCIDR(ipRangeString)

				if err != nil {
					return []*net.IPNet{}, err
				}

				parsedRanges = append(parsedRanges, ipRange)
			}
		}
	}

	return parsedRanges, nil
}

func init() {
	var err error

	rangesOfCloudFlareIPs, err = getRangesOfCloudFlareIPs()

	if err != nil {
		panic(err)
	}
}

func cloudflareRangeHasIP(IP string) bool {
	parsedIP := net.ParseIP(IP)

	if parsedIP == nil {

	}

	for _, IPNet := range rangesOfCloudFlareIPs {
		if IPNet.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func parseRemoteAddr(request *http.Request) string {
	IPs := strings.Split(request.RemoteAddr, ", ")

	if len(IPs) == 0 {
		return ""
	}

	ip, _, err := net.SplitHostPort(IPs[0])

	// In the case there is no port on the remote addr, ex. ::1
	if err != nil {
		return IPs[0]
	}

	// In the case there was a port on the remote addr, ex. [::1]:9001
	return ip
}

func MakeDomainLock(lockIP string) func(*gin.Context) {
	return func(c *gin.Context) {
		if (lockIP == c.Request.Header.Get("Cf-Connecting-Ip") && cloudflareRangeHasIP(parseRemoteAddr(c.Request))) || (c.Request.Header.Get("Cf-Connecting-Ip") == "" && parseRemoteAddr(c.Request) == lockIP) {
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
