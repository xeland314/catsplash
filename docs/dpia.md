# Evaluación de Impacto de Protección de Datos (DPIA) — Catsplash

> Última actualización: 2026-07-16
> Referencia: LOPDP del Ecuador, Art. 44 (Evaluación de Impacto)
> Basada en: Hallazgos de seguridad del Punto 5 del análisis LOPDP

---

## 1. Identificación del Evaluador

| Campo | Valor |
|-------|-------|
| **Responsable** | [Nombre / organización administradora de la red] |
| **Contacto** | [correo electrónico] |
| **Evaluador** | Desarrollador principal |
| **Fecha de evaluación** | 2026-07-16 |
| **Alcance** | Sistema Catsplash completo (portal cautivo WiFi) |
| **Ciclo de vida** | Evaluación inicial — primera versión con cumplimiento LOPDP |

---

## 2. Descripción del Tratamiento

### 2.1 Naturaleza del tratamiento

Catsplash es un portal cautivo WiFi que gestiona el acceso a una red inalámbrica. El sistema:

- **Captura** la dirección MAC e IP de dispositivos que se conectan
- **Almacena** estos datos en SQLite junto con estado de sesión y timestamps
- **Autoriza** el acceso ejecutando reglas iptables por MAC
- **Expira** sesiones después de un período de inactividad
- **Registra** logs con MACs anonimizadas (SHA-256 truncado)

### 2.2 Propósito del tratamiento

| Propósito | Base legal | Necesidad |
|-----------|-----------|-----------|
| Identificar dispositivos para autorizar acceso WiFi | Consentimiento explícito | Necesario — sin MAC no hay control de acceso |
| Mantener registro de sesiones activas | Consentimiento explícito | Necesario para expiración y gestión |
| Registrar consentimiento del titular | Obligación legal (LOPDP) | Obligatorio — Art. 14 LOPDP |
| Estadísticas de tráfico de red | Consentimiento explícito | Opcional — se puede deshabilitar |

### 2.3 Alcance territorial

El tratamiento se realiza exclusivamente en el servidor local donde opera Catsplash. No hay transferencias a terceros ni a países extranjeros.

---

## 3. Catalogación de Datos Personales

| # | Dato | Sensibilidad | Permanencia | Ver Inventario |
|---|------|-------------|-------------|----------------|
| 1 | Dirección MAC | **Alto** | Temporal (expira con sesión) | `docs/data_inventory.md` §2.1 |
| 2 | Dirección IP | **Medio** | Temporal (expira con sesión) | `docs/data_inventory.md` §2.1 |
| 3 | Estado de sesión | Bajo | Temporal | `docs/data_inventory.md` §2.1 |
| 4 | Timestamps de actividad | Bajo | Temporal | `docs/data_inventory.md` §2.1 |
| 5 | Tráfico de bytes | Bajo | Temporal | `docs/data_inventory.md` §2.1 |
| 6 | Consentimiento + timestamp | Bajo | Temporal | `docs/data_inventory.md` §2.1 |
| 7 | MAC enmascarada en logs | Bajo | **No se elimina automáticamente** | `docs/data_inventory.md` §2.3 |

---

## 4. Evaluación de Riesgos

### 4.1 Metodología

Se utiliza una matriz de riesgo simplificada:

| Probabilidad ↓ / Impacto → | Bajo | Medio | Alto |
|----------------------------|------|-------|------|
| **Alta** | Riesgo Medio | Riesgo Alto | Riesgo Crítico |
| **Media** | Riesgo Bajo | Riesgo Medio | Riesgo Alto |
| **Baja** | Riesgo Muy Bajo | Riesgo Bajo | Riesgo Medio |

### 4.2 Riesgos identificados

#### R1: Acceso no autorizado a la base de datos

