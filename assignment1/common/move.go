package common

import (
	"math/rand"
)

type GameStart struct {
	Seed int8
}

type ClientMove StateMoveMessage

type ServerMoveReceive StateMoveMessage

type GameComplete struct {
	Winner string
}

/** Message structs **/

type StateMoveMessage struct {
	GameState []uint8
	MoveRow   int8
	MoveCount int8
}

func InitialStateMoveMessage(seed int8) *StateMoveMessage {
	// game start
	return &StateMoveMessage{
		GameState: nil,
		MoveRow:   -1,
		MoveCount: seed,
	}
}

func (s *StateMoveMessage) HasFinish() bool {
	if s == nil {
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
	s.MoveRow = int8(targetRowI)
	ncount := 0
	// 取り除くコイン数をランダムに設定
	for ncount == 0 {
		ncount = rand.Intn(int(s.GameState[targetRow] + 1))
	}
	s.MoveCount = int8(ncount)
	s.GameState[targetRow] = s.GameState[targetRow] - uint8(ncount)
}
