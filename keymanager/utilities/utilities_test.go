package utilities_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/the-rileyj/KeyMan/keymanager/keymanaging"
	"github.com/the-rileyj/KeyMan/keymanager/utilities"

	"github.com/gin-gonic/gin"
)

type roundTripRequestHandler func(*http.Request) *http.Response

func (rTRH roundTripRequestHandler) RoundTrip(request *http.Request) (*http.Response, error) {
	return rTRH(request), nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetKeyValue(t *testing.T) {
	testRouter, tmpRouter := gin.New(), gin.New()
	testRouter.GET("/key/:key", keymanaging.HandleGetKey)

	reader, writer := io.Pipe()
	mockRequest, err := http.NewRequest("POST", "/key", reader)

	if err != nil {
		t.Fatal("could not create the mock request")
	}

	go func() { json.NewEncoder(writer).Encode(keymanaging.RequestSingle{Key: "test", Value: "test"}) }()

	tmpRouter.POST("/key", keymanaging.HandlePostKey)
	tmpRouter.ServeHTTP(httptest.NewRecorder(), mockRequest)

	client := &http.Client{
		Transport: roundTripRequestHandler(func(request *http.Request) *http.Response {
			responseRecorder := httptest.NewRecorder()

			testRouter.ServeHTTP(responseRecorder, request)

			return responseRecorder.Result()
		}),
	}

	str, err := utilities.GetKeyValueWithContextAndClient(context.Background(), client, "test")

	if str != "test" {
		t.Error(err)
	}
}
