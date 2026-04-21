package main

import (
	"fmt"
	"time"
)

func printf(format string, a ...any) {
	print(fmt.Sprintf(format, a...))
}

func print(a ...any) {
	a = append([]any{time.Now().In(time.FixedZone("CST", 28800)).Format("\r[01-02 15:04:05]")}, a...)
	fmt.Println(a...)
}
