package main

import (
	"log"
	"time"

	"github.com/keisuke713/fcheck"
)

func main() {
	log.Println("start")
	// server
	fcheck.Start(fcheck.StartStruct{
		AckLocalIPAckLocalPort: "localhost:12345",
	})
	// client
	ch, err := fcheck.Start(fcheck.StartStruct{
		AckLocalIPAckLocalPort:       "localhost:12347",
		EpochNonce:                   1,
		HBeatLocalIPHBeatLocalPort:   "localhost:12346",
		HBeatRemoteIPHBeatRemotePort: "localhost:12345",
		LostMsgThresh:                3,
	})
	if err != nil {
		log.Fatalf("Failed to start client fchecker: %v", err)
	}

	select {
	case <-ch:
		log.Println("Failure detected!")
		fcheck.Stop()
	case <-time.After(120 * time.Second):
		log.Println("No failure detected within 20 seconds.")
	}
	time.Sleep(5 * time.Second)
}
