# Características de Catsplash 🐱

Este documento resume las capacidades actuales del proyecto y su alcance real.

## Funcionalidades disponibles
- Redirección de tráfico HTTP al portal cautivo.
- Compatibilidad con el flujo CNA de Android e iOS para abrir el portal automáticamente.
- Gestión de clientes y sesiones en SQLite con expiración por tiempo absoluto e inactividad.
- Reglas de firewall dinamicas con iptables y limpieza segura al cerrar.
- CLI `catsctl` para consultar el estado y listar clientes.

## Limitaciones conocidas
- La interceptación HTTPS no está implementada como solución de producción.
- No hay todavía un panel web de administración ni autenticación por usuarios.
- La gestión avanzada de ancho de banda o vouchers queda pendiente para futuras iteraciones.

## Estado operativo
Catsplash está pensado para despliegues sencillos en Linux, normalmente con una interfaz Wi-Fi de hotspot y otra interfaz de salida a Internet.
