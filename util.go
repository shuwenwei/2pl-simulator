package homework

import "log"

var Debug = false

func Dprintf(format string, values ...interface{}) {
	if !Debug {
		return
	}
	log.Printf(format, values...)
}