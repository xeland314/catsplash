# Características de Catsplash 🐱

Este documento resume las capacidades actuales del proyecto y su alcance real.

## Funcionalidades disponibles
- Redirección de tráfico HTTP al portal cautivo.
- Compatibilidad con el flujo CNA de Android e iOS para abrir el portal automáticamente.
- Gestión de clientes y sesiones en SQLite con expiración por tiempo absoluto e inactividad.
- Reglas de firewall dinamicas con iptables y limpieza segura al cerrar.
- CLI `catsctl` para consultar el estado y listar clientes.
- Panel de administración con vista de sesiones activas, control de cuota y ancho de banda.
- Consentimiento explícito del usuario antes de la conexión (LOPDP).
- Endpoints ARCO+: acceso a datos (`/data-request`) y eliminación (`/data-deletion`).
- Registro de auditoría para trazabilidad de accesos y eliminaciones de datos personales.
- Anonimización de MACs en logs del sistema (SHA-256 truncado).
- Rate limiting por IP en endpoints de autenticación y ARCO+.
- Política de privacidad accesible desde el portal (`/privacy`).
- CI/CD automatizado: build, test, vet, gosec (security scan), lint, LOPDP compliance check.

## Cumplimiento LOPDP
- Consentimiento explícito con checkbox + validación backend + timestamp.
- Derechos ARCO+: acceso, portabilidad (JSON), cancelación, oposición.
- Identidad por cookie (anti-IDOR) en endpoints de datos.
- Evaluación de impacto (DPIA) documentada.
- Inventario de datos personales con clasificación de sensibilidad.
- Plan de retención de datos vinculado al reaper de sesiones.
- Plan de respuesta a incidentes con plantilla de notificación.

## Limitaciones conocidas
- La interceptación HTTPS no está implementada como solución de producción.
- SQLite no está cifrado en reposo (pendiente SQLCipher).
- El panel admin no tiene HTTPS (pendiente certificado).
- La retención de logs del sistema no está automatizada.

## Estado operativo
Catsplash está pensado para despliegues sencillos en Linux, normalmente con una interfaz Wi-Fi de hotspot y otra interfaz de salida a Internet.
