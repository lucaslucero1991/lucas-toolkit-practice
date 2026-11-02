package main

import (
	"fmt"
	"strings"
)

type User struct {
	Name string
	Age  int
}

func (u User) IsAdult() bool {
	return u.Age >= 18
}

func MostrarUsuarioMayorDeEdad() {
	user := User{
		Name: "John",
		Age:  20,
	}

	fmt.Printf("Processing User:\n %+v.\n Is Adult? %v", user, user.IsAdult())
}

func ContadorDePalabras() {
	contadorPalabras := make(map[string]int)
	parrafo := strings.Split(strings.ToLower("gato perro casa perro sol"+
		" luna gato perro auto casa luna sol perro"), " ")

	for _, palabra := range parrafo {
		if palabra != "" {
			contadorPalabras[palabra]++
		}
	}

	for k, v := range contadorPalabras {
		fmt.Printf("%s: %d\n", k, v)
	}
}

const (
	Lunes = iota
	Martes
	Miercoles
	Jueves
	Viernes
	Sabado
	Domingo
)

func main() {

	// 1. Mostrar si un usuario es mayo de edad
	//MostrarUsuarioMayorDeEdad()

	// 2. Contar frecuencias usando un map
	ContadorDePalabras()

	// 3. Enum para dias de la semana
	//fmt.Printf("Processing lunes and Martes\n %+v.\n", Lunes)
}
