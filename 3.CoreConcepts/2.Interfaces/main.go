package main

import "log"

// se usa para crear software mantenible y escalable en el tiempo
// sin atar logica a una implementacion concreta

func main() {
	person := make(map[string]any)
	person["age"] = 20
	person["name"] = "Lucas"

	age, exist := person["age"].(int)
	if !exist {
		log.Fatalf("Error processing age")
	}

	log.Printf("Age: %d \n", age)
}
