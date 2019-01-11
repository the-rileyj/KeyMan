package keymanaging

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	ErrorBadRequest       string = "could not find a handler for the provided request"
	ErrorInvalidKey       string = "one or many character in the key provided make it invalid for creation; only numbers and letters are allowed"
	ErrorKeyAlreadyExists string = "the key provided for creation already exists"
	ErrorKeyDoesNotExist  string = "the key provided does not exist"
)

type keyData struct {
	keys  map[string]string
	mutex *sync.Mutex
}

var keys keyData

func (kd keyData) cloneKeys() map[string]string {
	var clone map[string]string

	kd.mutex.Lock()

	for key, value := range kd.keys {
		clone[key] = value
	}

	kd.mutex.Unlock()

	return clone
}

func (kd *keyData) delete(key string) {
	kd.mutex.Lock()

	delete(kd.keys, key)

	kd.mutex.Unlock()
}

func (kd keyData) get(key string) (string, bool) {
	kd.mutex.Lock()

	value, exists := kd.keys[key]

	kd.mutex.Unlock()

	return value, exists
}

func (kd keyData) getMany(keys ...string) map[string]string {
	keyData := make(map[string]string)

	kd.mutex.Lock()

	defer kd.mutex.Unlock()

	for _, key := range keys {
		value, exists := kd.keys[key]

		if exists {
			keyData[key] = value
		}
	}

	return keyData
}

func newKeyData() keyData {
	return keyData{
		keys:  make(map[string]string),
		mutex: &sync.Mutex{},
	}
}

func (kd keyData) set(key, value string) {
	kd.mutex.Lock()

	kd.keys[key] = value

	kd.mutex.Unlock()
}

// LoadKeyDataKeys tries to read from the provided reader into "keys.keys"
func LoadKeyDataKeys(r io.Reader) error {
	keys.mutex.Lock()

	err := json.NewDecoder(r).Decode(&keys.keys)

	keys.mutex.Unlock()

	return err
}

// UnloadKeyDataKeys tries to write "keys.keys" to the provided writer
func UnloadKeyDataKeys(w io.Writer) error {
	keys.mutex.Lock()

	err := json.NewEncoder(w).Encode(&keys.keys)

	keys.mutex.Unlock()

	return err
}

// Response represents the response which will occur from any route in the keymanaging package
type Response struct {
	Error   bool        `json:"error"`
	Message interface{} `json:"msg"`
}

// RequestSingle is the struct representing the format that POST and PUT requests will use to create and update a single key value pair
type RequestSingle struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RequestMany is the struct representing the format that requests will use to get many key value pairs
type RequestMany struct {
	Keys []string `json:"keys"`
}

func init() {
	keys = newKeyData()
}

// NewKeyManagingRouter creates a router with the recovery middleware and a no route handler attached
func NewKeyManagingRouter() *gin.Engine {
	KeyManagingRouter := gin.New()

	KeyManagingRouter.Use(gin.Recovery())

	KeyManagingRouter.NoRoute(func(c *gin.Context) {
		c.JSON(404, Response{Error: true, Message: ErrorBadRequest})
	})

	return KeyManagingRouter
}

// CreateInfoHandler creates a handler which sends back
func CreateInfoHandler(r *gin.Engine) func(c *gin.Context) {
	routesInfo := append(gin.RoutesInfo{}, r.Routes()...)

	// Sort by path length (shortest to longest) and by HTTP method (alphabetical order)
	sort.Slice(routesInfo, func(r1, r2 int) bool {
		return routesInfo[r1].Method < routesInfo[r2].Method && len(routesInfo[r1].Path) < len(routesInfo[r2].Path)
	})

	routesInfoList := make([]gin.H, 0)

	for _, routeInfo := range routesInfo {
		routesInfoList = append(routesInfoList, gin.H{
			"method": routeInfo.Method,
			"path":   routeInfo.Path,
		})
	}

	prettyPrintedRoutes, _ := json.MarshalIndent(routesInfoList, "", "  ")

	return func(c *gin.Context) {
		c.Status(200)
		c.Writer.Write(prettyPrintedRoutes)
	}
}

// HandleDeleteKey handles the DELETE request for deletion of an existing key/value pair
func HandleDeleteKey(c *gin.Context) {
	_, exists := keys.get(c.Param("key"))

	if exists {
		keys.delete(c.Param("key"))
	}

	if !exists {
		c.AbortWithStatusJSON(400, Response{true, ErrorKeyDoesNotExist})

		return
	}

	c.Writer.Header().Set("update", "update")
	c.JSON(200, Response{false, ""})
}

// HandleGetKey handles the GET request for the value of an existing key
func HandleGetKey(c *gin.Context) {
	value, exists := keys.get(c.Param("key"))

	if !exists {
		c.AbortWithStatusJSON(400, Response{true, ErrorKeyDoesNotExist})

		return
	}

	c.JSON(200, Response{false, value})
}

// HandleGetManyKeys handles a POST request for the values of many existing keys
func HandleGetManyKeys(c *gin.Context) {
	var GetManyRequest RequestMany

	err := json.NewDecoder(c.Request.Body).Decode(&GetManyRequest)

	if err != nil {
		c.AbortWithStatusJSON(400, Response{true, "could not unmarshal JSON"})

		return
	}

	c.JSON(200, gin.H{
		"err": false,
		"msg": keys.getMany(GetManyRequest.Keys...),
	})
}

// HandlePostKey handles the POST request for the creation of a key/value pair which does not already exist
func HandlePostKey(c *gin.Context) {
	var UpdateRequest RequestSingle

	err := json.NewDecoder(c.Request.Body).Decode(&UpdateRequest)

	if err != nil {
		c.AbortWithStatusJSON(400, Response{true, "could not unmarshal JSON"})

		return
	}

	encodedKey := url.QueryEscape(UpdateRequest.Key)

	if encodedKey != UpdateRequest.Key {
		c.AbortWithStatusJSON(400, Response{true, ErrorInvalidKey})

		return
	}

	_, exists := keys.get(UpdateRequest.Key)

	if !exists {
		keys.set(UpdateRequest.Key, UpdateRequest.Value)
	}

	if exists {
		c.AbortWithStatusJSON(400, Response{true, ErrorKeyAlreadyExists})

		return
	}

	c.Writer.Header().Set("update", "update")
	c.JSON(201, Response{false, ""})
}

// HandlePutKey handles the PUT request for the updating of a key/value pair which already exists
func HandlePutKey(c *gin.Context) {
	var UpdateRequest RequestSingle

	err := json.NewDecoder(c.Request.Body).Decode(&UpdateRequest)

	if err != nil {
		c.AbortWithStatusJSON(400, Response{true, "could not unmarshal JSON"})

		return
	}

	_, exists := keys.get(UpdateRequest.Key)

	if exists {
		keys.set(UpdateRequest.Key, UpdateRequest.Value)
	}

	if !exists {
		c.AbortWithStatusJSON(400, Response{true, ErrorKeyDoesNotExist})

		return
	}

	c.Writer.Header().Set("update", "update")
	c.JSON(200, Response{false, ""})
}

// WriteToFileOnUpdate is a middleware handler which writes "keys.keys" to the file path provided when the update header on the response writer is equal to "update"
func WriteToFileOnUpdate(keyDataFilePath string) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Writer.Header().Set("update", "")

		c.Next()

		if c.Writer.Header().Get("update") == "update" {
			keysFile, err := os.OpenFile(keyDataFilePath, os.O_RDWR, 0666)

			if err != nil {
				return
			}

			UnloadKeyDataKeys(keysFile)
		}
	}
}
