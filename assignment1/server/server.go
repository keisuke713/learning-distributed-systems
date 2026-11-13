package main

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"net"
	"os"
	"time"

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
		log.Printf("recieved state: %v \n", state)
		// decodeしたステートが入っているのでリセットする
		buf.Reset()

		// game start
		if state.GameState == nil {
			state.GameState = []uint8{4, 4, 4}
		} else if state.HasFinish() {
			log.Print("lose")
			os.Exit(0)
		} else {
			state.MoveCoin()
		}

		if n := rand.Intn(5); n == 4 && state.MoveRow != -1 {
			// 1/5の確率で値を不正に操作
			log.Printf("modify game state: from %v\n", state)
			state.GameState[0] = 10
		} else if n < 1 && state.MoveRow != -1 {
			// 1/5の確率で返答しない
			log.Println("pause")
			time.Sleep(2 * time.Second)
		}
		log.Printf("send state: %+v\n", state)
		if err := encoder.Encode(state); err != nil {
			log.Fatalf(err, "failed to encode state")
		}
		ln.WriteToUDP(buf.Bytes(), clientAddr)
		if err != nil {
			log.Fatalf(err, "failed to write data")
		}
	}
}
