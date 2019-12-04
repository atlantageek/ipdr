package main

import (
	"fmt"
	"github.com/prprprus/scheduler"
)
func keepAlive2() {
	fmt.Println("Keep alive 2")
}
func main() {
	s, schedulerErr := scheduler.NewScheduler(1000)
	if schedulerErr != nil  {
		panic(schedulerErr)
	}
	s.Every().Second(1).Do(keepAlive2)
	for {

	}
}