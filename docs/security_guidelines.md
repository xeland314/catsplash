# Directrices de Seguridad — Catsplash

> Última actualización: 2026-07-16
> Referencia: LOPDP del Ecuador, Art. 50 (Medidas de seguridad)
> Vinculado a: `docs/dpia.md` §6, `docs/data_inventory.md` §7

---

## 1. Principios de seguridad

1. **Defensa en profundidad**: Múltiples capas de protección, nunca depender de un solo control.
2. **Mínimo privilegio**: Cada componente solo accede a lo estrictamente necesario.
3. **Seguridad por diseño**: La seguridad se integra desde el diseño, no se añade después.
4. **No confiar en el cliente**: Todos los inputs del usuario se validan en el servidor.

---

## 2. Controles implementados

### 2.1 Autenticación

| Control | Implementación | Archivo |
|---------|---------------|---------|
| Contraseña admin hasheada | bcrypt con auto-detección de pre-hash | `config/config.go:hashAdminPassword()` |
| Rate limiting en `/auth` | Sliding window: 5 intentos / 60s por IP | `server/ratelimit.go` |
| CSRF en portal | Nonce aleatorio en cookie + campo hidden | `server/handler_portal.go`, `server/handler_auth.go` |
| CSRF admin | Nonce en cookie + `subtle.ConstantTimeCompare` | `server/handler_admin.go:validateAdminCSRF` |
| Consentimiento explícito | Checkbox `required` + validación backend | `server/handler_auth.go:30-35` |

### 2.2 Sanitización de inputs

| Control | Implementación | Archivo |
|---------|---------------|---------|
| Validación MAC | Regex estricta `^[0-9A-Fa-f]{2}(:|-)?...` | `server/validate.go:isValidMAC()` |
| Args separados en exec | `exec.Command("iptables", ...)` sin shell | `firewall/allow.go`, `firewall/block.go` |
| Sin interpolación de strings | Nunca se construyen comandos con `fmt.Sprintf` | Todo el código firewall |

### 2.3 Protección de datos

| Control | Implementación | Archivo |
|---------|---------------|---------|
| Anonimización de logs | SHA-256 truncado para MACs | `server/session.go:maskMAC()` |
| Nonces no registrados | Logs nunca contienen valores CSRF | `server/handler_auth.go` |
| Identidad por cookie | ARCO+ endpoints resuelven MAC desde cookie, no parámetro | `server/handler_data.go:resolveSession` |
| SameSite=Strict | Cookies de admin con SameSite estricto | `server/handler_admin.go:140-146` |

### 2.4 Infraestructura

| Control | Implementación | Archivo |
|---------|---------------|---------|
| SQLite WAL mode | Resiste crashes sin corromperse | `state/db.go:Open()` |
| Firewall limpieza | `signal.Notify` limpia iptables al cerrar | `firewall/firewall.go` |
| Sin datos en texto plano | Contraseña hasheada, MACs enmascaradas | `config/config.go`, `server/session.go` |

---

## 3. Controles pendientes

| Control | Prioridad | Esfuerzo | Descripción |
|---------|-----------|----------|-------------|
| Cifrado SQLite en reposo | Alta | Medio | Usar SQLCipher o cifrar el volumen |
| HTTPS en panel admin | Alta | Bajo | Certificado auto-firmado o Let's Encrypt |
| Auditoría de acceso a DB | Media | Bajo | Log de queries ejecutadas contra la DB |
| Backup cifrado | Media | Bajo | Copia automática del archivo SQLite |
| Monitoreo de MACs duplicadas | Media | Bajo | Detectar spoofing básico |
| WPA2-Enterprise | Baja | Alto | Autenticación por certificado de dispositivo |
| CI/CD con gosec | Baja | Medio | Análisis de seguridad estático en cada commit |

---

## 4. Directrices para contributors

### 4.1 Reglas de código seguro

| Regla | Ejemplo incorrecto | Ejemplo correcto |
|-------|-------------------|------------------|
| Nunca interpolar en exec | `exec.Command("sh", "-c", "iptables -A "+mac)` | `exec.Command("iptables", "-A", mac)` |
| Siempre validar inputs | `mac := r.FormValue("mac")` | `mac := r.FormValue("mac"); if !isValidMAC(mac) { ... }` |
| No loguear datos sensibles | `log.Printf("MAC: %s", mac)` | `log.Printf("MAC: %s", maskMAC(mac))` |
| Usar constant-time compare | `if cookie == formToken` | `subtle.ConstantTimeCompare([]byte(cookie), []byte(formToken))` |
| Hashear credenciales | `config.AdminPass = "mipassword"` | `config.AdminPass = hashAdminPassword("mipassword")` |