| Campo | Valor |
|-------|-------|
| **Descripción** | Un atacante obtiene acceso al archivo SQLite y extrae MACs, IPs e historial de sesiones |
| **Vectores** | Acceso físico al servidor, explotación de vulnerabilidad en el OS, copia del archivo DB |
| **Datos afectados** | MAC (alto), IP (medio), timestamps, tráfico, consentimiento |
| **Probabilidad** | **Baja** — requiere acceso al servidor |
| **Impacto** | **Alto** — exposición de identificadores de dispositivos de todos los usuarios |
| **Riesgo residual** | **Medio** |
| **Medidas actuales** | Control de acceso administrativo (Basic Auth con bcrypt) |
| **Medidas pendientes** | Cifrado de SQLite en reposo, auditoría de acceso a la DB |

#### R2: Exposición de datos personales en logs

| Campo | Valor |
|-------|-------|
| **Descripción** | MACs o IPs de usuarios aparecen en texto plano en los logs del sistema |
| **Vectores** | Logs accesibles por otros procesos, exportación de logs, integración con sistemas de monitoreo |
| **Datos afectados** | MAC (histórico — ya anonimizada), IP |
| **Probabilidad** | **Media** — los logs se generan continuamente |
| **Impacto** | **Medio** — la MAC anonimizada no es reversable, pero la IP puede correlacionarse |
| **Riesgo residual** | **Bajo** |
| **Medidas actuales** | SHA-256 truncado para MACs en logs, IP sin identificador personal |
| **Medidas pendientes** | Política de retención y eliminación de logs |

#### R3: Suplantación de identidad de dispositivo (spoofing MAC)

| Campo | Valor |
|-------|-------|
| **Descripción** | Un atacante cambia su MAC para suplantar a un dispositivo autorizado |
| **Vectores** | Software de spoofing MAC ( disponible en Linux/macOS/Windows) |
| **Datos afectados** | MAC del dispositivo suplantado |
| **Probabilidad** | **Media** — herramientas de spoofing son accesibles |
| **Impacto** | **Medio** — acceso no autorizado a la red bajo identidad de otro |
| **Riesgo residual** | **Medio** |
| **Medidas actuales** | Nonce CSRF para prevenir automación, pero no previene spoofing manual |
| **Medidas pendientes** | Autenticación de dispositivos (WPA2-Enterprise), monitoreo de MACs duplicadas |

#### R4: Fuerza bruta contra `/auth`

| Campo | Valor |
|-------|-------|
| **Descripción** | Un atacante intenta adivinar nonces CSRF o enviar formularios masivamente |
| **Vectores** | Bots, scripts automatizados, ataques de fuerza bruta |
| **Datos afectados** | Nonces CSRF, posibilidad de acceso no autorizado |
| **Probabilidad** | **Media** — ataques automatizados son comunes |
| **Impacto** | **Medio** — potencial acceso no autorizado |
| **Riesgo residual** | **Bajo** |
| **Medidas actuales** | Rate limiting por IP (5 intentos/60s), nonce CSRF |
| **Medidas pendientes** | Rate limiting más agresivo, monitoreo de patrones |

#### R5: Inyección de comandos en iptables

| Campo | Valor |
|-------|-------|
| **Descripción** | Un atacante inyecta comandos del sistema operativo a través del campo MAC en el formulario admin |
| **Vectores** | Payloads como `; rm -rf /`, `\| cat /etc/passwd`, `` `id` `` |
| **Datos afectados** | Integridad del sistema |
| **Probabilidad** | **Baja** — el endpoint admin requiere autenticación |
| **Impacto** | **Alto** — compromiso completo del servidor |
| **Riesgo residual** | **Muy Bajo** |
| **Medidas actuales** | `isValidMAC()` con regex estricta (`^[0-9A-Fa-f]{2}(:|-)?...`), `exec.Command()` con args separados (no shell), CSRF en admin |
| **Medidas pendientes** | Auditoría de comandos ejecutados |

#### R6: Contraseña admin expuesta

