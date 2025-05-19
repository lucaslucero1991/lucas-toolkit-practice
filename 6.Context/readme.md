# Contextos en Go: Teoría y Conceptos Avanzados

Este documento ofrece una guía completa sobre el paquete `context` en Go, desde fundamentos hasta conceptos avanzados, dirigida a desarrolladores que buscan un entendimiento de nivel senior. Cubre la teoría esencial, patrones avanzados, errores comunes y consideraciones prácticas para sistemas concurrentes y distribuidos.

## ¿Qué es un Contexto?

- **Definición**: El paquete `context` permite gestionar señales de cancelación, deadlines, y metadatos entre funciones y goroutines.
- **Propósito**: Coordina procesos concurrentes, evitando tareas innecesarias si un servicio falla, un cliente se desconecta, o se excede un tiempo límite.
- **Características**:
    - **Inmutables**: Para modificar un contexto, se crea uno derivado que apunta al contexto padre.
    - **Seguros para concurrencia**: Pueden compartirse entre múltiples goroutines sin riesgos.
    - Soportan **cancelación**, **timeouts**, y **valores asociados** (metadatos).

## Usos Principales

1. **Pasar información entre funciones**:
    - Ejemplo: Propagar IDs de usuario, trazas de solicitud (request IDs), o tokens de autenticación.
    - Método: `context.WithValue(ctx, key, value)` con claves tipadas (ej. `type ContextKey string`) para evitar errores.
    - Nota: Usar solo para datos de control, no para parámetros de negocio.

2. **Manejo de cancelación y timeouts**:
    - Ejemplo: Cancelar operaciones si un servicio demora demasiado o un cliente aborta la solicitud.
    - Métodos:
        - `context.WithTimeout(ctx, duration)`: Establece un límite de tiempo.
        - `context.WithCancel(ctx)`: Permite cancelación manual.
    - Propagación: Los contextos derivados transmiten la señal de cancelación a todas las goroutines.

## Patrones Avanzados

1. **Cancelación en cascada**:
    - Crear un contexto raíz con `context.WithCancel` y derivar contextos hijos para subprocesos.
    - Cancelar el contexto raíz propaga la señal a todos los hijos, útil en pipelines o flujos de trabajo complejos.
    - Ejemplo: Un servidor HTTP que cancela todas las goroutines de una solicitud cuando el cliente cierra la conexión.

2. **Contextos anidados**:
    - Usar múltiples niveles de contextos para diferentes scopes (ej. uno para la solicitud HTTP, otro para una consulta a base de datos).
    - Cuidado: Evitar cadenas largas de contextos derivados, ya que pueden dificultar el seguimiento.

3. **Propagación en sistemas distribuidos**:
    - En microservicios o gRPC, propagar metadatos (como request IDs o trazas) a través de contextos.
    - Ejemplo: Usar `context.WithValue` para incluir un trace ID que se pasa entre servicios, facilitando el rastreo de solicitudes.

4. **Contextos con deadlines absolutas**:
    - Usar `context.WithDeadline(ctx, time.Time)` para fijar un tiempo de finalización absoluto, útil en sistemas con horarios estrictos (ej. tareas programadas).

## Mejores Prácticas

- **Siempre liberar recursos**: Llama `defer cancel()` en funciones que crean contextos derivados para evitar fugas de recursos.
- **Minimizar uso de `context.WithValue`**: Úsalo solo para datos esenciales (ej. IDs de trazas). Evita usarlo como reemplazo de parámetros.
- **Propagar el contexto**: Pasa el contexto a todas las funciones y goroutines dependientes, incluso si no lo usan directamente (futuras expansiones).
- **Detectar cancelaciones**: Usa `select { case <-ctx.Done(): ... }` para manejar cancelaciones o timeouts de forma reactiva.
- **Evitar context.TODO en producción**: Usa `context.Background()` como contexto raíz en servidores o `context.TODO()` solo en código temporal.

## Errores Comunes y Anti-Patrones

1. **Abuso de `context.WithValue`**:
    - Problema: Usar contextos para pasar datos de negocio (ej. structs complejos) en lugar de parámetros explícitos.
    - Solución: Mantener `WithValue` para metadatos de control (trace IDs, auth tokens).

