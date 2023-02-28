package server

import (
	"fmt"
	"medialpha-backend/utils"
	"testing"
)

func TestAny(t *testing.T) {
	configs, err := handlerScan(utils.GetProjectRoot() + "/controllers")
	if err != nil {
		panic(err)
	}

	for _, config := range configs {
		fmt.Println("======================")
		fmt.Println("Struct: ", config.Struct)
		fmt.Println("BaseURI: ", config.BaseURI)
		fmt.Println("Handlers: ")
		for _, handler := range config.Handlers {
			fmt.Println("  ", handler)
		}
	}
}
