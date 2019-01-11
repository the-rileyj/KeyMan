package utilities_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func AreStringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// Need to test the following:
// If a key does exist then the value for it is returned
// If a key does not exist then the correct error is returned
func TestGetKeyValue(t *testing.T) {
	// This is setup for the tests
	testRouter, tmpRouter := gin.New(), gin.New()
	testRouter.GET("/key/:key", keymanaging.HandleGetKey)

	setupReader, setupWriter := io.Pipe()
	mockRequest, err := http.NewRequest("POST", "/key", setupReader)

	if err != nil {
		t.Fatal("could not create the mock request for the tests")
	}

	go func() { json.NewEncoder(setupWriter).Encode(keymanaging.RequestSingle{Key: "test", Value: "test"}) }()

	tmpRouter.POST("/key", keymanaging.HandlePostKey)
	tmpRouter.ServeHTTP(httptest.NewRecorder(), mockRequest)

	serveCorrectResponseClient := &http.Client{
		Transport: roundTripRequestHandler(func(request *http.Request) *http.Response {
			responseRecorder := httptest.NewRecorder()

			testRouter.ServeHTTP(responseRecorder, request)

			return responseRecorder.Result()
		}),
	}

	tests := []struct {
		ExpectedError                                                                    bool
		TestContext                                                                      context.Context
		TestClient                                                                       *http.Client
		ExpectedErrorString, ExpectedValue, TestClientString, TestContextString, TestKey string
	}{
		{
			ExpectedError:       false,
			TestContext:         context.Background(),
			TestContextString:   "context.Background()",
			ExpectedErrorString: "",
			ExpectedValue:       "test",
			TestClient:          serveCorrectResponseClient,
			TestClientString:    "serveCorrectResponseClient",
			TestKey:             "test",
		},
		{
			ExpectedError:       true,
			TestContext:         context.Background(),
			TestContextString:   "context.Background()",
			ExpectedErrorString: keymanaging.ErrorKeyDoesNotExist,
			ExpectedValue:       "",
			TestClient:          serveCorrectResponseClient,
			TestClientString:    "serveCorrectResponseClient",
			TestKey:             "tester",
		},
	}

	for _, test := range tests {
		value, err := utilities.GetKeyValueWithContextAndClient(test.TestContext, test.TestClient, test.TestKey)

		if value != test.ExpectedValue || (test.ExpectedError && (err == nil || err.Error() != test.ExpectedErrorString)) || (!test.ExpectedError && err != nil) {
			var fmtExpectedErrString, fmtResultErrString string

			if test.ExpectedError {
				fmtExpectedErrString = fmt.Sprintf("err{%s}", test.ExpectedErrorString)
			} else {
				fmtExpectedErrString = "nil"
			}

			if err != nil {
				fmtResultErrString = fmt.Sprintf("err{%s}", err.Error())
			} else {
				fmtResultErrString = "nil"
			}

			t.Errorf(
				`utilities.GetKeyValueWithContextAndClient(%s, %s, %s) = "%s", err{%s}, expected "%s", %s`,
				test.TestContextString,
				test.TestClientString,
				test.TestKey,
				value,
				fmtResultErrString,
				test.ExpectedValue,
				fmtExpectedErrString,
			)
		}
	}

	// Perform tear down operations for the test
	mockRequest, err = http.NewRequest("DELETE", "/key/test", nil)

	if err != nil {
		t.Fatal("could not create the mock request for tearing down the test")
	}

	tmpRouter.DELETE("/key/:key", keymanaging.HandleDeleteKey)
	tmpRouter.ServeHTTP(httptest.NewRecorder(), mockRequest)
}

