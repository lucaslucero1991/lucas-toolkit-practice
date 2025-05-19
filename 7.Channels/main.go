package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ContextKeyValue es un tipo seguro para evitar errores en las claves de contexto.
type ContextKeyValue string

// UserIDKey es la clave para almacenar el ID del usuario en el contexto.
const UserIDKey ContextKeyValue = "userID"

// Truck define la interfaz para camiones que cargan y descargan carga.
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

// processLoad ejecuta la carga y descarga de un camión, respetando el contexto.
func processLoad(ctx context.Context, truck Truck) error {
	log.Printf("Iniciando carga para camión %+v", truck)

	// Ejemplo de acceso a metadatos en el contexto.
	if userID, ok := ctx.Value(UserIDKey).(int); ok {
		log.Printf("Usuario procesando: %d", userID)
	}

	// Crear un contexto derivado con timeout de 4 segundos.
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel() // Liberar recursos al finalizar.

	// Simular una tarea que toma tiempo (3 segundos).
	select {
	case <-time.After(3 * time.Second):
		// Tarea completada dentro del tiempo permitido.
	case <-ctx.Done():
		// Contexto cancelado (por timeout o cancelación manual).
		return fmt.Errorf("procesamiento cancelado: %w", ctx.Err())
	}

	// Ejecutar operaciones del camión.
	if err := truck.LoadCargo(); err != nil {
		return fmt.Errorf("error cargando camión: %w", err)
	}
	if err := truck.UnloadCargo(); err != nil {
		return fmt.Errorf("error descargando camión: %w", err)
	}

	log.Printf("Carga finalizada para camión %+v", truck)
	return nil
}

// asyncProcessTrucks procesa una lista de camiones de forma concurrente.
func asyncProcessTrucks(ctx context.Context, trucks []Truck) error {
	var wg sync.WaitGroup

	for _, truck := range trucks {
		wg.Add(1)
		go func(t Truck) {
			defer wg.Done()
			if err := processLoad(ctx, t); err != nil {
				log.Printf("Error procesando camión %+v: %v", t, err)
			}
		}(truck)
	}

	wg.Wait()
	return nil
}

func main() {
	// Crear un contexto base.
	ctx := context.Background()

	// Añadir metadatos al contexto (ID de usuario).
	ctx = context.WithValue(ctx, UserIDKey, 24)

	// Lista de camiones a procesar.
	trucks := []Truck{
		&NormalTruck{id: "truck-1", cargo: 0},
		&ComplexTruck{id: "truck-2", cargo: 0, battery: 100},
		&NormalTruck{id: "truck-3", cargo: 0},
		&ComplexTruck{id: "truck-4", cargo: 0, battery: 80},
	}

	// Procesar camiones concurrentemente.
	if err := asyncProcessTrucks(ctx, trucks); err != nil {
		log.Fatalf("Error procesando camiones: %v", err)
	}

	fmt.Println("¡Todos los camiones fueron procesados exitosamente!")
}
