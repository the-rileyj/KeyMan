package gatekeeping

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Need to test the following:
// "Cf-Connecting-Ip" header is incorrect, then reject,
// however if it is correct we must check that:
//   the originating IP is in the cloudflare IP ranges
// and if the header is unset we must check:
//   the originating IP is the Domain Locked IP
func TestDomainLocking(t *testing.T) {
	tests := []struct {
		LockingDomain, RequestDomain string
		ExpectedStatusCode           int
		Headers                      map[string]string
	}{
		// Testing for incorrect connecting IP header
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "192.168.1.2",
			ExpectedStatusCode: 400,
			Headers: map[string]string{
				"Cf-Connecting-Ip": "192.168.1.2",
			},
		},
		// Testing for correct connecting IP header and a request IP which is not in the Cloudflare range
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "192.168.1.2",
			ExpectedStatusCode: 400,
			Headers: map[string]string{
				"Cf-Connecting-Ip": "192.168.1.1",
			},
		},
		// Testing for correct connecting IP header and a request IP which is in the Cloudflare range
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "103.21.244.1",
			ExpectedStatusCode: 200,
			Headers: map[string]string{
				"Cf-Connecting-Ip": "192.168.1.1",
			},
		},
		// Testing for no connecting IP header and a request IP which is not the domain locked IP
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "192.168.1.2",
			ExpectedStatusCode: 400,
			Headers:            map[string]string{},
		},
		// Testing for no connecting IP header and a request IP which is the domain locked IP
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "192.168.1.1",
			ExpectedStatusCode: 200,
			Headers:            map[string]string{},
		},
		// Testing for no connecting IP header and a request IP which is the domain locked IP with an attached port
		{
			LockingDomain:      "192.168.1.1",
			RequestDomain:      "192.168.1.1:9900",
			ExpectedStatusCode: 200,
			Headers:            map[string]string{},
		},
	}

	testHandlerFunc := func(c *gin.Context) {
		c.String(200, "OK")
	}

	for _, testItem := range tests {
		desc := fmt.Sprintf("router(makeDomainLock(%s))", testItem.LockingDomain)

		router := gin.New()

		router.Use(MakeDomainLock(testItem.LockingDomain))

		router.GET("/", testHandlerFunc)

		mockRequest, err := http.NewRequest("GET", "/", &bytes.Reader{})

		if err != nil {
			t.Fatal("could not create the mock request")
		}

		mockRequest.RemoteAddr = testItem.RequestDomain

		for headerKey, headerValue := range testItem.Headers {
			mockRequest.Header.Set(headerKey, headerValue)
		}

		mockResponseWriter := httptest.NewRecorder()

		router.ServeHTTP(mockResponseWriter, mockRequest)

		if mockResponseWriter.Code != testItem.ExpectedStatusCode {
			t.Errorf(
				`%s = HTTP/%d, HTTP/%d was expected for IP "%s" and Cf-Connecting-Ip "%s"`,
				desc,
				mockResponseWriter.Code,
				testItem.ExpectedStatusCode,
				testItem.RequestDomain,
				mockRequest.Header.Get("Cf-Connecting-Ip"),
			)
		}
	}
}
