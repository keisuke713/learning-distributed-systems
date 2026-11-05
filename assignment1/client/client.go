package main

import (
	"bytes"
	"encoding/gob"
	"net"
	"os"

	"github.com/DistributedClocks/tracing"
	"github.com/keisuke713/cs416/nim/common"
)

var (
	log = common.NewLog("client")
)

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Println("Usage: client.go [seed]")
	// 	return
	// }
	// arg, err := strconv.Atoi(os.Args[1])
	// CheckErr(err, "Provided seed could not be converted to integer", arg)
	// seed := int8(arg)
	seed := int8(6)

	// setup
	tracingServer := tracing.NewTracingServerFromFile("../config/tracing_server_config.json")
	err := tracingServer.Open()
	if err != nil {
		log.Fatalf(err, "failed to open")
	}
	defer tracingServer.Close()
	go tracingServer.Accept()

	config := common.ParseConfig("../config/config.json")
	tracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddress,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	defer tracer.Close()

	trace := tracer.CreateTrace()
	trace.RecordAction(
		common.GameStart{
			Seed: seed,
		})

	clientAddr, err := net.ResolveUDPAddr("udp", config.ClientAddress)
	if err != nil {
		log.Fatalf(err, "failed to resolve client address")
	}

	ln, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		log.Fatalf(err, "failed to connect to a server")
	}

	serverAddr, err := net.ResolveUDPAddr("udp", config.NimServerAddress)
	if err != nil {
		log.Fatalf(err, "failed to resolve server address")
	}
	// ゲーム開始直後
	state := common.InitialStateMoveMessage(seed)
	var (
		buf     *bytes.Buffer = &bytes.Buffer{}
		encoder               = gob.NewEncoder(buf)
		decoder               = gob.NewDecoder(buf)
	)
	if err := encoder.Encode(state); err != nil {
		log.Fatalf(err, "failed to encode state")
	}
	if _, err := ln.WriteToUDP(buf.Bytes(), serverAddr); err != nil {
		log.Fatalf(err, "failed to send data")
	}
	log.Println("send initial state")

	// encodeステート情報が入っているのでリセットする
	buf.Reset()
	b := make([]byte, 1024)
	_, serverAddr, err = ln.ReadFromUDP(b)
	if err != nil {
		log.Fatalf(err, "failed to read server data")
	}

	buf.Write(b)
	if err := decoder.Decode(&state); err != nil {
		log.Fatalf(err, "failed to decode server response")
	}
	log.Printf("recieved initial data: %v \n", state)
	for {
		if state.HasFinish() {
			log.Println("lose")
			os.Exit(0)
		}

		var (
			buf     *bytes.Buffer = &bytes.Buffer{}
			encoder               = gob.NewEncoder(buf)
			decoder               = gob.NewDecoder(buf)
		)
		state.MoveCoin()
		if err := encoder.Encode(state); err != nil {
			log.Fatalf(err, "failed to encode state")
		}
		log.Printf("send state  state: %v\n", state)
		if _, err := ln.WriteToUDP(buf.Bytes(), serverAddr); err != nil {
			log.Fatalf(err, "failed to send data")
		}

		// encodeステート情報が入っているのでリセットする
		buf.Reset()
		b := make([]byte, 1024)
		_, serverAddr, err = ln.ReadFromUDP(b)
		if err != nil {
			log.Fatalf(err, "failed to read server data")
		}

		buf.Write(b)
		if err := decoder.Decode(&state); err != nil {
			log.Fatalf(err, "failed to decode server response")
		}
		log.Printf("recieved data: %v \n", state)
	}
}
