/*

This package specifies the API to the failure checking library to be
used in assignment 2 of UBC CS 416 2021W2.

You are *not* allowed to change the API below. For example, you can
modify this file by adding an implementation to Stop, but you cannot
change its API.

*/

package fcheck

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

////////////////////////////////////////////////////// DATA
// Define the message types fchecker has to use to communicate to other
// fchecker instances. We use Go's type declarations for this:
// https://golang.org/ref/spec#Type_declarations

// Heartbeat message.
type HBeatMessage struct {
	EpochNonce uint64 // Identifies this fchecker instance/epoch.
	SeqNum     uint64 // Unique for each heartbeat in an epoch.
}

// An ack message; response to a heartbeat.
type AckMessage struct {
	HBEatEpochNonce uint64 // Copy of what was received in the heartbeat.
	HBEatSeqNum     uint64 // Copy of what was received in the heartbeat.
}

// Notification of a failure, signal back to the client using this
// library.
type FailureDetected struct {
	UDPIpPort string    // The RemoteIP:RemotePort of the failed node.
	Timestamp time.Time // The time when the failure was detected.
}

////////////////////////////////////////////////////// API

type StartStruct struct {
	AckLocalIPAckLocalPort       string
	EpochNonce                   uint64
	HBeatLocalIPHBeatLocalPort   string
	HBeatRemoteIPHBeatRemotePort string
	LostMsgThresh                uint8
}

var (
	ctx, cancel = context.WithCancel(context.TODO())
)

