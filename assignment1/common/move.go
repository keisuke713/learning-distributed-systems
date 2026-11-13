package common

import (
	"log"
	"math/rand"
	"slices"
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
	SeqNumber uint8
	GameState []uint8
	MoveRow   int8
	MoveCount int8
}

func InitialStateMoveMessage(seed int8) *StateMoveMessage {
	// game start
	return &StateMoveMessage{
		SeqNumber: uint8(seed),
		GameState: nil,
		MoveRow:   -1,
		MoveCount: -1,
	}
}

func ShallowCopyStateMoveMessage(s *StateMoveMessage) *StateMoveMessage {
	if s == nil {
		return nil
	}
	return &StateMoveMessage{
		SeqNumber: s.SeqNumber,
		GameState: slices.Clone(s.GameState),
		MoveRow:   s.MoveRow,
		MoveCount: s.MoveCount,
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
