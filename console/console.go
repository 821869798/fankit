package console

import (
	"log"
	"os"
)

func AnyKeyToQuit() {
	log.Println("Press any key to exit...")
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}

func AnyKeyToQuitWithStr(str string) {
	log.Println(str)
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}
