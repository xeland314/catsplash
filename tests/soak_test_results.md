# Informe de Estabilidad y Monitoreo de Carga (Soak Test) - Catsplash

Este documento detalla los resultados obtenidos durante el análisis de estabilidad a largo plazo (_Soak Test_) realizado sobre **Catsplash**. El objetivo de esta prueba es evaluar el comportamiento del _runtime_ de Go, el consumo de memoria física real (RSS), la fragmentación de la base de datos embebida (SQLite) y el determinismo en la inyección/limpieza de reglas de firewall (`iptables`) dentro del espacio de red aislado (`network namespaces`).

## 1. Resumen Ejecutivo

| Métrica Evaluada               | Estado Inicial (Reposito) | Pico de Carga (100 Clientes) | Estado Final (Post-Purga) | Diagnóstico                     |
| ------------------------------ | ------------------------- | ---------------------------- | ------------------------- | ------------------------------- |
| **Memoria Física (RSS)**       | 8.7 MB                    | 24.8 MB                      | 14.7 MB                   | **Estable (Sin Leaks)**         |
| **Reglas NAT (`iptables`)**    | 9                         | 108                          | 9                         | **0% Reglas Huérfanas**         |
| **Reglas Filter (`iptables`)** | 10                        | 208                          | 10                        | **0% Reglas Huérfanas**         |
| **Tamaño de Base de Datos**    | 12.2 KB                   | 12.2 KB                      | 12.2 KB                   | **Reutilización de Páginas OK** |
| **Tasa de Errores HTTP**       | 0%                        | 0%                           | 0%                        | **Robustez Completa**           |

**Resultado Global de la Prueba:** **APROBADO**

---

## 2. Metodología de la Prueba

La prueba se ejecutó de forma continua durante **60 minutos**, dividida en tres fases operativas críticas:

1. **Fase de Aclimatación (Minutos 0-19):** El sistema opera en reposo con un tráfico simulado de mantenimiento (_trickle_) donde un único cliente re-autentica periódicamente para asegurar el flujo de E/S.
2. **Fase de Estrés por Concurrencia Masiva (Minutos 20-34):** Inyección simultánea e instantánea de **100 clientes Wi-Fi paralelos** levantando sockets TCP y forzando escrituras concurrentes de sesión y alteraciones de reglas en el kernel.
3. **Fase de Expiración y Recolección (Minutos 35-59):** Activación del temporizador de limpieza automatizado de Catsplash para purgar sesiones expiradas, destruir hilos remanentes y remover reglas del firewall.

---

## 3. Cronología Analítica del Log (`soak_results.log`)

A continuación se desglosa el comportamiento interno del sistema extraído de los vectores métricos del log:

### A. Gestión de Memoria Residente (RSS) y Colector de Basura (GC)

- **Comportamiento en Carga:** Al irrumpir los 100 clientes en el minuto 20, la memoria física se elevó controladamente a **17.0 MB**, alcanzando su máximo histórico absoluto de **24.8 MB** en el minuto 23.
- **Dinámica del Asignador de Go:** Entre los minutos 21 y 34, la memoria RSS describe un patrón cicloidal (fluctuando entre 24.8 MB, 19.8 MB y 10.8 MB). Esto demuestra que el _Garbage Collector_ de Go se encuentra liberando activamente las estructuras de sockets y buffers HTTP cerrados hacia el pool del _runtime_.
- **Línea Base Post-Carga:** Tras la purga total de clientes (Minuto 36 en adelante), la memoria RSS se estabiliza en una meseta matemática completamente plana de **~14.7 MB** durante más de 20 minutos. El hecho de que la memoria no continúe creciendo linealmente confirma la **ausencia total de fugas de memoria (_memory leaks_)** en el gestor de estado.

### B. Determinismo Lineal y Atómico del Firewall (`iptables`)

El motor de Catsplash demostró un control estricto sobre el Subsistema de Red de Linux Netfilter, reflejado de forma exacta en la relación matemática de inyección de reglas:

$$\text{Reglas NAT} = 8 + \text{Clientes Activos}$$

$$\text{Reglas Filter} = 8 + (2 \times \text{Clientes Activos})$$

- **Verificación Numérica:** \* En reposo con 1 cliente activo (Minuto 19): 9 reglas NAT, 10 reglas Filter.
- Bajo ataque con 100 clientes activos (Minuto 20): 108 reglas NAT ($8 + 100$), 208 reglas Filter ($8 + 200$).

- **Recuperación del Estado del Kernel:** En el minuto 35 se inicia el proceso de expiración por _timeout_. En el minuto 36, tras el desalojo de los 100 clientes concurrentes, las reglas disminuyen de golpe volviendo exactamente a **9 NAT y 10 Filter**. El motor recolector no dejó atrás ninguna regla huérfana, mitigando la degradación del rendimiento de red en el router a largo plazo.

### C. Eficiencia del Motor de Almacenamiento (SQLite)

A lo largo de toda la hora de estrés severo y transacciones concurrentes, el tamaño del archivo de la base de datos se mantuvo estático en **12,288 bytes (12 KB)**.
Esto ratifica que la configuración del driver en modo **WAL (Write-Ahead Logging)** junto con políticas eficientes de indexación permite que SQLite recicle internamente las páginas de base de datos de sesiones viejas en lugar de provocar un crecimiento desmedido en disco, un factor crucial para despliegues en almacenamiento Flash o sistemas embebidos.

---

## 4. Conclusiones Técnicas de Arquitectura

1. **Aislamiento de Recursos Sólido:** El consumo máximo de memoria física del binario compilado no superó el umbral de los **25 MB**, lo que valida la viabilidad arquitectónica de Catsplash para operar en entornos con severas restricciones de hardware (ej. arquitecturas heredadas o plataformas IoT).
2. **Resiliencia ante Ráfagas:** El backend no experimentó _panics_, bloqueos (_deadlocks_) ni corrupción en las tablas de ruteo del kernel ante inundaciones masivas de peticiones concurrentes, respondiendo con éxito (`HTTP 200 OK`) a la totalidad de los flujos inyectados.
3. **Liberación de Recursos Determinista:** El ciclo de expiración de sesiones funciona en sincronía exacta con el subsistema de red, garantizando que el estado del firewall refleje la realidad lógica de la base de datos en tiempo real.
