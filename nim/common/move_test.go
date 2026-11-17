package common

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateMoveMessage test", func() {
	Describe("VerifyMove", func() {
		var (
			s *StateMoveMessage
			o *StateMoveMessage

			verifies bool
		)

		JustBeforeEach(func() {
			verifies = s.VerifyMove(o)
		})
		JustAfterEach(func() {
			s = nil
			o = nil
		})
		Context("state is nil", func() {
			It("trueになる", func() {
				Expect(verifies).To(BeTrue())
			})
		})
		Context("past state is nil", func() {
			It("trueになる", func() {
				Expect(verifies).To(BeTrue())
			})
		})
		Context("game start state", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					MoveRow: -1,
				}
			})
			It("trueになる", func() {
				Expect(verifies).To(BeTrue())
			})
		})
		Context("correct move", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					GameState: []uint8{4, 4, 0},
					MoveRow:   2,
					MoveCount: 1,
				}
				o = &StateMoveMessage{
					GameState: []uint8{4, 4, 1},
					MoveRow:   2,
					MoveCount: 3}
			})
			It("trueになる", func() {
				Expect(verifies).To(BeTrue())
			})
		})
		Context("inaccurate move", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					GameState: []uint8{4, 3, 1}, // row2の代わりに1が1つ減っている
					MoveRow:   2,
					MoveCount: 1,
				}
				o = &StateMoveMessage{
					GameState: []uint8{4, 4, 1},
					MoveRow:   2,
					MoveCount: 3}
			})
			It("falseになる", func() {
				Expect(verifies).To(BeFalse())
			})
		})
	})
	Describe("IsSubsequent", func() {
		var (
			s *StateMoveMessage
			o *StateMoveMessage

			isSubsequent bool
		)

		JustBeforeEach(func() {
			isSubsequent = s.IsSubsequent(o)
		})
		JustAfterEach(func() {
			s = nil
			o = nil
		})
		Context("state is nil", func() {
			It("trueになる", func() {
				Expect(isSubsequent).To(BeTrue())
			})
		})
		Context("past state is nil", func() {
			It("trueになる", func() {
				Expect(isSubsequent).To(BeTrue())
			})
		})
		Context("game start state", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					MoveRow: -1,
				}
			})
			It("trueになる", func() {
				Expect(isSubsequent).To(BeTrue())
			})
		})
		Context("correct seq", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					SeqNumber: 2,
				}
				o = &StateMoveMessage{
					SeqNumber: 1,
				}
			})
			It("trueになる", func() {
				Expect(isSubsequent).To(BeTrue())
			})
		})
		Context("incorrect seq", func() {
			BeforeEach(func() {
				s = &StateMoveMessage{
					SeqNumber: 2,
				}
				o = &StateMoveMessage{
					SeqNumber: 2,
				}
			})
			It("trueになる", func() {
				Expect(isSubsequent).To(BeFalse())
			})
		})
	})
})
