package main

import (
	"log"
	"net"

	"github.com/keisuke713/cs416/nim/common"
)

func main() {
	config := common.ParseConfig("../config/config.json")
	udpAddr, err := net.ResolveUDPAddr("udp", config.NimServerAddress)
	if err != nil {
		common.CheckErr(err, "failed to resolve address")
	}
	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		common.CheckErr(err, "failed to start server")
	}
	log.Println("server start")

	for {
		b := make([]byte, 1024)
		n, err := ln.Read(b)
		if err != nil {
			common.CheckErr(err, "failed to read data")
		}
		if n > 0 {
			log.Printf("content i %s \n", string(b))
		}
	}
}
