# Política de Retención de Datos — Catsplash

> Última actualización: 2026-07-16
> Referencia: LOPDP del Ecuador, Art. 13 (Principio de conservación / minimización)
> Vinculado a: `docs/data_inventory.md` §8, `state/reaper.go`

---

## 1. Principio

Solo se conservan los datos personales durante el tiempo estrictamente necesario para cumplir la finalidad para la que fueron recopilados. Una vez expirada la sesión, los datos se eliminan automáticamente sin intervención manual.

---

## 2. Ciclo de vida de cada dato

### 2.1 Datos de la tabla `clients`

| Dato | Campo DB | Retención | Mecanismo de eliminación | Trigger |
|------|----------|-----------|-------------------------|---------|
| Dirección MAC | `mac` | **Duración de la sesión** (inactividad o cierre) | `state/reaper.go` ejecuta `DELETE FROM clients WHERE ...` | Temporizador periódico (reaper) |
| Dirección IP | `ip` | Igual que MAC | Mismo DELETE que MAC | Igual que MAC |
| Estado de sesión | `state` | Igual que MAC | Igual que MAC | Igual que MAC |
| Timestamp de conexión | `connected_at` | Igual que MAC | Igual que MAC | Igual que MAC |
| Última actividad | `last_seen` | Igual que MAC | Igual que MAC | Igual que MAC |
| Bytes recibidos/envidados | `bytes_in`, `bytes_out` | Igual que MAC | Igual que MAC | Igual que MAC |
| Límite de tráfico | `max_bytes` | Igual que MAC | Igual que MAC | Igual que MAC |
| Velocidades | `download_speed`, `upload_speed` | Igual que MAC | Igual que MAC | Igual que MAC |
| Consentimiento | `consent_given`, `consent_timestamp` | Igual que MAC | Igual que MAC | Igual que MAC |
| Token de sesión | `session_token` | Igual que MAC | Igual que MAC | Igual que MAC |

### 2.2 Datos de la tabla `audit_log`

| Dato | Campo DB | Retención | Mecanismo de eliminación | Trigger |
|------|----------|-----------|-------------------------|---------|
| Evento de auditoría | Todas las columnas | **6 meses** | `state.PurgeOldAuditEvents(6 * 30 * 24 * time.Hour)` | Tarea periódica o manual |

### 2.3 Datos en cookies del navegador

| Cookie | Retención | Mecanismo de eliminación |
|--------|-----------|-------------------------|
| `catsplash_nonce` | Sesión del navegador (se borra al cerrar) | El navegador la elimina automáticamente |
| `catsplash_admin_csrf` | Sesión del navegador | El navegador la elimina automáticamente |

### 2.4 Datos en logs del sistema

| Dato | Retención | Mecanismo de eliminación |
|------|-----------|-------------------------|
| MAC enmascarada (SHA-256) | **No se eliminan automáticamente** | Requiere política manual del administrador |
| IP sin identificador personal | No se eliminan automáticamente | Requiere política manual del administrador |

> **Nota**: Los logs del sistema no contienen datos personales identificables (la MAC está enmascarada con SHA-256 truncado). Sin embargo, se recomienda implementar una retención periódica de logs por higiene operativa.

---

## 3. Configuración de expiración de sesiones

Los parámetros de retención se configuran en `config.toml`:

```toml
session_timeout = 3600   # Tiempo máximo de sesión en segundos (1 hora por defecto)
idle_timeout    = 1800   # Tiempo máximo de inactividad en segundos (30 min por defecto)
```

| Parámetro | Descripción | Valor por defecto | Ejemplo |
|-----------|-------------|-------------------|---------|
| `session_timeout` | Tiempo absoluto máximo desde la conexión | `0` (sin límite) | `3600` (1 hora) |
| `idle_timeout` | Tiempo máximo sin actividad detectada | `0` (sin límite) | `1800` (30 minutos) |

### Comportamiento del reaper

El reaper (`state/reaper.go`) se ejecuta periódicamente y elimina clientes que cumplan al menos una de estas condiciones:

