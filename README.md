# Catsplash 🐱

Catsplash es un portal cautivo ligero y eficiente escrito en Go, diseñado para gestionar el acceso a Internet en redes locales (WiFi/Ethernet) mediante la intercepción de tráfico HTTP y la gestión dinámica de reglas de firewall con `iptables`.

## Características Principales

- **Intercepción Inteligente:** Redirige automáticamente el tráfico HTTP (puerto 80) al portal de bienvenida.
- **Detección de CNA:** Compatible con los asistentes de conectividad (Captive Network Assistant) de Android e iOS.
- **Gestión de Estado:** Utiliza SQLite para mantener un registro en tiempo real de los clientes y sus sesiones.
- **Seguridad Robusta:** Incluye protección CSRF con nonces y bypass de NAT automático para usuarios autenticados (evita el signo `!`).
- **Fail-Secure:** Limpieza automática de reglas de firewall al cerrar el programa.
- **Session Reaper:** Expira sesiones automáticamente por tiempo total (absoluto) o por inactividad.
- **Mínimas Dependencias:** Binario único con plantillas HTML embebidas.

## Estructura del Proyecto

- `config/`: Carga de configuración (TOML y CLI).
- `state/`: Persistencia en SQLite y lógica de sesiones.
- `firewall/`: Manipulación de reglas `iptables`.
- `server/`: Servidor HTTP, controladores y plantillas.
- `docs/`: Documentación detallada de red, seguridad y arquitectura.

## Instalación y Uso

### Requisitos
- Linux con `iptables`.
- `gcc` instalado (para el driver de SQLite).
- Go 1.26

### Compilación
```bash
make build
```

### Ejecución
```bash
sudo ./catsplash
```

*Nota: Asegúrate de configurar correctamente `hostapd` y `dnsmasq` siguiendo las guías en la carpeta `docs/`.*

## Pruebas
El proyecto mantiene una política de tests co-localizados dentro de cada paquete.
```bash
make test
```

## Licencia
Este proyecto es software libre bajo la licencia MIT.
