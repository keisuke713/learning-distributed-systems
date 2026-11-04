package common

import (
	"fmt"
	"log"
)

type Log struct {
	env string
}

func NewLog(env string) *Log {
	return &Log{env: env}
}

func (l *Log) Print(msg string) {
	message := fmt.Sprintf("[%s] %s \n", l.env, msg)
	log.Print(message)
}

func (l *Log) Println(msg string) {
	message := fmt.Sprintf("[%s] %s \n", l.env, msg)
	log.Println(message)
}

func (l *Log) Printf(format string, v ...any) {
	message := fmt.Sprintf("[%s] %s \n", l.env, format)
	log.Printf(message, v...)
}

func (l *Log) Fatalf(err error, format string, v ...any) {
	message := fmt.Sprintf("[%s] err: %s, %s \n", l.env, err.Error(), format)
	log.Fatalf(message, v)
}
