# Contextos en Go: Teoría y Conceptos Clave

Este documento resume los conceptos fundamentales del paquete `context` en Go, utilizado para manejar concurrencia, cancelación y metadatos en aplicaciones.

## ¿Qué es un Contexto?

- **Definición**: El paquete `context` proporciona una forma de gestionar señales de cancelación, deadlines (tiempos límite) y metadatos entre funciones y goroutines.
- **Propósito**: Permite coordinar procesos concurrentes, evitando que tareas sigan ejecutándose innecesariamente si un servicio falla o se excede un tiempo límite.
- **Características**:
    - Los contextos son **inmutables**. Para modificar un contexto, se crea uno derivado que apunta al contexto padre.
    - Son seguros para compartir entre múltiples goroutines.
    - Soportan **cancelación**, **timeouts** y **valores asociados** (metadatos).

## Usos Principales

1. **Pasar información entre funciones**:
    - Ejemplo: Propagar el ID de un usuario logueado o datos de sesión.
    - Método: Usar `context.WithValue(ctx, key, value)` para asociar un par clave-valor.
    - Nota: Se recomienda usar tipos personalizados para las claves (ej. `type ContextKey string`) para evitar errores de tipeo.

2. **Manejo de cancelación y timeouts**:
    - Ejemplo: Cancelar una operación si un servicio demora demasiado o falla.
    - Métodos:
        - `context.WithTimeout(ctx, duration)`: Establece un tiempo límite.
        - `context.WithCancel(ctx)`: Permite cancelación manual.
    - Los contextos derivados propagan la señal de cancelación a todas las goroutines que los usan.

## Mejores Prácticas

- **Siempre liberar recursos**: Usa `defer cancel()` en funciones que crean contextos derivados para evitar fugas de recursos.
- **Evitar abuso de `context.WithValue`**: Úsalo solo para datos esenciales (como IDs de usuario) y no para pasar parámetros complejos.
- **Propagar el contexto**: Pasa el contexto a todas las funciones y goroutines que dependan de él.
- **Detectar cancelaciones**: Usa `select` con `ctx.Done()` para manejar cancelaciones o timeouts de forma elegante.

## Resumen Rápido

- **Crear un contexto**: Usa `context.Background()` o `context.TODO()` como base.
- **Añadir metadatos**: `context.WithValue(ctx, key, value)` con claves tipadas.
- **Configurar timeouts**: `context.WithTimeout(ctx, duration)` y siempre `defer cancel()`.
- **Detectar cancelación**: Usa `select { case <-ctx.Done(): ... }`.
- **Propagar contexto**: Pasa el contexto a todas las funciones y goroutines.
- **Sincronización**: Usa `sync.WaitGroup` para esperar goroutines en procesos concurrentes.

## Ejemplo Práctico

Consulta el archivo `main.go` en este proyecto para ver un ejemplo de contextos en acción, donde se procesan camiones concurrentemente con soporte para metadatos (ID de usuario) y timeouts.

---
*Última actualización: 19 de mayo de 2025*