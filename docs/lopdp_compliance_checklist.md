# Checklist de Cumplimiento LOPDP — Catsplash

> Última actualización: 2026-07-16
> Propósito: Checklist de cierre para revisiones periódicas de cumplimiento
> Frecuencia de revisión: Cada 6 meses

---

## Instrucciones

Marcar con `[x]` los puntos que están conformes. Los puntos con `[ ]` requieren acción antes de la próxima revisión.

---

## 1. Consentimiento y notificación

- [x] El portal cautivo muestra la política de privacidad antes del consentimiento
- [x] El checkbox de consentimiento es obligatorio (HTML5 `required` + validación backend)
- [x] El timestamp de consentimiento se almacena en la base de datos
- [x] La política de privacidad está accesible en `/privacy` sin autenticación
- [x] La política identifica al responsable del tratamiento
- [x] La política describe los datos recopilados y su finalidad
- [x] La política describe los derechos ARCO+ y cómo ejercerlos
- [x] La política indica el período de retención de datos

## 2. Derechos ARCO+

- [x] Endpoint `/data-request` devuelve datos del titular en JSON
- [x] Endpoint `/data-deletion` elimina datos del titular
- [x] La identidad se resuelve desde cookie, nunca desde parámetros (anti-IDOR)
- [x] Si la sesión expiró, la respuesta es 404 (no revela si existió)
- [x] La eliminación es irreversible (registro DELETE en DB)
- [x] La eliminación incluye limpieza de reglas de firewall
- [x] Ambos endpoints tienen rate limiting
- [x] Los derechos de rectificación se manejan por eliminación + reconexión

## 3. Retención de datos

- [x] Las sesiones expiran automáticamente por `session_timeout` o `idle_timeout`
- [x] El reaper elimina registros expirados de la DB
- [x] Los datos de sesión se eliminan al expirar (no hay retención indefinida)
- [x] Los nonces de sesión se eliminan con el registro del cliente
- [ ] La retención de logs del sistema está documentada y automatizada
- [ ] El purge de `audit_log` antiguo (>6 meses) está automatizado

## 4. Seguridad de los datos

- [x] Las contraseñas de admin se almacenan con bcrypt
- [x] Las MACs se enmascarán en logs (SHA-256 truncado)
- [x] Los nonces CSRF nunca aparecen en logs
- [x] La validación de MAC usa regex estricta
- [x] Los comandos del sistema usan args separados (sin shell)
- [x] El CSRF admin usa `subtle.ConstantTimeCompare`
- [x] Las cookies de admin usan `SameSite=Strict`
- [ ] SQLite está cifrado en reposo
- [ ] El panel admin se accede por HTTPS
- [ ] Las dependencias se actualizan periódicamente

## 5. Registro de auditoría

- [x] Existe tabla `audit_log` en la base de datos
- [x] Se registra evento en `data_access` al acceder a datos
- [x] Se registra evento en `data_deletion` al eliminar datos
- [x] Se registra evento en `auth_success` al autenticar
- [x] Se registra evento en `auth_denied` al fallar autenticación
- [x] La MAC en audit_log siempre está enmascarada
- [x] El audit_log es visible desde el panel admin
- [ ] El purge de eventos antiguos está automatizado

## 6. Documentación

- [x] `docs/lopdp.md` — Análisis de cumplimiento LOPDP
- [x] `docs/privacy_policy.md` — Política de privacidad
- [x] `docs/data_inventory.md` — Inventario de datos personales
- [x] `docs/dpia.md` — Evaluación de impacto
- [x] `docs/data_retention.md` — Política de retención
- [x] `docs/incident_response.md` — Plan de respuesta a incidentes
- [x] `docs/security_guidelines.md` — Directrices de seguridad
- [x] `docs/compliance_plan.md` — Plan de cumplimiento
- [x] `docs/requirements.md` — Requisitos actualizados con LOPDP
- [x] `CONTRIBUTING.md` — Sección de seguridad para contributors

## 7. Testing

- [x] Tests de consentimiento (4 subtests)
- [x] Tests de IDOR en endpoints ARCO+ (2 tests)
- [x] Tests de auditoría (5 tests state + 2 tests handler)
- [x] Tests de CSRF admin (11 tests)
- [x] Tests de sanitización de MAC (16 payloads)
- [x] Tests de rate limiting (6 tests)
- [x] Tests de anonimización de logs (5 tests)
- [ ] Tests de cifrado de DB (pendiente: Fase 4)
- [ ] Tests de HTTPS (pendiente: Fase 4)

## 8. Incidentes

- [x] Plan de respuesta a incidentes documentado
- [x] Plantilla de notificación a titulares disponible
- [x] Contactos de emergencia definidos
- [ ] Simulacro de respuesta realizado (último: nunca)
- [ ] Registro de incidentes actualizado

---

## Resumen de cumplimiento

| Categoría | Conformes | Total | Porcentaje |
|-----------|-----------|-------|------------|
| Consentimiento y notificación | 8 | 8 | 100% |
| Derechos ARCO+ | 8 | 8 | 100% |
| Retención de datos | 6 | 8 | 75% |
| Seguridad de los datos | 9 | 12 | 75% |
| Registro de auditoría | 7 | 8 | 88% |
| Documentación | 10 | 10 | 100% |
| Testing | 7 | 9 | 78% |
| Incidentes | 3 | 5 | 60% |
| **TOTAL** | **58** | **68** | **85%** |

---

## Próxima revisión

| Campo | Valor |
|-------|-------|
| **Fecha programada** | [2027-01-16] |
| **Responsable** | [Responsable de cumplimiento] |
| **Pendientes para próxima revisión** | Cifrado SQLite, HTTPS admin, automatizar purge de audit_log y logs |
