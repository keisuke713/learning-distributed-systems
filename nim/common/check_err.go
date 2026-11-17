package common

import (
	"log"
	"os"
)

func CheckErr(err error, errfmsg string) {
	if err != nil {
		log.Printf("err: %v, errmsg: %s \n", err.Error(), errfmsg)
		os.Exit(1)
	}
}