| Campo | Valor |
|-------|-------|
| **Descripción** | La contraseña de administrador se almacena en texto plano en `config.toml` |
| **Vectores** | Acceso al archivo de configuración, lectura por procesos no autorizados |
| **Datos afectados** | Credenciales de administración |
| **Probabilidad** | **Baja** — requiere acceso al servidor |
| **Impacto** | **Alto** — acceso completo al panel de administración |
| **Riesgo residual** | **Bajo** |
| **Medidas actuales** | Contraseña hasheada con bcrypt en carga (`config.go:hashAdminPassword`) |
| **Medidas pendientes** | Eliminar la contraseña en texto plano del config.toml después del hash |

#### R7: Pérdida de datos por fallo del sistema

| Campo | Valor |
|-------|-------|
| **Descripción** | El servidor se reinicia o falla y se pierden todas las sesiones activas |
| **Vectores** | Fallo de energía, crash del proceso, error de disco |
| **Datos afectados** | Todas las sesiones activas |
| **Probabilidad** | **Media** — servidores domésticos son menos confiables |
| **Impacto** | **Bajo** — los usuarios pueden reconectarse, no hay datos permanentes perdidos |
| **Riesgo residual** | **Bajo** |
| **Medidas actuales** | WAL mode en SQLite para resistencia a crashes |
| **Medidas pendientes** | Backup periódico de la DB (opcional) |

#### R8: Falta de consentimiento válido

| Campo | Valor |
|-------|-------|
| **Descripción** | Los usuarios se conectan sin haber otorgado consentimiento válido para el tratamiento de sus datos |
| **Vectores** | Portal sin checkbox, bypass del formulario, consentimiento grabado sin verificación |
| **Datos afectados** | Todos los datos personales tratados sin base legal |
| **Probabilidad** | **Baja** — el checkbox `required` y la validación backend previenen esto |
| **Impacto** | **Alto** — tratamiento sin base legal = violación de LOPDP |
| **Riesgo residual** | **Muy Bajo** |
| **Medidas actuales** | Checkbox `required` en HTML, validación `consent=true` en backend, timestamp de consentimiento en DB |
| **Medidas pendientes** | Verificación periódica de integridad del consentimiento |

---

## 5. Matriz de Riesgos Residuales

| Riesgo | Probabilidad | Impacto | Riesgo Inicial | Medidas | Riesgo Residual |
|--------|-------------|---------|----------------|---------|-----------------|
| R1: Acceso no autorizado a DB | Baja | Alto | **Alto** | Basic Auth bcrypt | **Medio** |
| R2: Datos en logs | Media | Medio | **Medio** | SHA-256 truncado | **Bajo** |
| R3: Spoofing MAC | Media | Medio | **Medio** | Nonce CSRF | **Medio** |
| R4: Fuerza bruta `/auth` | Media | Medio | **Medio** | Rate limiting | **Bajo** |
| R5: Inyección en iptables | Baja | Alto | **Alto** | Regex + args separados | **Muy Bajo** |
| R6: Contraseña admin expuesta | Baja | Alto | **Alto** | bcrypt auto-hash | **Bajo** |
| R7: Pérdida de datos por fallo | Media | Bajo | **Bajo** | WAL mode | **Bajo** |
| R8: Falta de consentimiento | Baja | Alto | **Alto** | Checkbox + backend validation | **Muy Bajo** |

---

## 6. Medidas de Mitigación Implementadas

### 6.1 Medidas técnicas

| Medida | Riesgo mitigado | Implementación | Estado |
|--------|----------------|---------------|--------|
| Hash bcrypt de contraseña admin | R6 | `config/config.go:hashAdminPassword()` | ✅ |
| Nonce CSRF en `/auth` | R3, R4 | `server/handler_portal.go`, `server/handler_auth.go` | ✅ |
| CSRF admin con `subtle.ConstantTimeCompare` | R5 | `server/handler_admin.go:validateAdminCSRF()` | ✅ |
| Rate limiting por IP (5/60s) | R4 | `server/ratelimit.go` | ✅ |
| Validación MAC con regex estricta | R5 | `server/validate.go:isValidMAC()` | ✅ |
| `exec.Command` con args separados | R5 | `firewall/allow.go`, `firewall/block.go` | ✅ |
| Anonimización de MACs en logs | R2 | `server/session.go:maskMAC()` | ✅ |
| Consentimiento explícito + timestamp | R8 | `server/handler_auth.go`, `state/schema.sql` | ✅ |
| WAL mode en SQLite | R7 | `state/db.go:Open()` | ✅ |

