package helper

import "log"

func Log(v ...interface{}) {
	log.Printf("%+v\n", v)
}

func Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