2. **No liberar contextos**:
    - Problema: Olvidar `defer cancel()` puede causar fugas de recursos, especialmente en bucles o goroutines longevas.
    - Solución: Siempre incluir `defer cancel()` al crear un contexto con `WithCancel`, `WithTimeout`, o `WithDeadline`.

3. **Ignorar `ctx.Done()`**:
    - Problema: No verificar cancelaciones puede llevar a ejecutar tareas innecesarias tras un timeout o cancelación.
    - Solución: Integrar `select` con `ctx.Done()` en operaciones bloqueantes (ej. I/O, canales).

4. **Contextos no propagados**:
    - Problema: No pasar el contexto a funciones hijas puede romper la cancelación en cascada.
    - Solución: Hacer del contexto un parámetro obligatorio en funciones que puedan ser canceladas.

## Consideraciones de Rendimiento

- **Overhead de contextos**: Crear y derivar contextos tiene un costo bajo, pero en sistemas de alta concurrencia (miles de goroutines por segundo), el uso excesivo de contextos derivados puede acumularse.
    - Mitigación: Reutilizar contextos cuando sea posible y evitar cadenas profundas de derivación.
- **Cancelación activa**: Verificar `ctx.Done()` en bucles apretados puede introducir latencia. Usar con moderación o en puntos estratégicos (ej. antes de operaciones costosas).
- **Goroutines huérfanas**: Sin `WaitGroup` o manejo adecuado, goroutines que ignoran cancelaciones pueden seguir ejecutándose, consumiendo recursos.
    - Mitigación: Combinar contextos con `sync.WaitGroup` para garantizar sincronización.

## Contextos en Sistemas Distribuidos

- **Propagación de metadatos**: En arquitecturas de microservicios, los contextos son clave para propagar información como trace IDs, tokens de autenticación, o niveles de prioridad.
- **Integración con gRPC**: gRPC usa contextos nativamente para manejar cancelaciones y metadatos. Asegúrate de pasar el contexto del cliente al servidor y respetar `ctx.Done()`.
- **Trazabilidad**: Usa bibliotecas como OpenTelemetry para estandarizar la propagación de trazas a través de contextos, facilitando la depuración en sistemas distribuidos.

## Cuándo NO Usar Contextos

- **Operaciones no cancelables**: Si una tarea no puede o no debe ser interrumpida (ej. escritura crítica a disco), evita usar contextos para cancelación.
- **Datos de negocio complejos**: No uses `context.WithValue` para estructuras grandes o parámetros que deberían ser explícitos.
- **Alternativas**: Para tareas simples sin concurrencia, considera canales o mecanismos más ligeros en lugar de contextos.

## Herramientas y Debugging

- **Logging de contextos**: Añade trace IDs en `context.WithValue` para correlacionar logs en sistemas distribuidos.
- **Inspección en producción**: Usa herramientas como `context.String()` o bibliotecas de trazas (OpenTelemetry, Jaeger) para depurar el estado de un contexto.
- **Testing**: Usa `context.WithCancel` en pruebas unitarias para simular cancelaciones o timeouts y verificar el comportamiento.
- **Bibliotecas útiles**:
    - `ctxlog`: Para asociar logs con contextos.
    - `contextutil`: Para utilidades avanzadas como contextos con retry.

## Resumen Rápido

- **Crear**: Usa `context.Background()` (servidores) o `context.TODO()` (temporal).
- **Metadatos**: `context.WithValue(ctx, key, value)` solo para datos de control.
- **Timeouts/Deadlines**: `context.WithTimeout` o `WithDeadline` con `defer cancel()`.
- **Cancelación**: `context.WithCancel` y `select { case <-ctx.Done(): ... }`.
- **Propagación**: Pasa el contexto a todas las funciones/goroutines.
- **Sincronización**: Usa `sync.WaitGroup` para esperar goroutines.
- **Avanzado**: Maneja cancelación en cascada, propagación en microservicios, y optimiza para rendimiento.

## Ejemplo Práctico

Consulta el archivo `main.go` para un ejemplo funcional que procesa camiones concurrentemente, usando contextos para metadatos (ID de usuario) y timeouts. El código ilustra cancelación, propagación, y sincronización.

---
*Última actualización: 19 de mayo de 2025*