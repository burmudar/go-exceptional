package main

import (
	"fmt"
	"time"
)

func main() {
	expectedTime, err := time.Parse("2006-01-02 15:04:05.000", "2016-03-23 15:41:48,564")
	if err != nil {
		fmt.Printf("Time parse Error: %v\n", err)
	}
	fmt.Printf("Parsed Time: %v\n", expectedTime)
}
