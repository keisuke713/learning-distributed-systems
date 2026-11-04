package main

import (
	"bytes"
	"encoding/gob"
	"net"

	"github.com/keisuke713/cs416/nim/common"
)

var (
	log = common.NewLog("server")
)

func main() {
	config := common.ParseConfig("../config/config.json")
	serverAddr, err := net.ResolveUDPAddr("udp", config.NimServerAddress)
	if err != nil {
		log.Fatalf(err, "failed to resolve address")
	}
	ln, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalf(err, "failed to start server")
	}
	log.Println("server start")

	for {
		var (
			state common.StateMoveMessage

			buf *bytes.Buffer = &bytes.Buffer{}
			// 作り直しをすることで内部の情報がリセットされるらしい
			encoder *gob.Encoder = gob.NewEncoder(buf)
			decoder *gob.Decoder = gob.NewDecoder(buf)
		)

		b := make([]byte, 1024)
		_, clientAddr, err := ln.ReadFromUDP(b)
		if err != nil {
			log.Fatalf(err, "failed to read data")
		}
		buf.Write(b)
		decoder = gob.NewDecoder(buf)
		if err := decoder.Decode(&state); err != nil {
			log.Fatalf(err, "failed to decode start state")
		}
		log.Printf("recieved initial state: %v \n", state)
		// decodeしたステートが入っているのでリセットする
		buf.Reset()
		state.GameState = []uint8{4, 4, 4}

		log.Printf("raw data before encoding: %s, len: %d \n", buf.String(), buf.Len())
		if err := encoder.Encode(state); err != nil {
			log.Fatalf(err, "failed to encode initial state")
		}
		log.Printf("state: %+v\n", state)
		log.Printf("raw data: %s, len: %v \n", buf.String(), buf.Len())
		ln.WriteToUDP(buf.Bytes(), clientAddr)
		if err != nil {
			log.Fatalf(err, "failed to write data")
		}
	}
}
