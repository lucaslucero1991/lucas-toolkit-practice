package main

import "fmt"

// Ejemplo 1: un channel buffered es asíncrono MIENTRAS tenga espacio,
// y se vuelve bloqueante (síncrono) en los bordes: cuando está lleno al
// enviar, o vacío al recibir.
//
// Todo este ejemplo corre en una sola goroutine (main) a propósito, para
// ver con claridad cuándo hay espacio y cuándo NO. Por eso las operaciones
// que bloquearían están comentadas: en una sola goroutine provocarían un
// deadlock (nadie del otro lado para destrabar la operación).
func main() {
	ch := make(chan int, 2) // buffer de 2

	ch <- 1 // no bloquea: hay espacio -> buffer [1]
	fmt.Println("enviado 1, buffer:", len(ch), "/", cap(ch))

	ch <- 2 // no bloquea: hay espacio -> buffer [1 2]
	fmt.Println("enviado 2, buffer:", len(ch), "/", cap(ch))

	// ch <- 3 // ⚠️ BLOQUEARÍA: buffer lleno y nadie recibe -> deadlock.
	//            Descoméntalo para ver: "all goroutines are asleep - deadlock".

	fmt.Println("recibido:", <-ch, "-> buffer:", len(ch), "/", cap(ch)) // saca 1
	fmt.Println("recibido:", <-ch, "-> buffer:", len(ch), "/", cap(ch)) // saca 2

	// <-ch // ⚠️ BLOQUEARÍA: buffer vacío y nadie envía -> deadlock.

	fmt.Println("fin")
}
