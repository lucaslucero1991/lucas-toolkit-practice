# Goroutines en Go: Notas Básicas

Este directorio contiene ejemplos prácticos sobre **goroutines** en Go, enfocados en concurrencia y sincronización con `sync.WaitGroup`. El código en `main.go` procesa camiones concurrentemente y incluye un experimento para entender los problemas de no sincronizar goroutines.

## Notas sobre Goroutines

- **Definición**: Goroutines permiten escribir software concurrente, ejecutando tareas al mismo tiempo.
- **Características**: Son una versión ligera de un hilo, gestionadas por el runtime de Go.
- **Creación**: Se inician con la palabra clave `go` seguida de una función (ej. `go func()`).
- **Desafíos**:
    - Sin sincronización, el programa puede terminar antes que las goroutines finalicen.
    - Solución: Usar `sync.WaitGroup` para esperar a que un grupo de goroutines termine.
- **Uso**:
    - Ejecutar múltiples tareas que toman tiempo (ej. procesamiento de camiones en `main.go`).
    - Sincronizar resultados de forma ordenada, como con `WaitGroup`.
    - Futuro: Usar canales para comunicar resultados entre goroutines.

## Ejemplo Práctico

El archivo `main.go` muestra cómo procesar camiones concurrentemente usando goroutines y `sync.WaitGroup`. La función `syncProcessTruck` es un experimento para observar el error de no sincronizar goroutines, donde el programa puede terminar antes de que las tareas finalicen.

---
*Última actualización: 19 de mayo de 2025*