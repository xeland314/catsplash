# Requisitos de Catsplash

Catsplash está pensado para funcionar en un equipo Linux que actúe como puerta de enlace o punto de acceso temporal.

## Requisitos del sistema
- Linux con kernel moderno y acceso root.
- Debian/Ubuntu recomendados para el asistente `setup.sh`.
- `hostapd`, `dnsmasq`, `iptables`, `gcc`, `make`, `sqlite3` y Go 1.26+.
- Al menos dos interfaces de red: una para el hotspot y otra para Internet.

## Requisitos funcionales
- El sistema debe aislar el tráfico de los clientes no autenticados.
- Debe redirigir peticiones HTTP al portal cautivo.
- Debe registrar clientes y sesiones en SQLite.
- Debe expirar sesiones por tiempo absoluto o inactividad.
- Debe limpiar las reglas de firewall al cerrar o fallar de forma segura.

## Requisitos no funcionales
- Bajo consumo de recursos.
- Compatibilidad con navegadores móviles y CNA.
- Configuración simple para despliegues locales o de laboratorio.

## Requisitos de cumplimiento LOPDP

### Consentimiento (Art. 14)
- El portal debe solicitar consentimiento explícito antes de recopilar datos personales.
- El consentimiento debe registrado con timestamp en la base de datos.
- El usuario debe poder acceder a la política de privacidad antes de consentir.

### Derechos ARCO+ (Arts. 26-30)
- **Acceso**: El titular puede solicitar sus datos personales almacenados.
- **Portabilidad**: Los datos se exportan en formato JSON.
- **Cancelación**: El titular puede solicitar la eliminación de sus datos.
- **Oposición**: El titular puede desconectarse en cualquier momento.
- La identidad del solicitante se resuelve desde cookie de sesión, nunca desde parámetros URL.

### Retención de datos (Art. 13)
- Los datos de sesión se eliminan automáticamente al expirar la sesión.
- No hay retención indefinida de datos personales.
- Los logs no contienen datos personales identificables (MACs anonimizadas).

### Seguridad (Art. 50)
- Contraseñas de administrador hasheadas con bcrypt.
- Protección CSRF en todos los endpoints que modifican datos.
- Rate limiting en endpoints de autenticación.
- Sanitización de inputs (validación MAC con regex).
- Comandos del sistema con argumentos separados (sin shell).

### Auditoría
- Registro de accesos a datos personales (`audit_log`).
- Registro de eliminaciones de datos.
- Registro de intentos de autenticación (éxito y fallo).
- MACs enmascaradas en registros de auditoría.

### Documentación
- Política de privacidad accesible para usuarios.
- Inventario de datos personales.
- Evaluación de impacto (DPIA).
- Plan de retención de datos.
- Plan de respuesta a incidentes.
- Directrices de seguridad para contributors.
