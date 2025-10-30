package logger

import "log"

func Info(msg string) {
	log.Println("[INFO]", msg)
}

func Error(err error) {
	log.Fatalln("[ERROR]", err)
}

func Fatalln(msg string) {
	log.Fatalln("[FATAL]", msg)
}
