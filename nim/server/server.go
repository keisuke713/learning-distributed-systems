package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"time"

	fcheck "learning-distributed-systems/fchecker"
	"learning-distributed-systems/nim/common"

	"github.com/DistributedClocks/tracing"
)

func Listen(done chan int) {
	config := common.ParseConfig("./config/config.json")
	for i := 0; i < len(config.NimServerAddressList); i++ {
		go listen(done, i, config.NimServerAddressList[i], config.FCheckHbeatRemoteAddrList[i], config.TracingServerAddress, config.Secret)
	}
}

func listen(done chan int, n int, nimServerAddr string, ackLocalIPAckLocalPort string, tracingServerAddress string, secret []byte) {
	tracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  tracingServerAddress,
		TracerIdentity: fmt.Sprintf("server_%d", n),
		Secret:         secret,
	})
	defer tracer.Close()

	log := common.NewLog(fmt.Sprintf("server_%d", n))

	fcheck.Start(fcheck.StartStruct{
		AckLocalIPAckLocalPort: ackLocalIPAckLocalPort,
	})

	serverAddr, err := net.ResolveUDPAddr("udp", nimServerAddr)
	if err != nil {
		log.Fatalf(err, "failed to resolve address")
	}
	ln, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalf(err, "failed to start server")
	}
	log.Printf("server_%d start\n", n)

	var (
		trace      *tracing.Trace
		isFirstMsg bool = true
	)
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
			log.Fatalf(err, "failed to decode state")
		}
		log.Printf("recieved state: %v \n", state)
		trace = tracer.ReceiveToken(state.Token)
		if isFirstMsg {
			isFirstMsg = false
			if len(state.GameState) == 0 {
				trace.RecordAction(common.ServerGameStart{NimServerAddress: serverAddr.String()})
			} else {
				// 2台目以降のサーバー
				trace.RecordAction(common.GameResume{NimServerAddress: serverAddr.String()})
			}
		} else {
			trace.RecordAction(
				common.ClietnMoveReceive(state),
			)
		}
		// decodeしたステートが入っているのでリセットする
		buf.Reset()

		// game start
		if state.GameState == nil {
			state.GameState = []uint8{4, 4, 4}
		} else if state.HasFinish() {
			trace.RecordAction(common.GameComplete{Winner: "client"})
			done <- 1
			return
		} else {
			state.MoveCoin()
		}

		if n := rand.Intn(5); n == 4 && state.MoveRow != -1 {
			// 1/5の確率で値を不正に操作
			log.Printf("modify game state: from %v\n", state)
			state.GameState[0] = 10
		} else if n < 4 && state.MoveRow != -1 {
			// 4/5の確率で返答しない
			log.Println("pause")
			time.Sleep(30 * time.Second)
		}
		log.Printf("send state: %+v\n", state)
		if err := encoder.Encode(state); err != nil {
			log.Fatalf(err, "failed to encode state")
		}
		ln.WriteToUDP(buf.Bytes(), clientAddr)
		if err != nil {
			log.Fatalf(err, "failed to write data")
		}
		trace.RecordAction(
			common.ServerMove(state),
		)
	}
}
