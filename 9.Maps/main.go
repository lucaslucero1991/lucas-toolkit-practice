package main

import (
	"fmt"
	"log"
	"maps"
	"sync"
	"time"
)

/*
	los mapas se usan cuando sabes la key, son mas rapidos, es un conjunto de clave/valor
	las claves son unicas
	se puede usar por sobre un slice cuando conoces la key.
	no son tan seguras cuando usas gourtinas, pueden suceder race conditions
	se puede usar la lib maps para clonar, copiar, validar si son iguales,
	no me queda muy claro el uso de estos ultmos ejemplos de la libreria maps tal vez me

puedas yudar a ponerlos en el codigo.
*/
func simpleUseCase() {
	intMap := make(map[string]int)
	intMap["primero"] = 1
	intMap["segundo"] = 2
	intMap["tercero"] = 3
	log.Printf("creamos un mapa: %+v", intMap)

	intMap["primero"] = 10
	log.Printf("cambiamos el valor del primer elemento: %+v", intMap)

	delete(intMap, "tercero")
	log.Printf("borramos el ultimo elemento: %+v", intMap)

	if _, ok := intMap["primero"]; ok {
		log.Printf("el valor de la key 'primero' es: %+v", intMap["primero"])
	}

	if _, ok := intMap["cualquiera"]; !ok {
		log.Printf("no existe el elemento buscado: %+v", intMap["cualquiera"])
	}

	clear(intMap)
	log.Printf("borramos todo con clear(): %+v", intMap)

	maps.Equal(intMap, intMap)
	//secondMap := maps.Clone(intMap)

}

func raceCondition() {
	var wg sync.WaitGroup
	m := make(map[string]int)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			time.Sleep(1 * time.Second)

			// se puede generar una condicion de carrera
			m[fmt.Sprintf("key-%d", i)] = i
		}(i)
	}

	wg.Wait() // Esperar a que todas las goroutines terminen.
	log.Println(m)
}

func main() {

	// casos de uso simple con mapas
	//simpleUseCase()

	// problema de condicion de carrera con mapas
	raceCondition()
}
