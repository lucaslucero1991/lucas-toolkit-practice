package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Truck define la interfaz para camiones.
type Truck interface {
	LoadCargo() error
	UnloadCargo() error
}

// NormalTruck es un camión simple con ID y carga.
type NormalTruck struct {
	id    string
	cargo int
}

func (t *NormalTruck) LoadCargo() error {
	t.cargo += 1
	return nil
}

func (t *NormalTruck) UnloadCargo() error {
	t.cargo -= 1
	return nil
}

// ComplexTruck es un camión con batería adicional.
type ComplexTruck struct {
	id      string
	cargo   int
	battery float32
}

func (t *ComplexTruck) LoadCargo() error {
	t.cargo += 1
	t.battery += 1
	return nil
}

func (t *ComplexTruck) UnloadCargo() error {
	t.cargo -= 1
	t.battery -= 1
	return nil
}

// processLoad simula la carga y descarga de un camión, con un retardo.
func processLoad(truck Truck) error {
	log.Printf("Iniciando carga para camión %+v", truck)
	// Simular una tarea que toma tiempo (1 segundo).
	time.Sleep(1 * time.Second)

	if err := truck.LoadCargo(); err != nil {
		return fmt.Errorf("error cargando camión: %w", err)
	}
	if err := truck.UnloadCargo(); err != nil {
		return fmt.Errorf("error descargando camión: %w", err)
	}

	log.Printf("Carga finalizada para camión %+v", truck)
	return nil
}

// asyncProcessTrucks procesa camiones concurrentemente usando goroutines y WaitGroup.
func asyncProcessTrucks(trucks []Truck) error {
	var wg sync.WaitGroup

	for _, truck := range trucks {
		wg.Add(1) // Incrementar el contador para cada goroutine.
		go func(t Truck) {
			defer wg.Done() // Decrementar el contador al finalizar.
			if err := processLoad(t); err != nil {
				log.Printf("Error procesando camión %+v: %v", t, err)
			}
		}(truck)
	}

	wg.Wait() // Esperar a que todas las goroutines terminen.
	return nil
}

// syncProcessTruck lanza goroutines sin sincronización (experimental).
// Nota: Esta función está diseñada para probar el error de no usar WaitGroup,
// lo que puede causar que el programa termine antes de que las goroutines finalicen.
func syncProcessTruck(trucks []Truck) error {
	for _, t := range trucks {
		go processLoad(t)
	}
	return nil
}

func main() {
	// Lista de camiones a procesar.
	trucks := []Truck{
		&NormalTruck{id: "truck-1", cargo: 0},
		&ComplexTruck{id: "truck-2", cargo: 0, battery: 100},
		&NormalTruck{id: "truck-3", cargo: 0},
		&ComplexTruck{id: "truck-4", cargo: 0, battery: 80},
	}

	// Procesar camiones concurrentemente con sincronización.
	if err := asyncProcessTrucks(trucks); err != nil {
		log.Fatalf("Error procesando camiones: %v", err)
	}

	// Opcional: Descomenta para probar syncProcessTruck y observar el problema de no sincronizar.
	// if err := syncProcessTruck(trucks); err != nil {
	// 	log.Fatalf("Error procesando camiones: %v", err)
	// }

	fmt.Println("¡Todos los camiones fueron procesados exitosamente!")
}
