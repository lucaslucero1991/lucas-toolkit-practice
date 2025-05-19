package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
	- Permite escribir software concurrente, ejecutar tareas al mismo tiempo
	- Version ligera de un hilo,
	- Con go, pero tiene algunos problemas a resolver:
	- el programa termina y los procesos todavia no
	- Usamos WaitGroup waits for a collection of goroutines to finish.
	- Usamos goroutinas para hacer multiples tareas que llevan tiempo, y luego
	 buscamos una manera syncronica de agrupar esos resultados, luego usaremos canales
*/

type Truck interface {
	LoadCargo() error
	UnloadCargo() error
}

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

func asyncProcessTrucks(trucks []Truck) error {
	var wg sync.WaitGroup

	for _, t := range trucks {
		wg.Add(1)
		go func(t Truck) {
			processLoad(t)
			wg.Done()
		}(t)
	}

	wg.Wait()
	return nil
}

func syncProcessTruck(trucks []Truck) error {
	for _, t := range trucks {
		go processLoad(t)
	}
	return nil
}

func processLoad(truck Truck) {
	log.Printf("Init load cargo from %+v", truck)
	time.Sleep(1 * time.Second)
	err := truck.LoadCargo()
	if err != nil {
		log.Println(err)
	}

	err = truck.UnloadCargo()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Finish load cargo from %+v", truck)

}

func main() {
	trucks := []Truck{
		&NormalTruck{id: "truck-1", cargo: 0},
		&ComplexTruck{id: "truck-2", cargo: 0, battery: 100},
		&NormalTruck{id: "truck-3", cargo: 0},
		&ComplexTruck{id: "truck-4", cargo: 0, battery: 80},
	}

	if err := asyncProcessTrucks(trucks); err != nil {
		log.Println(err)
	}

	time.Sleep(2 * time.Second)
	fmt.Println("All trucks are process successfully!")
}
