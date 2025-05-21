package main

import (
	"errors"
	"log"
)

// NormalTruck usa un método por valor (copia).
type NormalTruck struct {
	id    string
	cargo int
}

func (t NormalTruck) LoadCargo() error {
	t.cargo += 1 // Solo modifica la copia local.
	return nil
}

// ComplexTruck usa un método por referencia (puntero).
type ComplexTruck struct {
	id    string
	cargo int
}

func (t *ComplexTruck) LoadCargo() error {
	t.cargo += 1
	return nil
}

// processCargo modifica un ComplexTruck usando un puntero.
func processCargo(t *ComplexTruck) error {
	if t == nil {
		return errors.New("cannot process nil truck")
	}
	t.cargo += 1
	return nil
}

func BasicFuncs() {
	truckID := 42
	anotherTruck := &truckID

	log.Printf("Dirección de truckID: %p", &truckID)
	log.Printf("Valor de anotherTruck (dirección): %p", anotherTruck)
	log.Printf("Valor apuntado por anotherTruck: %d", *anotherTruck)

	// Cambiar truckID afecta el valor apuntado por anotherTruck.
	truckID = 0
	log.Printf("Nuevo valor apuntado por anotherTruck: %d", *anotherTruck)
}

func main() {

	normalTruck := &NormalTruck{id: "truck-1", cargo: 0}
	normalTruck.LoadCargo()
	log.Printf("NormalTruck después de LoadCargo: %+v", normalTruck)

	complexTruck := &ComplexTruck{id: "truck-2", cargo: 0}
	complexTruck.LoadCargo() // Cambia cargo porque LoadCargo usa un puntero.
	log.Printf("ComplexTruck después de LoadCargo: %+v", complexTruck)

	err := processCargo(complexTruck)
	if err != nil {
		log.Printf("Error en processCargo: %v", err)
	}
	log.Printf("ComplexTruck después de processCargo: %+v", complexTruck)

	err = processCargo(nil)
	if err != nil {
		log.Printf("Error esperado con nil: %v", err)
	}

	// Demostrar conceptos básicos de punteros (descomentar para probar).
	//BasicFuncs()
}
