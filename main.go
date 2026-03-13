//go:build linux

package main

import (
	"flag"
	"log"

	"github.com/ronijaat/redisDB/config"
	"github.com/ronijaat/redisDB/server"
)

func setupFlags() {
	flag.StringVar(&config.HOST, "host", "0.0.0.0", "host for the redis server")
	flag.IntVar(&config.PORT, "port", 7379, "port for redis db server")
	flag.Parse()
}
func main() {
	setupFlags()
	log.Println("rolling the redis")
	server.RunAsyncTCPServer()
}