func Start(arg StartStruct) (notifyCh <-chan FailureDetected, err error) {
	if arg.HBeatLocalIPHBeatLocalPort == "" {
		// ONLY arg.AckLocalIPAckLocalPort is set

		go func() {
			serverAddr, err := net.ResolveUDPAddr("udp", arg.AckLocalIPAckLocalPort)
			if err != nil {
				log.Fatalf("Failed to resolve local UDP address: %v", err)
			}
			ln, err := net.ListenUDP("udp", serverAddr)
			if err != nil {
				log.Fatalf("Failed to listen for UDP messages: %v", err)
			}
			defer func() {
				ln.Close()
				log.Println("Stopped listening for heartbeat messages on", arg.AckLocalIPAckLocalPort)
			}()
			log.Println("Listening for heartbeat messages on", arg.AckLocalIPAckLocalPort)

			for {
				b := make([]byte, 1024)
				n, clientAddr, err := ln.ReadFromUDP(b)
				if err != nil {
					log.Fatalf("Failed to read heartbeat message: %v", err)
				}

				var msg HBeatMessage
				buf := bytes.NewBuffer(b[:n])
				decoder := gob.NewDecoder(buf)
				if err := decoder.Decode(&msg); err != nil {
					log.Fatalf("Failed to decode heartbeat message: %v", err)
				}
				log.Printf("Received heartbeat: %+v on %s from %s\n", msg, serverAddr.String(), clientAddr.String())

				ack := AckMessage{
					HBEatEpochNonce: msg.EpochNonce,
					HBEatSeqNum:     msg.SeqNum,
				}
				buf.Reset()
				encoder := gob.NewEncoder(buf)
				if err := encoder.Encode(ack); err != nil {
					log.Fatalf("Failed to encode ack message: %v", err)
				}

				if n := rand.Intn(5); n < 5 {
					log.Print("Simulating intentional delay for testing purposes...")
					time.Sleep(20 * time.Second)
				}
				if _, err := ln.WriteToUDP(buf.Bytes(), clientAddr); err != nil {
					log.Fatalf("Failed to send ack message: %v", err)
				}
				log.Printf("Sent ack: %+v to %s from %s\n", ack, clientAddr.String(), serverAddr.String())
			}
		}()
		return nil, nil
	}

	ch := make(chan FailureDetected)
	go func(ch chan<- FailureDetected) {
		clientAddr, err := net.ResolveUDPAddr("udp", arg.HBeatLocalIPHBeatLocalPort)
		if err != nil {
			log.Fatalf("Failed to resolve local UDP address: %v", err)
		}
		ln, err := net.ListenUDP("udp", clientAddr)
		if err != nil {
			log.Fatalf("Failed to listen for UDP messages: %v", err)
		}
		defer func() {
			ln.Close()
			log.Println("Stopped listening for ack messages on", arg.HBeatLocalIPHBeatLocalPort)
		}()

		serverAddr, err := net.ResolveUDPAddr("udp", arg.HBeatRemoteIPHBeatRemotePort)
		if err != nil {
			log.Fatalf("Failed to resolve remote UDP address: %v", err)
		}

		msg := HBeatMessage{
			EpochNonce: arg.EpochNonce,
			SeqNum:     5,
		}

		nServerTimeout := 0
		rttSec := 3 * time.Second
		elapsedTime := 3 * time.Second

		startTime := make(map[int]time.Time)

		for {
			select {
			case <-ctx.Done():
				log.Println("stop monitoring node: ", arg.HBeatRemoteIPHBeatRemotePort)
				return
			default:
			}

			if nServerTimeout == int(arg.LostMsgThresh) {
				continue
			}

			var (
				buf     *bytes.Buffer = &bytes.Buffer{}
				encoder               = gob.NewEncoder(buf)
				decoder               = gob.NewDecoder(buf)
			)

			msg.SeqNum++ // Increment the sequence number for each heartbeat
			if err := encoder.Encode(msg); err != nil {
				log.Fatalf("Failed to encode heartbeat message: %v", err)
			}

			if _, err := ln.WriteToUDP(buf.Bytes(), serverAddr); err != nil {
				log.Fatalf("Failed to send heartbeat message: %v", err)
			}
			log.Printf("Sent heartbeat: %+v to %s\n", msg, serverAddr.String())
			startTime[int(msg.SeqNum)] = time.Now()

			buf.Reset()
			b := make([]byte, 1024)
			rttSec = (rttSec + elapsedTime) / 2
			log.Printf("next rtt duration: %v\n", rttSec)
			err = ln.SetReadDeadline(time.Now().Add(rttSec)) // Set a read deadline for the ack response
			if err != nil {
				log.Fatalf("Failed to set read deadline: %v", err)
			}
			_, tmpServerAddr, err := ln.ReadFromUDP(b)
			if err != nil {
				if strings.Contains(err.Error(), "i/o timeout") {
					nServerTimeout++
					rttSec = time.Duration(3) * time.Second
					log.Printf("Timeout waiting for ack from %s. Timeout count: %d\n", serverAddr.String(), nServerTimeout)
					if nServerTimeout >= int(arg.LostMsgThresh) {
						ch <- FailureDetected{
							UDPIpPort: serverAddr.String(),
							Timestamp: time.Now(),
						}
						log.Printf("Failure detected for %s after %d consecutive timeouts.\n", serverAddr.String(), nServerTimeout)
						return
					}
					continue
				}
				log.Fatalf("Failed to read ack message: %v", err)
			}
			// 過去に監視していたサーバーからレスポンスが来た場合discardする
			if tmpServerAddr.String() != serverAddr.String() {
				log.Printf("received ack from %s, but expected: %s\n", tmpServerAddr.String(), serverAddr.String())
				continue
			}

			buf.Write(b)
			var ack AckMessage
			if err := decoder.Decode(&ack); err != nil {
				log.Fatalf("Failed to decode ack message: %v", err)
			}
			if ack.HBEatEpochNonce != arg.EpochNonce || ack.HBEatSeqNum != msg.SeqNum {
				log.Printf("different epok nonce or seq number so discard it: expxected: %d, got: %d\n", msg.SeqNum, ack.HBEatSeqNum)
				continue
			}
			nServerTimeout = 0 // Reset timeout count on successful ack
			log.Printf("startTime: %+v\n", startTime)
			elapsedTime = time.Since(startTime[int(ack.HBEatSeqNum)])
			log.Printf("Received ack: %+v from %s, secNum: %d, elapsedTme: %v\n", ack, serverAddr.String(), ack.HBEatSeqNum, elapsedTime)
		}
	}(ch)
	return ch, nil
}

// Tells the library to stop monitoring/responding acks.
func Stop() {
	cancel()
}