1. **Expiración por sesión**: `now - connected_at >= session_timeout`
2. **Expiración por inactividad**: `now - last_seen >= idle_timeout`
3. **Estado pendiente**: clientes en estado `pending` que nunca completaron la autenticación

```go
// state/reaper.go — lógica simplificada
DELETE FROM clients
WHERE (state = 'pending' AND last_seen < ?)          -- pendientes antiguos
   OR (session_timeout > 0 AND connected_at + session_timeout < ?)  -- sesión expirada
   OR (idle_timeout > 0 AND last_seen + idle_timeout < ?)           -- inactividad
```

---

## 4. Eliminación por solicitud ARCO+

Además de la expiración automática, el usuario puede solicitar la eliminación inmediata de sus datos a través del endpoint `/data-deletion`:

| Aspecto | Detalle |
|---------|---------|
| **Endpoint** | `POST /data-deletion` |
| **Identidad** | Resuelta desde cookie `catsplash_nonce` (no parámetro) |
| **Qué se borra** | Registro completo del cliente en `clients` + reglas iptables |
| **Auditoría** | Se registra evento `data_deletion` en `audit_log` ANTES de borrar |
| **Irreversible** | Una vez borrado, no hay forma de recuperar los datos |
| **Sesión expirada** | Si la sesión ya expiró, el reaper ya borró los datos → 404 |

Referencia: `server/handler_data.go:handleDataDeletion`

---

## 5. Retención de auditoría

Los eventos de auditoría (`audit_log`) se conservan por un período de **6 meses** para fines de cumplimiento LOPDP y trazabilidad.

### Purge periódico

```go
// Ejemplo de llamada al purge (puede integrarse en el reaper o una tarea cron)
deleted, err := db.PurgeOldAuditEvents(6 * 30 * 24 * time.Hour)
```

### Qué se purga

- Eventos con `timestamp` mayor a 6 meses de antigüedad
- No se purgean eventos de `data_deletion` (se conservan indefinidamente para trazabilidad)

### Recomendación de integración

Integrar `PurgeOldAuditEvents` en el ciclo del reaper o en una tarea cron del sistema:

```bash
# Ejemplo: ejecutar purge cada mes vía cron
0 0 1 * * /usr/local/bin/catsctl audit-purge
```

---

## 6. Datos que NUNCA se eliminan

| Dato | Razón |
|------|-------|
| Eventos de auditoría de eliminación (`data_deletion`) | Trazabilidad LOPDP — evidencia de que se procesó la solicitud |
| Configuración del sistema (`config.toml`) | No contiene datos personales |
| Schema de la base de datos | Estructura, no datos personales |

---

## 7. Resumen visual del ciclo de vida

```
[Usuario conecta WiFi]
    │
    ▼
GET /portal → cookie catsplash_nonce creada
    │
    ▼
POST /auth → MAC + IP + consent guardados en DB
    │          session_token = nonce
    │
    ├──→ [Sesión activa] ──→ Datos conservados
    │
    ├──→ [POST /data-deletion] ──→ Eliminación inmediata
    │                              → audit_log: data_deletion
    │                              → firewall: BlockClient
    │                              → DB: DELETE FROM clients
    │
    ├──→ [Inactividad > idle_timeout] ──→ reaper elimina registro
    │
    └──→ [Tiempo > session_timeout] ──→ reaper elimina registro
                                        → Datos perdidos permanentemente

[audit_log]
    │
    └──→ [> 6 meses] ──→ PurgeOldAuditEvents elimina eventos antiguos
```

---

## 8. Responsabilidades

| Rol | Responsabilidad |
|-----|----------------|
| **Administrador de red** | Configurar `session_timeout` e `idle_timeout` adecuadamente |
| **Administrador de red** | Ejecutar purge de audit_log periódicamente |
| **Desarrollador** | Mantener el reaper funcionando correctamente |
| **Desarrollador** | No agregar campos de datos personales sin evaluar retención |

---

## 9. Revisión

| Fecha | Revisión | Responsable |
|-------|----------|-------------|
| 2026-07-16 | Política inicial | Desarrollador principal |
| [cada 6 meses] | Revisión periódica | [Responsable de cumplimiento] |
