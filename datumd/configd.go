package main

import (
	"flag"
	"net/http"

	"github.com/vektra/datum"
)

var fAddr = flag.String("addr", ":80", "Port to listen on")
var fDir = flag.String("dir", "config", "Config dir to use")

func main() {
	flag.Parse()

	tg := datum.UUIDTokenGen()
	bs := datum.NewDiskStore(*fDir)
	be := datum.NewMsgpackBackend(bs)

	api := datum.NewHTTPApi(tg, be)

	err := http.ListenAndServe(*fAddr, api)
	if err != nil {
		panic(err)
	}
}
