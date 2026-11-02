# Canales en Go: Notas Básicas

Este directorio contiene ejemplos prácticos sobre **canales** en Go, enfocados en comunicación entre goroutines. El código en `main.go` procesa camiones concurrentemente, recolectando errores en un canal y usando `sync.WaitGroup` para sincronización.

## Notas sobre Canales

- **Definición**: Los canales permiten compartir información entre goroutines.
- **Propósito**: Facilitan la comunicación segura, trabajando junto con goroutines y `sync.WaitGroup`.
- **Uso**:
    - Crear un canal con `make(chan T)`, usarlo para enviar/recibir datos, y cerrarlo con `close`.
    - Combinar con `select` y `case` para manejar múltiples canales.
- **Desafíos**:
    - **Deadlock**: Ocurre si un canal sin buffer se bloquea (ej. enviar sin receptor) o si no se sincronizan goroutines.
    - **Canales cerrados**: Escribir en un canal cerrado causa pánico.
- **Buenas prácticas**:
    - Usar canales con buffer para evitar deadlocks (ej. `make(chan error, n)`).
    - Iterar canales con `for range` para recolectar datos de forma segura (ej. errores en `main.go`).

## Ejemplo Práctico

El archivo `main.go` muestra cómo procesar camiones concurrentemente, enviando errores a un canal con buffer y recolectándolos en un slice con `for range`. Incluye pruebas para evitar deadlocks y demuestra la sincronización con `WaitGroup`.

---
*Última actualización: 19 de mayo de 2025*