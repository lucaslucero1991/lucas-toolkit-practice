package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
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

// processLoad simula la carga y descarga de un camión.
func processLoad(ctx context.Context, truck Truck) error {
	log.Printf("Iniciando carga para camión %+v", truck)
	return errors.New("error manual simulando fallo para todos")
}

// asyncProcessTrucks procesa camiones concurrentemente, recolectando errores en un canal.
func asyncProcessTrucks(ctx context.Context, trucks []Truck) error {
	var wg sync.WaitGroup
	// Canal con buffer para evitar deadlocks, tamaño igual al número de camiones.
	errorsChan := make(chan error, len(trucks))

	for _, truck := range trucks {
		wg.Add(1) // Incrementar contador para cada goroutine.
		go func(t Truck) {
			defer wg.Done() // Decrementar contador al finalizar.
			if err := processLoad(ctx, t); err != nil {
				log.Printf("Error procesando camión %+v: %v", t, err)
				errorsChan <- err // Enviar error al canal.
			}
		}(truck)
	}

	wg.Wait()         // Esperar a que todas las goroutines terminen.
	close(errorsChan) // Cerrar el canal para permitir iteración con for range.

	// Iterar errores usando for range (buena práctica).
	var errors []error
	for err := range errorsChan {
		log.Printf("Error procesado: %v", err)
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		log.Printf("Cantidad de errores: %d", len(errors))
		return fmt.Errorf("se encontraron %d errores durante el procesamiento", len(errors))
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Lista de camiones a procesar.
	trucks := []Truck{
		&NormalTruck{id: "truck-1", cargo: 0},
		&ComplexTruck{id: "truck-2", cargo: 0, battery: 100}, // Generará un error.
		&NormalTruck{id: "truck-3", cargo: 0},
		&ComplexTruck{id: "truck-4", cargo: 0, battery: 80},
	}

	// Procesar camiones concurrentemente.
	if err := asyncProcessTrucks(ctx, trucks); err != nil {
		log.Fatalf("Error procesando camiones: %v", err)
	}

	fmt.Println("¡Todos los camiones fueron procesados exitosamente!")
}
