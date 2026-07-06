# Estructura del proyecto Catsplash

## Árbol de carpetas

```text
catsplash/
├── main.go                 # Punto de entrada que arranca la app
├── Makefile                # Compila en bin/ y limpia artefactos
├── setup.sh                # Instalador asistido para preparar el sistema
├── go.mod                  # Módulo Go y dependencias
├── config/                 # Carga de configuración desde TOML y flags
├── firewall/               # Reglas de iptables y control del tráfico
├── server/                 # Portal web, redirección y handlers HTTP
├── state/                  # Base de datos SQLite, clientes y sesiones
├── catsctl/                # Cliente CLI para inspeccionar y gestionar Catsplash
├── docs/                   # Documentación de requisitos, instalación y red
└── bin/                    # Binarios generados por make build
```

## Qué hace cada parte
- `config/`: define la configuración del hotspot, puertos y rutas de la base de datos.
- `firewall/`: aplica reglas de bloqueo y liberación de tráfico para clientes no autenticados.
- `server/`: sirve el portal de bienvenida y procesa la autenticación.
- `state/`: guarda clientes, estados y tiempos de sesión en SQLite.
- `catsctl/`: ofrece una herramienta CLI para status y listados.
- `setup.sh`: automatiza la instalación de dependencias y la configuración del sistema.

## Flujo de arranque
1. El proceso carga la configuración.
2. Abre la base de datos SQLite.
3. Inicializa las reglas de firewall.
4. Arranca el servidor web del portal.
5. Mantiene las sesiones con el reaper y la CLI de control.

## Dependencias principales
- Go 1.26+
- `github.com/BurntSushi/toml`
- `github.com/mattn/go-sqlite3`
- `gcc` y CGO para la integración con SQLite