### 4.2 Antes de submitir un PR

- [ ] `go vet ./...` no reporta errores
- [ ] `go test ./...` pasa todos los tests
- [ ] No se añaden campos de datos personales sin evaluar retención
- [ ] No se modifican endpoints ARCO+ sin actualizar tests de IDOR
- [ ] No se cambian controles CSRF sin actualizar tests de CSRF
- [ ] Se añaden tests para funcionalidad nueva

### 4.3 Revisión de seguridad en PRs

Los reviewers deben verificar:

1. **No hay interpolación de strings en comandos del sistema**
2. **Los inputs se validan antes de usarlos**
3. **Los logs no contienen datos personales en texto plano**
4. **Los nuevos endpoints tienen rate limiting si procesan datos personales**
5. **Los cambios en templates no rompen la protección CSRF**
6. **No se exponen secretos o credenciales en el código**

---

## 5. Directrices para administradores

### 5.1 Configuración segura

```toml
# config.toml — configuración recomendada
portal_port = 8080           # Puerto del portal cautivo
admin_port  = 8081           # Puerto del panel admin (en producción, detrás de proxy)
admin_user  = "admin"        # Usuario admin (no usar nombres comunes)
admin_pass  = "<bcrypt hash>" # SIEMPRE usar hash, nunca texto plano
session_timeout = 3600       # 1 hora máximo de sesión
idle_timeout    = 1800       # 30 min máximo de inactividad
```

### 5.2 Operaciones de seguridad periódicas

| Tarea | Frecuencia | Comando |
|-------|------------|---------|
| Cambiar contraseña admin | Cada 3 meses | `catsctl setpass <nueva>` |
| Revisar logs de auditoría | Semanal | `catsctl audit-list` |
| Verificar integridad de archivos | Mensual | `sha256sum -c checksums.sha256` |
| Actualizar dependencias | Mensual | `go get -u ./... && go mod tidy` |
| Purge de audit_log antiguo | Mensual | `catsctl audit-purge` |
| Backup de la DB | Semanal | `cp /var/lib/catsplash/state.db /backup/` |
| Verificar reglas iptables | Semanal | `iptables -L -n -v` |

### 5.3 Monitoreo

Señales que requieren investigación inmediata:

| Señal | Posible causa |
|-------|---------------|
| Picos de requests a `/auth` desde una IP | Fuerza bruta |
| IPs desconocidas en logs de admin | Acceso no autorizado |
| Reglas iptables no reconocidas | Manipulación del firewall |
| Archivo `state.db` modificado externamente | Acceso no autorizado a la DB |
| Consumo inusual de CPU/red | DoS o minero |

---

## 6. Seguridad de la red

| Aspecto | Recomendación |
|---------|--------------|
| Separación de interfaces | WiFi hotspot en interfaz dedicada, no en la misma que la administración |
| Acceso admin restringido | Solo desde red interna o VPN, nunca expuesto a Internet |
| Firewall del host | Reglas iptables base que limiten acceso al panel admin |
| Actualizaciones del OS | Mantener el sistema operativo actualizado |

Referencia: `docs/network_setup.md`, `docs/wifi_security.md`

---

## 7. Seguridad de dependencias

| Dependencia | Uso | Verificación |
|------------|-----|-------------|
| `github.com/mattn/go-sqlite3` | Driver SQLite | Verificar CVEs periódicamente |
| `golang.org/x/crypto` | bcrypt para passwords | Mantener actualizado |
| `hostapd` | Access point daemon | Actualizar con el OS |
| `dnsmasq` | DNS/DHCP | Actualizar con el OS |
| `iptables` | Firewall | Verificar configuración |

---

## 8. Referencias

| Documento | Contenido |
|-----------|-----------|
| `docs/dpia.md` | Evaluación de impacto con matriz de riesgos |
| `docs/data_inventory.md` | Inventario de datos y clasificación de sensibilidad |
| `docs/data_retention.md` | Política de retención de datos |
| `docs/incident_response.md` | Plan de respuesta a incidentes |
| `docs/privacy_policy.md` | Política de privacidad para usuarios |
| `docs/lopdp.md` | Análisis de cumplimiento LOPDP |
