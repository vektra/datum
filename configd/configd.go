package main

import (
	"flag"
	"net/http"

	"github.com/vektra/services/config"
)

var fAddr = flag.String("addr", ":80", "Port to listen on")
var fDir = flag.String("dir", "config", "Config dir to use")

func main() {
	flag.Parse()

	tg := config.UUIDTokenGen()
	bs := config.NewDiskStore(*fDir)
	be := config.NewMsgpackBackend(bs)

	api := config.NewHTTPApi(tg, be)

	err := http.ListenAndServe(*fAddr, api)
	if err != nil {
		panic(err)
	}
}
