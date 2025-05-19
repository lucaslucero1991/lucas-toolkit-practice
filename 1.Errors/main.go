package main

import (
	"errors"
	"fmt"
	"log"
)

// usar la interface error, siempre va al final como convencion
// se manejan de manera explicita, si no lo manejo go no hace nada,

type Truck struct {
	id string
}

func processTrack(truck Truck) error {
	fmt.Printf("Processing truck %s\n", truck.id)

	return errors.New("error error")
}

func main() {
	trucks := []Truck{
		Truck{"truck-1"},
		Truck{"truck-2"},
		Truck{"truck-3"},
	}
	for _, truck := range trucks {
		fmt.Printf("Truck %s is arrived.\n", truck.id)
		err := processTrack(truck)
		if err != nil {
			log.Fatalf("Error processing truck %s", err)
		}

		if err := processTrack(truck); err == nil {

		}
	}
}
