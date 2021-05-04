package main

import (
	"log"
)

//"github.com/fm2901/wallet/pkg/wallet"

func main() {
	ch := tick()
	for i := range ch {
		log.Print(i)
	}	
}

func tick() <- chan int {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i ++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}
