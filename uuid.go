package config

import (
	"code.google.com/p/go-uuid/uuid"
)

type iUUIDTokenGen int

func (_ iUUIDTokenGen) NewToken() string {
	return uuid.NewRandom().String()
}

func UUIDTokenGen() TokenGenerator {
	var i iUUIDTokenGen
	return i
}
