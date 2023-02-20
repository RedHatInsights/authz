package main

import (
	"authz/host"
	"authz/host/impl"
)

func main() {
	services := host.Services{Store: impl.StubAuthzStore{}}

	web := host.NewWeb(services)
	web.Host()
}
