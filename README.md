# Lucas Toolkit Practice

Repositorio personal para practicar y consolidar conocimientos en programación, con enfoque en Go. Los ejemplos están organizados por tema en carpetas para reforzar conceptos clave a través de la práctica.

## Estructura del repositorio

### 3.CoreConcepts
Fundamentos del lenguaje Go, cada carpeta cubre un concepto:

- `1.Errors` — manejo de errores
- `2.Interfaces` — interfaces
- `3.Tests` — testing
- `4.Pointers` — punteros
- `5.Goroutines` — concurrencia con goroutines
- `6.Context` — paquete context
- `7.Channels` — channels (incluye buffer y worker pool pattern)
- `8.Maps` — maps
- `9.Exercise` — ejercicios integradores

### 4.TCP-to-HTTP
De TCP a un servidor HTTP en Go: un recorrido desde primeros principios sobre lo que ocurre por debajo del paquete `net/http`.

- `encoding` — serialización / deserialización
- `http` — servidor HTTP
- `tcp` — fundamentos de TCP (servidor y cliente)
- `main.go`, `main_test.go`, `moq_test.go` — ejemplo con tests y mocks

Consulta `4.TCP-to-HTTP/README.md` para el detalle de este módulo.
