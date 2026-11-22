package common

import (
	"log"
	"math/rand"
	"slices"

	"github.com/DistributedClocks/tracing"
)

type (
	GameStart struct {
		Seed int8
	}

	ClientMove        StateMoveMessage
	ServerMoveReceive StateMoveMessage
	ServerMove        StateMoveMessage
	ClietnMoveReceive StateMoveMessage

	// クライアントが新しいサーバーに対してリクエストを送るようになったとき
	NewNimServer struct {
		NimServerAddress string
	}

	// サーバーが死んだのをクライアント側で検知したとき
	NimServerFailed NewNimServer

	AllNimServersDown struct{}

	GameComplete struct {
		Winner string
	}

	// ゲーム開始のリクエストをクライアント側から受け取ったとき
	ServerGameStart struct {
		NimServerAddress string
	}

	// ゲーム再開のリクエストをクライアントから受け取った時
	GameResume ServerGameStart
)

/** Message structs **/

type StateMoveMessage struct {
	SeqNumber         uint8
	GameState         []uint8
	MoveRow           int8
	MoveCount         int8
	TracingServerAddr string
	Token             tracing.TracingToken
}

func InitialStateMoveMessage(seed int8, tracingServerAddr string, token tracing.TracingToken) *StateMoveMessage {
	return &StateMoveMessage{
		SeqNumber:         uint8(seed),
		GameState:         nil,
		MoveRow:           -1,
		MoveCount:         -1,
		TracingServerAddr: tracingServerAddr,
		Token:             token,
	}
}

func ShallowCopyStateMoveMessage(s *StateMoveMessage) *StateMoveMessage {
	if s == nil {
		return nil
	}
	return &StateMoveMessage{
		SeqNumber:         s.SeqNumber,
		GameState:         slices.Clone(s.GameState),
		MoveRow:           s.MoveRow,
		MoveCount:         s.MoveCount,
		TracingServerAddr: s.TracingServerAddr,
		Token:             s.Token,
	}
}

func (s *StateMoveMessage) HasFinish() bool {
	if s == nil || len(s.GameState) == 0 {
		return false
	}

	for i := 0; i < len(s.GameState); i++ {
		if s.GameState[i] > 0 {
			return false
		}
	}
	return true
}

func (s *StateMoveMessage) MoveCoin() {
	if len(s.GameState) == 0 {
		return
	}

	nrows := len(s.GameState)
	targetRows := make([]int, 0)
	// まだコインが置いてある行を特定する
	for i := 0; i < nrows; i++ {
		if s.GameState[i] > 0 {
			targetRows = append(targetRows, i)
		}
	}
	// ターゲット行をランダムに設定
	targetRowI := rand.Intn(len(targetRows))
	targetRow := targetRows[targetRowI]
	s.MoveRow = int8(targetRow)
	ncount := 0
	// 取り除くコイン数をランダムに設定
	for ncount == 0 {
		ncount = rand.Intn(int(s.GameState[targetRow] + 1))
	}
	s.MoveCount = int8(ncount)
	s.GameState[targetRow] = s.GameState[targetRow] - uint8(ncount)

	s.IncSeq()
}

func (s *StateMoveMessage) VerifyMove(o *StateMoveMessage) bool {
	if s == nil || o == nil {
		return true
	}
	// ゲーム開始
	if s.MoveRow == -1 {
		return true
	}
	expectedState := slices.Clone(o.GameState)
	expectedState[s.MoveRow] -= uint8(s.MoveCount)
	log.Printf("s: %v, o: %v, expectedState: %v \n", s, o, expectedState)
	if isEqual := slices.Equal(s.GameState, expectedState); isEqual {
		return true
	}
	return false
}

func (s *StateMoveMessage) IncSeq() {
	if s == nil {
		return
	}
	s.SeqNumber += 1
}

func (s *StateMoveMessage) IsSubsequent(o *StateMoveMessage) bool {
	if s == nil || o == nil {
		return true
	}
	// ゲーム開始
	if s.MoveRow == -1 {
		return true
	}
	return s.SeqNumber-o.SeqNumber == 1
}