// Need to test the following:
// If a key does exist then the value for it is returned
// If a key does not exist then the correct error is returned
func TestGetKeyValues(t *testing.T) {
	// This is setup for the tests
	testRouter, tmpRouter := gin.New(), gin.New()

	testRouter.POST("/keys", keymanaging.HandleGetManyKeys)
	tmpRouter.POST("/key", keymanaging.HandlePostKey)

	testKeyValues := map[string]string{
		"test":       "test",
		"iexist":     "yes",
		"idontexist": "no",
	}

	for key, value := range testKeyValues {
		setupReader, setupWriter := io.Pipe()
		mockRequest, err := http.NewRequest("POST", "/key", setupReader)

		if err != nil {
			t.Fatal("could not create one of the mock requests for the tests")
		}

		go func() { json.NewEncoder(setupWriter).Encode(keymanaging.RequestSingle{Key: key, Value: value}) }()

		tmpRouter.ServeHTTP(httptest.NewRecorder(), mockRequest)
	}

	checkValidityOfResults := func(existenceSlice []bool, keySlice, expectedValueSlice []string, resultsMap map[string]string) bool {
		for index, key := range keySlice {
			value, keyExists := resultsMap[key]

			if existenceSlice[index] != keyExists {
				return false
			}

			if existenceSlice[index] {
				if value != expectedValueSlice[index] {
					return false
				}
			}
		}

		return true
	}

	serveCorrectResponseClient := &http.Client{
		Transport: roundTripRequestHandler(func(request *http.Request) *http.Response {
			responseRecorder := httptest.NewRecorder()

			testRouter.ServeHTTP(responseRecorder, request)

			return responseRecorder.Result()
		}),
	}

	tests := []struct {
		ExpectedError                                            bool
		ExpectedExists                                           []bool
		TestContext                                              context.Context
		TestClient                                               *http.Client
		ExpectedErrorString, TestClientString, TestContextString string
		ExpectedValues, TestKeys                                 []string
	}{
		{
			ExpectedError: false,
			ExpectedExists: []bool{
				true,
				true,
				true,
			},
			TestContext:         context.Background(),
			TestContextString:   "context.Background()",
			ExpectedErrorString: "",
			TestClient:          serveCorrectResponseClient,
			TestClientString:    "serveCorrectResponseClient",
			ExpectedValues: []string{
				"test",
				"yes",
				"no",
			},
			TestKeys: []string{
				"test",
				"iexist",
				"idontexist",
			},
		},
		{
			ExpectedError: false,
			ExpectedExists: []bool{
				false,
				true,
				true,
			},
			TestContext:         context.Background(),
			TestContextString:   "context.Background()",
			ExpectedErrorString: "",
			TestClient:          serveCorrectResponseClient,
			TestClientString:    "serveCorrectResponseClient",
			ExpectedValues: []string{
				"",
				"yes",
				"no",
			},
			TestKeys: []string{
				"tester",
				"iexist",
				"idontexist",
			},
		},
		{
			ExpectedError: false,
			ExpectedExists: []bool{
				false,
				false,
				false,
			},
			TestContext:         context.Background(),
			TestContextString:   "context.Background()",
			ExpectedErrorString: "",
			TestClient:          serveCorrectResponseClient,
			TestClientString:    "serveCorrectResponseClient",
			ExpectedValues:      []string{},
			TestKeys:            []string{"testing", "does_not_exist", "tester"},
		},
	}

	for _, test := range tests {
		keyValuePairs, err := utilities.GetManyKeyValuesWithContextAndClient(test.TestContext, test.TestClient, test.TestKeys...)

		if !checkValidityOfResults(test.ExpectedExists, test.TestKeys, test.ExpectedValues, keyValuePairs) || (test.ExpectedError && (err == nil || err.Error() != test.ExpectedErrorString)) || (!test.ExpectedError && err != nil) {
			var fmtExpectedErrString, fmtResultErrString, fmtTestKeys, fmtExpectedKeyPairsString, fmtResultsString string

			if test.ExpectedError {
				fmtExpectedErrString = fmt.Sprintf("err{%s}", test.ExpectedErrorString)
			} else {
				fmtExpectedErrString = "nil"
			}

			if err != nil {
				fmtResultErrString = fmt.Sprintf("err{%s}", err.Error())
			} else {
				fmtResultErrString = "nil"
			}

			fmtExpectedKeyPairsStrings := make([]string, 0)

			for index, key := range test.TestKeys {
				fmtTestKeys += fmt.Sprintf(`, "%s"`, key)

				if test.ExpectedExists[index] {
					fmtExpectedKeyPairsStrings = append(fmtExpectedKeyPairsStrings, fmt.Sprintf(`"%s": "%s"`, key, test.ExpectedValues[index]))
				}
			}

			fmtExpectedKeyPairsString = fmt.Sprintf("{ %s }", strings.Join(fmtExpectedKeyPairsStrings, ", "))

			fmtResultsStrings := make([]string, 0)

			for key, value := range keyValuePairs {
				fmtResultsStrings = append(fmtResultsStrings, fmt.Sprintf(`"%s": "%s"`, key, value))
			}

			fmtResultsString = fmt.Sprintf("{ %s }", strings.Join(fmtExpectedKeyPairsStrings, ", "))

			t.Errorf(
				`utilities.GetManyKeyValuesWithContextAndClient(%s, %s%s) = %s, %s, expected %s, %s`,
				test.TestContextString,
				test.TestClientString,
				fmtTestKeys,
				fmtResultsString,
				fmtResultErrString,
				fmtExpectedKeyPairsString,
				fmtExpectedErrString,
			)
		}
	}

	// Perform tear down operations for the test
	tmpRouter.DELETE("/key/:key", keymanaging.HandleDeleteKey)

	for key, _ := range testKeyValues {
		mockRequest, err := http.NewRequest("DELETE", fmt.Sprintf("/key/%s", key), nil)

		if err != nil {
			t.Fatal("could not create the mock request for tearing down the test")
		}

		tmpRouter.ServeHTTP(httptest.NewRecorder(), mockRequest)
	}
}
