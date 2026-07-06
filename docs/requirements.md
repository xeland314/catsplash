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
