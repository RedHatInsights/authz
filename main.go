package main

import (
	"authz/host"
	"authz/host/impl"
	"sync"
)

func main() {
	services := host.Services{Store: impl.StubAuthzStore{Data: map[string]bool{
		"token": true,
		"alice": true,
		"bob":   true,
		"chuck": false,
	}}}

	wait := sync.WaitGroup{}
	web := host.NewWeb(services)

	wait.Add(1)
	go web.Host(&wait)

	wait.Wait()
}
