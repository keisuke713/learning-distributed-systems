package main

import (
	"fmt"
	"net"

	"github.com/DistributedClocks/tracing"
	"github.com/keisuke713/cs416/nim/common"
)

/** Config struct **/

// type ClientConfig struct {
// 	ClientAddress        string
// 	NimServerAddress     string
// 	TracingServerAddress string
// 	Secret               []byte
// 	TracingIdentity      string
// }

/** Tracing structs **/

// type GameStart struct {
// 	Seed int8
// }

// type ClientMove StateMoveMessage

// type ServerMoveReceive StateMoveMessage

// type GameComplete struct {
// 	Winner string
// }

// /** Message structs **/

// type StateMoveMessage struct {
// 	GameState []uint8
// 	MoveRow   int8
// 	MoveCount int8
// }

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Println("Usage: client.go [seed]")
	// 	return
	// }
	// arg, err := strconv.Atoi(os.Args[1])
	// CheckErr(err, "Provided seed could not be converted to integer", arg)
	// seed := int8(arg)
	seed := int8(6)

	tracingServer := tracing.NewTracingServerFromFile("../config/tracing_server_config.json")
	err := tracingServer.Open()
	if err != nil {
		common.CheckErr(err, "failed to open")
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

	clientAddress, err := net.ResolveUDPAddr("udp", config.ClientAddress)
	if err != nil {
		common.CheckErr(err, "failed to resolve client address")
	}

	serverAddress, err := net.ResolveUDPAddr("udp", config.NimServerAddress)
	if err != nil {
		common.CheckErr(err, "failed to resolve server address")
	}
	conn, err := net.DialUDP("udp", clientAddress, serverAddress)
	if err != nil {
		common.CheckErr(err, "failed to connect to server")
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	conn.Write([]byte(fmt.Sprintf("hello from %s \n", config.ClientAddress)))
	// for true {
	// 	state := StateMoveMessage{
	// 		nil,
	// 		-1,
	// 		seed,
	// 	}
	// 	encoder := gob.NewEncoder(conn)
	// 	if err := encoder.Encode(state); err != nil {
	// 		CheckErr(err, "failed to encode start state")
	// 	}
	// }

}
