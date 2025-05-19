# Punteros en Go: Notas Básicas

Este directorio contiene ejemplos prácticos sobre **punteros** en Go, enfocados en cómo trabajan con referencias a memoria. El código en `main.go` compara métodos por valor (`NormalTruck`) y por referencia (`ComplexTruck`) para mostrar cómo afectan las modificaciones.

## Notas sobre Punteros

- **Definición**: Los punteros son referencias a la dirección de un valor en memoria, declarados con `*T` (ej. `*int`), obtenidos con `&x` (dirección), y desreferenciados con `*p` (valor).
- **Analogía**: Son como cajas en una estantería, donde la computadora sabe dónde está cada variable por su dirección en memoria.
- **Uso**:
    - Comunes en funciones y métodos para modificar structs directamente (ej. incrementar un campo).
    - Sin punteros, se crea una copia del valor que no afecta el original.
- **Alternativa**: Pasar structs por valor y retornar la copia modificada, aunque es menos común y menos eficiente.
- **Eficiencia**: Los punteros evitan copias costosas, especialmente para structs grandes, como los camiones en otros ejercicios.
- **Desafíos**:
    - Los punteros pueden ser `nil`, causando errores como `nil pointer dereference` si no se verifican.
- **Práctica**: Usar punteros en métodos que modifiquen datos para evitar copias innecesarias.

## Ejemplo Práctico

El archivo `main.go` demuestra cómo los punteros modifican structs (`ComplexTruck`) frente a copias por valor (`NormalTruck`). Incluye ejemplos de funciones con punteros, manejo de `nil`, y conceptos básicos de direcciones en memoria (como en `BasicFuncs`).

---
*Última actualización: 19 de mayo de 2025*