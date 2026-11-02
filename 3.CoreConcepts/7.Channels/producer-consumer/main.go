package main

import (
	"fmt"
	"sync"
	"time"
)

// Ejemplo 2: productor/consumidor con un channel buffered.
//
// El buffer (3) actúa de amortiguador: el productor puede ir hasta 3 valores
// "por delante" del consumidor sin esperar. El cuarto envío bloquea hasta que
// el consumidor saque algo, y ese bloqueo natural coordina a ambos sin que
// escribamos ningún lock.
//
// A diferencia de la versión con time.Sleep, aquí coordinamos con WaitGroup:
// main no termina hasta que el consumidor procesó todo.
func main() {
	ch := make(chan int, 3)

	var wg sync.WaitGroup
	wg.Add(1)

	// Consumidor: recibe con `range` hasta que el channel se cierra.
	// Es LENTO a propósito (800ms por item) para que se note cómo el
	// productor se adelanta hasta llenar el buffer y luego debe esperar.
	go func() {
		defer wg.Done()
		for n := range ch {
			fmt.Println("  recibido:", n)
			time.Sleep(800 * time.Millisecond)
		}
	}()

	// Productor: envía 5 valores a un buffer de 3.
	// Los primeros 3 entran sin esperar; a partir del 4º, cada envío
	// bloquea hasta que el consumidor libere un hueco.
	for i := 1; i <= 5; i++ {
		fmt.Println("intentando enviar:", i)
		ch <- i
		fmt.Println("enviado:", i)
		time.Sleep(200 * time.Millisecond)
	}

	close(ch) // señal de "no envío más": destraba el `range` del consumidor.

	wg.Wait() // espera a que el consumidor termine de procesar todo.
	fmt.Println("fin")
}
