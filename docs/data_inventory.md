# Inventario de Tratamiento de Datos Personales — Catsplash

> Última actualización: 2026-07-16
> Referencia: LOPDP del Ecuador, Art. 46 (Registro de Actividades de Tratamiento)

---

## 1. Identificación del Responsable

| Campo | Valor |
|-------|-------|
| **Responsable** | [Nombre / organización administradora de la red] |
| **Contacto** | [correo electrónico] |
| **Sistema** | Catsplash — Portal cautivo WiFi de acceso a Internet |
| **Alcance** | Red WiFi local administrada por el Responsable |
| **Finalidad** | Gestión del acceso WiFi: autorización, sesiones, control de tráfico |

---

## 2. Datos Personales Tratados

### 2.1 Datos almacenados en la base de datos

| # | Dato | Campo DB | Tipo | Sensibilidad | Identificador único | Archivo |
|---|------|----------|------|--------------|---------------------|---------|
| 1 | Dirección MAC del dispositivo | `clients.mac` | `TEXT PRIMARY KEY` | **Alto** | Sí — identificador permanente del hardware | `state/schema.sql:2` |
| 2 | Dirección IP asignada | `clients.ip` | `TEXT NOT NULL` | **Medio** | Sí — identificable en red local | `state/schema.sql:3` |
| 3 | Estado de sesión | `clients.state` | `TEXT DEFAULT 'pending'` | Bajo | No | `state/schema.sql:4` |
| 4 | Timestamp de conexión | `clients.connected_at` | `INTEGER` | Bajo | No | `state/schema.sql:5` |
| 5 | Última actividad | `clients.last_seen` | `INTEGER` | Bajo | No | `state/schema.sql:6` |
| 6 | Consentimiento otorgado | `clients.consent_given` | `INTEGER DEFAULT 0` | Bajo | No | `state/schema.sql:7` |
| 7 | Timestamp de consentimiento | `clients.consent_timestamp` | `INTEGER` | Bajo | No | `state/schema.sql:8` |
| 8 | Bytes recibidos | `clients.bytes_in` | `INTEGER DEFAULT 0` | Bajo | No | `state/db.go:34` |
| 9 | Bytes enviados | `clients.bytes_out` | `INTEGER DEFAULT 0` | Bajo | No | `state/db.go:35` |
| 10 | Límite de tráfico | `clients.max_bytes` | `INTEGER DEFAULT 0` | Bajo | No | `state/db.go:36` |
| 11 | Velocidad de bajada | `clients.download_speed` | `TEXT DEFAULT ''` | Bajo | No | `state/db.go:37` |
| 12 | Velocidad de subida | `clients.upload_speed` | `TEXT DEFAULT ''` | Bajo | No | `state/db.go:38` |

### 2.2 Datos procesados pero no almacenados permanentemente

| # | Dato | Dónde se procesa | Retención | Sensibilidad |
|---|------|------------------|-----------|--------------|
| 13 | Nonce de sesión (CSRF) | Cookie `catsplash_nonce` + campo del form | Duración de la petición HTTP | Bajo |
| 14 | Token CSRF admin | Cookie `catsplash_admin_csrf` | Duración de la petición HTTP | Bajo |
| 15 | IP remota del request | `r.RemoteAddr` en handlers | Log (anónimo, sin MAC) | Bajo |

### 2.3 Datos registrados en logs del sistema

| # | Dato | Formato en log | Protección | Archivo |
|---|------|----------------|------------|---------|
| 16 | MAC enmascarada | SHA-256 truncado (8 hex chars) | **Anonimizado** | `server/session.go:48-51` |
| 17 | IP de fallo de auth | `r.RemoteAddr` (sin MAC) | Sin identificador personal | `server/handler_auth.go:19,25` |
| 18 | IP en cookie mismatch | `r.RemoteAddr` | Sin identificador personal | `server/handler_auth.go:25` |

> **Nota**: Los nonces y MACs en texto plano **ya no** se registran en logs desde la implementación de anonimización.

---

## 3. Clasificación por Sensibilidad

| Nivel | Definición LOPDP | Datos aplicables |
|-------|------------------|------------------|
| **Alto** | Datos que permiten identificar directamente a una persona natural o que revelan información sobre ella | Dirección MAC (`clients.mac`) |
| **Medio** | Datos que pueden ser correlacionados con otros para identificar a una persona | Dirección IP (`clients.ip`) |
| **Bajo** | Datos que no permiten identificar a una persona ni revelar información sensible | Estado, timestamps, tráfico, consentimiento |

---

## 4. Base Legal del Tratamiento

| Dato | Base legal | Fundamento |
|------|-----------|------------|
| Dirección MAC | **Consentimiento explícito** (Art. 14 LOPDP) | El titular marca checkbox aceptando la política de privacidad |
| Dirección IP | **Consentimiento explícito** | Mismo consentimiento que la MAC, necesario para la sesión |
| Estado de sesión | **Consentimiento explícito** | Necesario para gestionar el acceso autorizado |
| Timestamps | **Consentimiento explícito** | Necesarios para control de expiración de sesión |
| Tráfico de bytes | **Consentimiento explícito** | Necesario para estadísticas técnicas de la red |
| Consentimiento | **Obligación legal** (Art. 14 LOPDP) | El Responsable está obligado a documentar el consentimiento |

