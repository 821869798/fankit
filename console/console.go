package console

import (
	"fmt"
	"os"
)

func AnyKeyToQuit() {
	fmt.Printf("Press any key to exit...")
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}

func AnyKeyToQuitWithStr(str string) {
	fmt.Printf(str)
	b := make([]byte, 1)
	_, _ = os.Stdin.Read(b)
}
