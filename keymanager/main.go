package main

import (
	"flag"
	"os"

	"github.com/the-rileyj/KeyMan/keymanager/keymanaging"
)

func main() {
	keysFilePathFlag := flag.String("keyFile", "./creds/keys.json", "File path to the json file storing the keys")

	flag.Parse()

	if _, err := os.Stat(*keysFilePathFlag); os.IsNotExist(err) {
		keysFile, err := os.Create(*keysFilePathFlag)

		if err != nil {
			panic(err)
		}

		err = keymanaging.UnloadKeyDataKeys(keysFile)

		if err != nil {
			panic(err)
		}
	} else {
		keysFile, err := os.Open(*keysFilePathFlag)

		if err != nil {
			panic(err)
		}

		err = keymanaging.LoadKeyDataKeys(keysFile)

		if err != nil {
			panic(err)
		}
	}

	router := keymanaging.NewKeyManagingRouter()

	router.Use(keymanaging.WriteToFileOnUpdate(*keysFilePathFlag))

	router.DELETE("/key/:key", keymanaging.HandleDeleteKey)
	router.GET("/key/:key", keymanaging.HandleGetKey)
	router.POST("/key", keymanaging.HandlePostKey)
	router.PUT("/key/:key", keymanaging.HandlePutKey)
	router.Any("/", keymanaging.CreateInfoHandler(router))

	router.Run(":9902")
}
