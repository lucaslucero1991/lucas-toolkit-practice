# Concurrencia en Recursos Compartidos: Notas Básicas

Este directorio contiene ejemplos prácticos sobre concurrencia en recursos compartidos en Go, enfocados en proteger un mapa contra accesos concurrentes. El código en `main.go` y `main_test.go` implementa un gestor de camiones con seguridad para concurrencia.

## Notas sobre Concurrencia en Recursos Compartidos

- **Problema**: Los recursos compartidos, como un mapa, no son seguros en concurrencia, lo que puede causar *race conditions* (condiciones de carrera) si múltiples goroutines acceden al mismo tiempo.
- **Solución**:
    - Usar `sync.RWMutex` (por composición) para proteger el mapa.
    - Usar `Lock` para operaciones de escritura (`AddTruck`, `RemoveTruck`, `UpdateTruckCargo`) y `RLock` para lectura (`GetTruck`), logrando mayor granularidad.
    - Proteger cada método individualmente para evitar accesos concurrentes no sincronizados.
    - Retornar una copia del camión en `GetTruck` (no un puntero) para evitar que goroutines modifiquen la misma dirección de memoria, previniendo *race conditions*.
    - Validar la seguridad con `go test *.go -v -race` para confirmar que no hay condiciones de carrera.
- **RWMutex vs. Mutex**:
    - `sync.Mutex`: Bloquea el recurso para lectura y escritura, permitiendo solo una goroutine a la vez.
    - `sync.RWMutex`: Permite múltiples lecturas simultáneas (`RLock`) pero solo una escritura (`Lock`), siendo más eficiente para operaciones de lectura frecuentes como `GetTruck`.

## Ejemplo Práctico

El archivo `main.go` implementa un `TruckManager` que gestiona un mapa de camiones, protegido con `sync.RWMutex`. `main_test.go` incluye pruebas unitarias y una prueba de concurrencia (`TestConcurrentUpdate`) que valida la seguridad con 100 goroutines actualizando el mismo camión. Ejecuta `go test *.go -v -race` para verificar.

---
*Última actualización: 19 de mayo de 2025*