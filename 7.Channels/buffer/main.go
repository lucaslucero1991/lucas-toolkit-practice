package main

import (
	"fmt"
	"time"
)

func main() {

	ch := make(chan int, 3)

	// Productor
	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Printf("[Productor] Intentando enviar: %d\n", i)
			ch <- i
			fmt.Printf("[Productor] Enviado: %d\n", i)
			time.Sleep(50 * time.Millisecond)
		}
		close(ch)
	}()

	// Consumidor
	go func() {
		for val := range ch {
			fmt.Printf("  [Consumidor] Recibido: %d\n", val)
			time.Sleep(800 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
}
