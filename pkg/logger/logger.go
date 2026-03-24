package logger

import "log"

func Info(msg string) {
	log.Println("[INFO]", msg)
}

func Warn(msg string) {
	log.Println("[WARN]", msg)
}

func Error(err error) {
	if err == nil {
		return
	}
	log.Println("[ERROR]", err)
}

func Fatalln(msg string) {
	log.Fatalln("[FATAL]", msg)
}
