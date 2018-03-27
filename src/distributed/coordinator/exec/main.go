package main

import (
	"fmt"

	"github.com/limianwang/test-queue/src/distributed/coordinator"
)

func main() {
	ql := coordinator.NewQueueListener()
	go ql.ListenForNewSource()

	var a string
	fmt.Scanln(&a)
}
