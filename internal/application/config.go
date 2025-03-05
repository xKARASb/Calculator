package application

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xKARASb/Calculator/pkg/dir"
)

type config struct {
	Debug bool `json:"debug"`
	Web   bool `json:"web"`
}

func newConfig() *config {
	fmt.Println(dir.Json_file())
	res := new(config)
	cf, err := os.Open(dir.Json_file())
	if err != nil {
		panic("cannot open config file")
	}
	decoder := json.NewDecoder(cf)
	err = decoder.Decode(res)
	if err != nil {
		panic(fmt.Sprintf("cannot decode config file: %v", err))
	}
	return res
}