### 6.2 Medidas organizativas

| Medida | Riesgo mitigado | Estado |
|--------|----------------|--------|
| Política de privacidad accesible desde el portal | R8 | ✅ `docs/privacy_policy.md` + `/privacy` |
| Datos sintéticos en tests | R1, R2 | ✅ Tests usan MACs de namespace aislado |

---

## 7. Medidas Pendientes

### 7.1 Prioridad alta

| Medida | Riesgo | Descripción | Esfuerzo |
|--------|--------|-------------|----------|
| Cifrado SQLite en reposo | R1 | Usar SQLCipher o cifrar el volumen | Medio |
| Auditoría de acceso a datos | R1 | Log de quién accede a la DB y cuándo | Bajo |
| Eliminar contraseña en texto plano del config | R6 | El config.toml aún tiene la contraseña original | Bajo |
| Endpoint `/data-deletion` | R8 | Solicitud de eliminación de datos (ARCO+) | Medio |
| Endpoint `/data-request` | R8 | Solicitud de acceso a datos (ARCO+) | Medio |

### 7.2 Prioridad media

| Medida | Riesgo | Descripción | Esfuerzo |
|--------|--------|-------------|----------|
| Política de retención de logs | R2 | Definir TTL para logs del sistema | Bajo |
| Monitoreo de MACs duplicadas | R3 | Detectar spoofing básico | Bajo |
| Backup periódico de la DB | R7 | Automatizar copias de seguridad | Bajo |

### 7.3 Prioridad baja

| Medida | Riesgo | Descripción | Esfuerzo |
|--------|--------|-------------|----------|
| WPA2-Enterprise | R3 | Autenticación de dispositivos por certificado | Alto |
| CI/CD con gosec | Todos | Análisis de seguridad estático en cada commit | Medio |
| Pruebas de penetración periódicas | Todos | Evaluación de vulnerabilidades cada 6 meses | Alto |

---

## 8. Consulta a Titulares

| Aspecto | Estado |
|---------|--------|
| Consulta a titulares sobre el tratamiento | ⚠️ Parcial — el checkbox de consentimiento informa sobre el tratamiento |
| Mecanismo de retroalimentación | ❌ Pendiente — no hay canal para que los titulares reporten preocupaciones |
| Canales de contacto | En la política de privacidad (`docs/privacy_policy.md`) |

---

## 9. Conclusiones

### 9.1 Estado general de riesgo

El sistema presenta un **riesgo MEDIO** después de las mitigaciones implementadas. Los riesgos más significativos (R1, R5, R6, R8) han sido reducidos a niveles aceptables mediante controles técnicos.

### 9.2 Cumplimiento LOPDP

| Requisito | Estado |
|-----------|--------|
| Base legal documentada | ✅ Consentimiento explícito |
| Consentimiento registrado | ✅ Checkbox + timestamp en DB |
| Política de privacidad | ✅ Accesible en `/privacy` |
| Derechos ARCO+ | ⚠️ Parcial — falta `/data-request` y `/data-deletion` |
| DPIA | ✅ Este documento |
| Inventario de datos | ✅ `docs/data_inventory.md` |

### 9.3 Revisión periódica

Esta DPIA debe revisarse:
- **Cada 6 meses** en operación normal
- **Inmediatamente** si hay un incidente de seguridad
- **Al cambiar** la funcionalidad del sistema que afecte el tratamiento de datos

---

## 10. Aprobación

| Rol | Nombre | Fecha | Firma |
|-----|--------|-------|-------|
| Evaluador | [desarrollador] | 2026-07-16 | _________ |
| Responsable | [administrador] | _________ | _________ |
