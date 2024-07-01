package main

import (
	"os"
	"proxy/proxy-go/spa"
)


func main() {
	arg := os.Args[1]
	srv := spa.NewSPA()
	srv.SetEnv(arg)
	srv.Server()
}