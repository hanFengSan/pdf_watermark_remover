package logutil

import (
	"fmt"
	"os"
	"strings"
)

func Prefix() string {
	return strings.TrimSpace(os.Getenv("WM_LOG_PREFIX"))
}

func Println(msg string) {
	p := Prefix()
	if p == "" {
		fmt.Println(msg)
		return
	}
	fmt.Printf("%s %s\n", p, msg)
}

func Printf(format string, args ...any) {
	p := Prefix()
	if p == "" {
		fmt.Printf(format, args...)
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s", p, msg)
}
