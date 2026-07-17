# Plan de Cumplimiento LOPDP — Catsplash

> Última actualización: 2026-07-16
> Referencia: Ley Orgánica de Protección de Datos Personales del Ecuador
> Estado: **En progreso** — Fase 2 completada

---

## 1. Objetivo

Documentar el plan de cumplimiento de la LOPDP para el sistema Catsplash, identificando las actividades realizadas, pendientes y futuras.

---

## 2. Alcance

| Componente | Alcance LOPDP |
|------------|---------------|
| Portal cautivo (`/portal`) | Captura de MAC + IP, consentimiento explícito |
| Autenticación (`/auth`) | Registro de sesión, consentimiento |
| Panel admin (`/admin`) | Gestión de sesiones, auditoría |
| ARCO+ (`/data-request`, `/data-deletion`) | Derechos de acceso y cancelación |
| Base de datos SQLite | Almacenamiento de datos personales |
| Logs del sistema | Anonimización de MACs |
| Firewall (iptables) | Control de acceso por MAC |

---

## 3. Fases de cumplimiento

### Fase 0: Fundamentos de seguridad ✅

| Actividad | Estado | Archivo |
|-----------|--------|---------|
| Análisis LOPDP de 10 puntos | ✅ | `docs/lopdp.md` |
| Política de privacidad | ✅ | `docs/privacy_policy.md` |
| Hash bcrypt de contraseña admin | ✅ | `config/config.go` |
| Anonimización de MACs en logs | ✅ | `server/session.go` |
| Rate limiting en `/auth` | ✅ | `server/ratelimit.go` |
| Sanitización de inputs (MAC) | ✅ | `server/validate.go` |
| CSRF protection | ✅ | `server/handler_portal.go`, `server/handler_admin.go` |

### Fase 1: Consentimiento y documentación ✅

| Actividad | Estado | Archivo |
|-----------|--------|---------|
| Checkbox de consentimiento | ✅ | `server/templates/portal.html` |
| Validación backend de consentimiento | ✅ | `server/handler_auth.go` |
| Timestamp de consentimiento en DB | ✅ | `state/schema.sql` |
| Inventario de datos personales | ✅ | `docs/data_inventory.md` |
| Evaluación de impacto (DPIA) | ✅ | `docs/dpia.md` |
| Política de retención de datos | ✅ | `docs/data_retention.md` |

### Fase 2: Derechos ARCO+ ✅

| Actividad | Estado | Archivo |
|-----------|--------|---------|
| Endpoint `/data-request` (acceso/portabilidad) | ✅ | `server/handler_data.go` |
| Endpoint `/data-deletion` (cancelación) | ✅ | `server/handler_data.go` |
| Identidad por cookie (anti-IDOR) | ✅ | `server/handler_data.go:resolveSession` |
| Rate limiting en endpoints ARCO+ | ✅ | `server/server.go` |
| Registro de auditoría LOPDP | ✅ | `state/audit.go` |
| Visor de auditoría en admin | ✅ | `server/templates/admin.html` |
| Tests de IDOR | ✅ | `server/handler_data_test.go` |
| Tests de auditoría | ✅ | `state/audit_test.go`, `server/handler_data_test.go` |

### Fase 3: Documentación y operaciones ✅

| Actividad | Estado | Archivo |
|-----------|--------|---------|
| Plan de retención de datos | ✅ | `docs/data_retention.md` |
| Plan de respuesta a incidentes | ✅ | `docs/incident_response.md` |
| Directrices de seguridad | ✅ | `docs/security_guidelines.md` |
| Checklist de cumplimiento | ✅ | `docs/lopdp_compliance_checklist.md` |
| Sección de seguridad en CONTRIBUTING.md | ✅ | `CONTRIBUTING.md` |
| Requisitos LOPDP actualizados | ✅ | `docs/requirements.md` |

---

## 4. Requisitos LOPDP y estado de cumplimiento

| # | Requisito LOPDP | Artículo | Estado | Evidencia |
|---|-----------------|----------|--------|-----------|
| 1 | Consentimiento explícito y libre | Art. 14 | ✅ | Checkbox + validación backend + timestamp |
| 2 | Información al titular | Art. 15 | ✅ | Política de privacidad en `/privacy` |
| 3 | Derecho de acceso | Art. 26 | ✅ | `GET /data-request` → JSON |
| 4 | Derecho de rectificación | Art. 27 | ⚠️ | No aplica (datos se eliminan, no se modifican) |
| 5 | Derecho de eliminación | Art. 28 | ✅ | `POST /data-deletion` → DELETE |
| 6 | Derecho de oposición | Art. 29 | ✅ | Desconexión = oposición al tratamiento |
| 7 | Derecho de portabilidad | Art. 30 | ✅ | `GET /data-request` → JSON exportable |
| 8 | Registro de actividades | Art. 46 | ✅ | `docs/data_inventory.md` |
| 9 | Evaluación de impacto | Art. 44 | ✅ | `docs/dpia.md` |
| 10 | Medidas de seguridad | Art. 50 | ✅ | `docs/security_guidelines.md` + controles técnicos |
| 11 | Notificación de incidentes | Art. 50 | ✅ | `docs/incident_response.md` |
| 12 | Retención mínima | Art. 13 | ✅ | `docs/data_retention.md` + reaper automático |
| 13 | Registro de auditoría | — | ✅ | `state/audit.go` + tabla `audit_log` |
| 14 | Transferencias internacionales | Art. 39 | ✅ | No hay transferencias (todo local) |

---

## 5. Próximos pasos

### Corto plazo (1-3 meses)

| # | Actividad | Prioridad |
|---|-----------|-----------|
| 1 | Cifrado SQLite en reposo (SQLCipher) | Alta |
| 2 | HTTPS en panel admin | Alta |
| 3 | Endpoint `/data-export` (formato JSON/CSV completo) | Media |
| 4 | Política de retención de logs del sistema | Media |
| 5 | Backup automatizado de la DB | Media |

### Mediano plazo (3-6 meses)

| # | Actividad | Prioridad |
|---|-----------|-----------|
| 6 | Endpoint de portabilidad en formato estándar | Media |
| 7 | Monitoreo de MACs duplicadas (anti-spoofing) | Media |
| 8 | CI/CD con análisis de seguridad estático (gosec) | Baja |
| 9 | Simulacro de respuesta a incidentes | Baja |
| 10 | Pruebas de penetración externas | Baja |

### Largo plazo (6-12 meses)

| # | Actividad | Prioridad |
|---|-----------|-----------|
| 11 | WPA2-Enterprise para autenticación por certificado | Baja |
| 12 | Integración con SIEM para monitoreo centralizado | Baja |
| 13 | Certificación o attestación de cumplimiento | Baja |

---

## 6. Métricas de cumplimiento

| Métrica | Valor actual | Meta |
|---------|-------------|------|
| Tests passing | 40+ | 100% |
| Cobertura de endpoints ARCO+ | 100% | 100% |
| Eventos de auditoría registrados | Automático | 100% de accesos |
| Tiempo de retención de datos | Sesión activa | ≤ 1 hora |
| Documentación LOPDP | 8 documentos | 10 controles |

---

## 7. Revisiones

| Fecha | Revisión | Responsable |
|-------|----------|-------------|
| 2026-07-16 | Plan inicial — Fases 0-3 | Desarrollador principal |
| [cada 6 meses] | Revisión de cumplimiento | [Responsable de cumplimiento] |
| [anualmente] | Auditoría externa | [Auditor] |
