package main

import (
	"bytes"
	"encoding/gob"
	"net"
	"os"
	"strings"
	"time"

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

		lastState *common.StateMoveMessage
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
	var resend, waitsResponse bool
	for {
		if resend {
			log.Printf("does not receive response. so resend state: %+v\n", state)
			lastState = common.ShallowCopyStateMoveMessage(state)
		} else if !state.IsSubsequent(lastState) {
			// サーバーが最新でないリクエストに対してレスポンスを送ってきた場合は破棄をするのみで再送はしない
			log.Printf("does not match sequece, discard it: state: %v, last: %v \n", state, lastState)
			waitsResponse = true
		} else if !state.VerifyMove(lastState) {
			state = common.ShallowCopyStateMoveMessage(lastState)
			log.Printf("will resend last state: %+v\n", state)
		} else if state.HasFinish() {
			log.Println("lose")
			os.Exit(0)
		} else {
			state.MoveCoin()
			lastState = common.ShallowCopyStateMoveMessage(state)
		}

		if !waitsResponse {
			var (
				buf     *bytes.Buffer = &bytes.Buffer{}
				encoder               = gob.NewEncoder(buf)
			)
			if err := encoder.Encode(state); err != nil {
				log.Fatalf(err, "failed to encode state")
			}
			log.Printf("send state  state: %v\n", state)
			if _, err := ln.WriteToUDP(buf.Bytes(), serverAddr); err != nil {
				log.Fatalf(err, "failed to send data")
			}

		}
		var (
			buf     *bytes.Buffer = &bytes.Buffer{}
			decoder               = gob.NewDecoder(buf)
		)

		// encodeステート情報が入っているのでリセットする
		buf.Reset()
		b := make([]byte, 1024)
		err = ln.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			log.Fatalf(err, "failed to set read deadline")
		}
		_, _, err = ln.ReadFromUDP(b)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				log.Println("time out")
				resend = true
				continue
			}
			log.Fatalf(err, "failed to read server data")
		}

		buf.Write(b)
		state = &common.StateMoveMessage{}
		if err := decoder.Decode(state); err != nil {
			log.Fatalf(err, "failed to decode server response")
		}
		log.Printf("recieved data: %v, lastState: %v \n", state, lastState)
		resend = false
		waitsResponse = false
	}
}
