package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTimer(5 * time.Second)
	fmt.Println("start")
	<-t.C
	fmt.Println("end")
}
