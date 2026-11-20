package main

import (
	"github.com/keisuke713/cs416/nim/client"
	"github.com/keisuke713/cs416/nim/common"
	"github.com/keisuke713/cs416/nim/server"

	"github.com/DistributedClocks/tracing"
)

var (
	log = common.NewLog("main")
)

func main() {
	tracingServer := tracing.NewTracingServerFromFile("./config/tracing_server_config.json")
	err := tracingServer.Open()
	if err != nil {
		log.Fatalf(err, "failed to open")
	}
	defer tracingServer.Close()
	go tracingServer.Accept()

	done := make(chan int)
	go server.Listen(done)
	go client.Dial(done)

	<-done
	log.Println("game finish")
}