---

## 5. Flujo de Datos

```
[Dispositivo del titular]
    │
    ▼
GET /portal ──→ Cookie: catsplash_nonce (token CSRF)
    │
    ▼
POST /auth ──→ Nonce + consent=true + MAC (del ARP table) + IP
    │
    ├──→ state/db.go ──→ SQLite: INSERT/UPDATE clients (mac, ip, consent, timestamps)
    │
    ├──→ firewall/allow.go ──→ iptables: regla de acceso por MAC
    │
    └──→ Log: MAC enmascarada (SHA-256) + IP sin identificador
```

### Puntos de captura de datos

| Punto | Dato capturado | Mecanismo |
|-------|---------------|-----------|
| `/portal` (GET) | IP del cliente | `r.RemoteAddr` |
| `/auth` (POST) | MAC del cliente | `getMACFromIP()` lee `/proc/net/arp` |
| `/auth` (POST) | Consentimiento | `r.FormValue("consent")` |
| `/admin` (POST) | MAC objetivo | `r.FormValue("mac")` (con `isValidMAC()`) |

### Puntos de eliminación de datos

| Punto | Dato eliminado | Mecanismo |
|-------|---------------|-----------|
| Expiración de sesión | Registro completo del cliente | `state/reaper.go` ejecuta `DELETE FROM clients` |
| Solicitud ARCO+ | Registro completo del cliente | [Pendiente: endpoint `/data-deletion`] |
| Reinicio del sistema | Todos los registros | SQLite se borra si no se persiste el archivo |

---

## 6. Destinatarios de los Datos

| Destinatario | Dato recibido | Base legal | transferencia internacional |
|-------------|---------------|-----------|---------------------------|
| iptables (firewall local) | MAC + IP | Consentimiento | No |
| SQLite (archivo local) | Todos los campos | Consentimiento | No |
| Logs del sistema | MAC enmascarada + IP | Interés legítimo | No |

> **No hay transferencia a terceros.** Todos los datos se procesan y almacenan exclusivamente en el servidor donde opera Catsplash.

---

## 7. Medidas de Seguridad Implementadas

| Medida | Descripción | Estado |
|--------|-------------|--------|
| Hash de contraseña admin | bcrypt con auto-detección de pre-hash | ✅ Implementado |
| Nonce CSRF | Tokens aleatorios en cookies y forms | ✅ Implementado |
| Rate limiting | Sliding window por IP (5 intentos/60s en `/auth`) | ✅ Implementado |
| Sanitización de inputs | Validación MAC con regex estricta | ✅ Implementado |
| Anonimización de logs | SHA-256 truncado para MACs | ✅ Implementado |
| Consentimiento explícito | Checkbox + timestamp en DB | ✅ Implementado |
| Validación constante-time | `subtle.ConstantTimeCompare` para tokens CSRF | ✅ Implementado |
| Cifrado en tránsito | HTTPS | ❌ No implementado (esperado en captive portal) |
| Cifrado en reposo | SQLite cifrado | ❌ No implementado |
| Auditoría de acceso a datos | Registro de accesos a la DB | ❌ No implementado |

---

## 8. Período de Retención

| Dato | Retención | Mecanismo de eliminación |
|------|-----------|------------------------|
| Registro completo del cliente | **Duración de la sesión** (inactividad o cierre) | `reaper.go` elimina registros expirados |
| Consentimiento | Se elimina junto con el registro del cliente | Automático con expiración de sesión |
| Logs del sistema | **No se eliminan automáticamente** | [Pendiente: política de retención de logs] |
| Base de datos SQLite | **Persistente** mientras el archivo exista | Eliminación manual o al desinstalar |

---

## 9. Derechos ARCO+ — Estado de Implementación

| Derecho | Endpoint | Estado |
|---------|----------|--------|
| **Acceso** (ver qué datos tenemos) | `/data-request` | ❌ Pendiente |
| **Rectificación** (corregir datos) | `/data-request` | ❌ Pendiente |
| **Cancelación** (eliminar datos) | `/data-deletion` | ❌ Pendiente |
| **Oposición** (detener tratamiento) | Desconexión de la red | ⚠️ Parcial (desconexión manual) |
| **Portabilidad** (exportar datos) | `/data-export` | ❌ Pendiente |
| **Limitación** (restringir uso) | N/A | ❌ Pendiente |

---

## 10. Revisiones

| Fecha | Revisión | Responsable |
|-------|----------|-------------|
| 2026-07-16 | Inventario inicial | Desarrollador principal |
| [cada 6 meses] | Revisión periódica | [Responsable de cumplimiento] |
