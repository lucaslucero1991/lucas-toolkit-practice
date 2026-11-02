# Contextos en Go: Notas Básicas

Este directorio contiene ejemplos prácticos sobre el paquete **context** en Go, enfocados en manejar concurrencia, cancelación y metadatos. El código en `main.go` procesa camiones concurrentemente usando contextos para timeouts y metadatos.

## Notas sobre Contextos

- **Definición**: El paquete `context` permite llevar señales de cancelación, deadlines o metadatos entre funciones y goroutines.
- **Propósito**: Facilita la coordinación de procesos concurrentes, evitando tareas innecesarias si un servicio falla o se excede un tiempo.
- **Características**:
    - Contextos son **inmutables**. Se crean contextos derivados que apuntan al padre.
    - Seguros para compartir entre múltiples goroutines.
- **Usos principales**:
    - **Pasar información**: Ejemplo, propagar el ID de un usuario logueado usando `context.WithValue`.
    - **Manejo de timeouts**: Configurar `context.WithTimeout` para cancelar operaciones si un servicio demora demasiado.
- **Prácticas**:
    - Usar tipos seguros para claves (ej. `type ContextKey string`) para evitar errores.
    - Crear contextos derivados para añadir valores o configurar cancelación.

## Ejemplo Práctico

El archivo `main.go` muestra cómo procesar camiones concurrentemente usando contextos. Incluye ejemplos de `context.WithValue` para metadatos (ID de usuario) y `context.WithTimeout` para manejar tiempos límite, con sincronización vía `sync.WaitGroup`.

---
*Última actualización: 19 de mayo de 2025*