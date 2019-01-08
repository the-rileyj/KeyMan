package utilities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/the-rileyj/KeyMan/keymanager/keymanaging"
)

const (
	KeyManURL = "https://keys.therileyjohnson.com"
)

// GetKeyValue gets the values for the provided key given that it exists in the API
func GetKeyValue(key string) (string, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))

	defer cancel()

	return GetKeyValueWithContext(ctx, key)
}

// GetKeyValueWithContext gets the values for the provided key given that it exists in the API with a request that uses the given context
func GetKeyValueWithContext(ctx context.Context, key string) (string, error) {
	return GetKeyValueWithContextAndClient(ctx, http.DefaultClient, key)
}

// GetKeyValueWithContextAndClient gets the values for the provided key given that it exists in the API with a request that uses the given context and HTTP client
func GetKeyValueWithContextAndClient(ctx context.Context, client *http.Client, key string) (string, error) {
	keyRequest, err := http.NewRequest("GET", fmt.Sprintf("%s/key/%s", KeyManURL, key), nil)

	if err != nil {
		return "", err
	}

	keyRequest = keyRequest.WithContext(ctx)

	keyResponse, err := client.Do(keyRequest)

	if err != nil {
		return "", err
	}

	var keyResponseData keymanaging.Response

	err = json.NewDecoder(keyResponse.Body).Decode(&keyResponseData)

	if err != nil {
		return "", err
	}

	if keyResponseData.Error {
		return "", errors.New(keyResponseData.Message.(string))
	}

	return keyResponseData.Message.(string), nil
}

// GetManyKeyValues gets the values for the provided keys given that they exist in the API
func GetManyKeyValues(keys ...string) (map[string]string, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))

	defer cancel()

	return GetManyKeyValuesWithContext(ctx, keys...)
}

// GetManyKeyValuesWithContext gets the values for the provided keys given that they exist in the API with a request that uses the given context
func GetManyKeyValuesWithContext(ctx context.Context, keys ...string) (map[string]string, error) {
	return GetManyKeyValuesWithContextAndClient(ctx, http.DefaultClient, keys...)
}

// GetManyKeyValuesWithContextAndClient gets the values for the provided keys given that they exist in the API with a request that uses the given context and HTTP client
func GetManyKeyValuesWithContextAndClient(ctx context.Context, client *http.Client, keys ...string) (map[string]string, error) {
	reader, writer := io.Pipe()
	keysRequest, err := http.NewRequest("POST", fmt.Sprintf("%s/keys", KeyManURL), reader)

	if err != nil {
		return nil, err
	}

	go func() { json.NewEncoder(writer).Encode(keymanaging.RequestMany{Keys: keys}) }()

	keysRequest = keysRequest.WithContext(ctx)

	keyResponse, err := client.Do(keysRequest)

	if err != nil {
		return nil, err
	}

	var keyResponseData keymanaging.Response

	err = json.NewDecoder(keyResponse.Body).Decode(&keyResponseData)

	if err != nil {
		return nil, err
	}

	if keyResponseData.Error {
		return nil, errors.New(keyResponseData.Message.(string))
	}

	return keyResponseData.Message.(map[string]string), nil
}
