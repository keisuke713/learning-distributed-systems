package client

import (
	"bytes"
	"encoding/gob"

	"net"
	"strings"
	"time"

	fcheck "learning-distributed-systems/fchecker"
	"learning-distributed-systems/nim/common"

	"github.com/DistributedClocks/tracing"
)

var (
	log = common.NewLog("client")
)

func Dial(done chan int) {
	// if len(os.Args) != 2 {
	// 	fmt.Println("Usage: client.go [seed]")
	// 	return
	// }
	// arg, err := strconv.Atoi(os.Args[1])
	// CheckErr(err, "Provided seed could not be converted to integer", arg)
	// seed := int8(arg)
	seed := int8(6)

	config := common.ParseConfig("./config/config.json")
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

	targetServerIndex := resolveNextServerIndex(len(config.NimServerAddressList), -1)
	serverAddr, err := net.ResolveUDPAddr("udp", config.NimServerAddressList[targetServerIndex])
	if err != nil {
		log.Fatalf(err, "failed to resolve server address")
	}
	trace.RecordAction(common.NewNimServer{NimServerAddress: serverAddr.String()})
	ch, err := fcheck.Start(fcheck.StartStruct{
		AckLocalIPAckLocalPort:       config.FCheckAckLocalAddr,
		EpochNonce:                   1,
		HBeatLocalIPHBeatLocalPort:   config.FCheckHbeatLocalAddr,
		HBeatRemoteIPHBeatRemotePort: config.FCheckHbeatRemoteAddrList[targetServerIndex],
		LostMsgThresh:                config.FCheckLostMsgsThresh,
	})
	if err != nil {
		log.Fatalf(err, "failed to start fcheck")
	}
	// ゲーム開始直後
	state := common.InitialStateMoveMessage(seed, config.TracingServerAddress, trace.GenerateToken())
	var (
		lastState *common.StateMoveMessage
	)

	var resend, waitsResponse bool
	for {
		select {
		case v := <-ch:
			log.Printf("stop server: %+v due to not respond to heartbeat\n", v)
			trace.RecordAction(common.NimServerFailed{NimServerAddress: serverAddr.String()})
			targetServerIndex = resolveNextServerIndex(len(config.NimServerAddressList), targetServerIndex)
			serverAddr, err = net.ResolveUDPAddr("udp", config.NimServerAddressList[targetServerIndex])
			if err != nil {
				log.Fatalf(err, "failed to resolve server address")
			}

			trace.RecordAction(common.NewNimServer{NimServerAddress: serverAddr.String()})
			ch, err = fcheck.Start(fcheck.StartStruct{
				AckLocalIPAckLocalPort:       config.FCheckAckLocalAddr,
				EpochNonce:                   1,
				HBeatLocalIPHBeatLocalPort:   config.FCheckHbeatLocalAddr,
				HBeatRemoteIPHBeatRemotePort: config.FCheckHbeatRemoteAddrList[targetServerIndex],
				LostMsgThresh:                config.FCheckLostMsgsThresh,
			})
			if err != nil {
				log.Fatalf(err, "failed to start fcheck")
			}
		default:
		}
		if resend {
			log.Printf("does not receive response. so resend state: %+v\n", state)
			state = common.ShallowCopyStateMoveMessage(lastState)
		} else if !state.IsSubsequent(lastState) {
			// サーバーが最新でないリクエストに対してレスポンスを送ってきた場合は破棄をするのみで再送はしない
			log.Printf("does not match sequece, discard it: state: %v, last: %v \n", state, lastState)
			waitsResponse = true
		} else if !state.VerifyMove(lastState) {
			state = common.ShallowCopyStateMoveMessage(lastState)
			log.Printf("will resend last state: %+v\n", state)
		} else if state.HasFinish() {
			log.Println("lose")
			done <- 1
			return
		} else {
			state.MoveCoin()
			lastState = common.ShallowCopyStateMoveMessage(state)
		}

		if !waitsResponse {
			var (
				buf     *bytes.Buffer = &bytes.Buffer{}
				encoder               = gob.NewEncoder(buf)
			)
			state.Token = trace.GenerateToken()
			if err := encoder.Encode(state); err != nil {
				log.Fatalf(err, "failed to encode state")
			}
			log.Printf("send state  state: %v\n", state)
			if _, err := ln.WriteToUDP(buf.Bytes(), serverAddr); err != nil {
				log.Fatalf(err, "failed to send data")
			}
			trace.RecordAction(
				common.ClientMove(*state),
			)

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
		trace = tracer.ReceiveToken(state.Token)
		trace.RecordAction(common.ServerMoveReceive(*state))
		resend = false
		waitsResponse = false
	}
}

func resolveNextServerIndex(nServer int, currentIndex int) int {
	return (currentIndex + 1) % nServer
}
