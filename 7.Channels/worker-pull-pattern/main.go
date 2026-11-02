package main

import (
	"fmt"
	"time"
)

// worker es una goroutine que procesa tareas (jobs) desde un canal de solo lectura
// y envía resultados a otro canal (results) de solo escritura.
// Cada worker tiene un ID para identificar sus logs.
func worker(id int, jobs <-chan int, results chan<- int) {
	// Iterar sobre el canal jobs usando for range, que termina cuando jobs se cierra.
	for j := range jobs {
		fmt.Println("worker", id, "started job", j)
		// Simular trabajo que toma tiempo (1 segundo).
		time.Sleep(time.Second)
		fmt.Println("worker", id, "finished job", j)
		// Enviar el resultado (job * 2) al canal results.
		results <- j * 2
	}
}

func main() {
	// Definir constantes para el número de tareas y workers.
	const numJobs = 5    // Total de tareas a procesar.
	const numWorkers = 2 // Número de goroutines (workers) en el pool.

	// Crear canales con buffer para evitar deadlocks.
	// Buffer igual a numJobs asegura que todas las tareas puedan enviarse sin bloquear.
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// Lanzar un pool de workers como goroutines.
	// Cada worker procesa tareas desde jobs y envía resultados a results.
	for w := 0; w < numWorkers; w++ {
		go worker(w, jobs, results)
	}

	// Enviar tareas (números 1 a numJobs) al canal jobs.
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}

	// Cerrar el canal jobs para indicar que no hay más tareas.
	// Esto permite que los workers terminen al procesar todas las tareas.
	close(jobs)

	// Recolectar resultados desde el canal results.
	// Iterar numJobs veces asegura que todas las tareas se completen.
	// Nota: No usamos WaitGroup porque esperamos todos los resultados aquí.
	for a := 1; a <= numJobs; a++ {
		<-results
	}

	// Opcional: No cerramos results porque no se lee después, pero podría cerrarse si se reutiliza.
	// close(results)
}
