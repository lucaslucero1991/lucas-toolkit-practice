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

// LoadCargo no modifica el struct original porque trabaja con una copia.
func (t NormalTruck) LoadCargo() error {
	t.cargo += 1 // Solo modifica la copia local.
	return nil
}

// ComplexTruck usa un método por referencia (puntero).
type ComplexTruck struct {
	id    string
	cargo int
}

// LoadCargo modifica el struct original porque usa un puntero.
func (t *ComplexTruck) LoadCargo() error {
	t.cargo += 1 // Modifica el valor en la dirección de memoria.
	return nil
}

// processCargo modifica un ComplexTruck usando un puntero.
func processCargo(t *ComplexTruck) error {
	// Verificar nil para evitar pánico.
	if t == nil {
		return errors.New("cannot process nil truck")
	}
	t.cargo += 1 // Modifica el valor en la dirección de memoria.
	return nil
}

// BasicFuncs demuestra conceptos básicos de punteros.
func BasicFuncs() {
	truckID := 42
	anotherTruck := &truckID // Puntero a truckID.

	// Imprimir dirección de memoria de truckID.
	log.Printf("Dirección de truckID: %p", &truckID)
	// Imprimir dirección almacenada en el puntero anotherTruck.
	log.Printf("Valor de anotherTruck (dirección): %p", anotherTruck)
	// Imprimir valor al que apunta anotherTruck.
	log.Printf("Valor apuntado por anotherTruck: %d", *anotherTruck)

	// Cambiar truckID afecta el valor apuntado por anotherTruck.
	truckID = 0
	log.Printf("Nuevo valor apuntado por anotherTruck: %d", *anotherTruck)
}

func main() {
	// Demostrar método por valor (copia).
	normalTruck := &NormalTruck{id: "truck-1", cargo: 0}
	normalTruck.LoadCargo() // No cambia cargo porque LoadCargo usa una copia.
	log.Printf("NormalTruck después de LoadCargo: %+v", normalTruck)

	// Demostrar método por referencia (puntero).
	complexTruck := &ComplexTruck{id: "truck-2", cargo: 0}
	complexTruck.LoadCargo() // Cambia cargo porque LoadCargo usa un puntero.
	log.Printf("ComplexTruck después de LoadCargo: %+v", complexTruck)

	// Demostrar función con puntero y manejo de nil.
	err := processCargo(complexTruck)
	if err != nil {
		log.Printf("Error en processCargo: %v", err)
	}
	log.Printf("ComplexTruck después de processCargo: %+v", complexTruck)

	// Probar con nil (manejo seguro).
	err = processCargo(nil)
	if err != nil {
		log.Printf("Error esperado con nil: %v", err)
	}

	// Demostrar conceptos básicos de punteros (descomentar para probar).
	BasicFuncs()
}
