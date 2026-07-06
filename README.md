# Catsplash 🐱

> Catsplash es un portal cautivo ligero para Linux que permite abrir una red temporal, bloquear el acceso a Internet hasta que el usuario acepta los términos y liberar el tráfico de forma segura con Go, SQLite y iptables.

![demo](assets/demo.svg)

## Features
- Portal cautivo compatible con el flujo de Captive Network Assistant de Android e iOS.
- Gestión de sesiones y clientes con SQLite, expiración por tiempo absoluto e inactividad.
- Herramienta de control `catsctl` para revisar estado, listar clientes y gestionar sesiones.
- Instalador guiado con `setup.sh` que prepara hostapd, dnsmasq, reglas de firewall y la configuración del sistema.

## Architecture

```text
Cliente Wi‑Fi ──> Punto de Acceso (hostapd) ──> Catsplash
                           │                         │
                           │                         ├─> Firewall (iptables)
                           │                         ├─> Base de datos SQLite
                           │                         └─> Portal web /auth
```

## Requerimientos
- Linux con privilegios de root.
- Debian/Ubuntu recomendado para el asistente de instalación.
- Dependencias del sistema: `hostapd`, `dnsmasq`, `iptables`, `gcc`, `make`, `sqlite3` y Go 1.26+.
- Compilador compatible con CGO, ya que el driver de SQLite requiere compilación nativa.

## Instalación
```bash
git clone https://github.com/xeland314/catsplash.git
cd catsplash
make build
sudo ./setup.sh
```

El build deja los binarios en la carpeta `bin/` y el instalador copia los artefactos a `/opt/catsplash`. Sigue los pasos que te indica setup.sh hasta el final. Recuerda que es muy probable que necesites lo básico en configuración de redes.

## Uso
```bash
sudo ./bin/catsplash
sudo ./bin/catsctl status
sudo ./bin/catsctl list
```

![catsctl](assets/catsctl_example.jpg)
![setup.sh](assets/setup.sh.jpg)


## Configuración
El asistente crea un archivo de configuración en `/opt/catsplash/config.toml` y una base de datos en `/opt/catsplash/catsplash.db`. Si prefieres ajustar manualmente el entorno, puedes editar la configuración generada y reiniciar el servicio.

## Performance

Benchmarks preliminares del flujo de autenticación (`POST /auth`), simulando
clientes Wi-Fi con Linux network namespaces (sin hardware real de radio, por
lo que estos números reflejan el costo de la capa de software — servidor Go +
SQLite + iptables — no las condiciones de una red Wi-Fi real con pérdida de
paquetes o interferencia).

**Hardware de prueba:**

| Componente | Detalle |
|---|---|
| CPU | Intel Core 2 Duo E7400 @ 2.79 GHz (2 núcleos) |
| RAM | 3.58 GiB (ya al 83% de uso antes de la prueba) |
| Disco | ext4, 100 GiB |
| OS | Debian GNU/Linux 13 (trixie), kernel 6.12 |

Es intencional correr los benchmarks en hardware modesto: si Catsplash
responde bien aquí, en un router/mini-PC moderno destinado a producción va a
sobrar margen de sobra.

**Resultados** (`POST /auth`, clientes simultáneos):

| Clientes simultáneos | Promedio | p50    | p95    | p99     | Errores |
|-----------------------|----------|--------|--------|---------|---------|
| 50                    | 24.8 ms  | 20.7 ms| 71.0 ms| 104.0 ms| 0 / 50  |
| 100                   | 32.7 ms  | 28.1 ms| 88.7 ms| 121.5 ms| 0 / 100 |

*Cada fila es una sola corrida — para cifras con intervalos de confianza,
correr varias repeticiones por N.*

### Reproducir estos benchmarks

```bash
sudo ./tests/test_catsplash_multiclient.sh up -n 100
# en otra terminal:
sudo ip netns exec ns_router ./bin/catsplash
# de vuelta en la primera terminal:
sudo ./tests/test_catsplash_multiclient.sh run
sudo ./tests/test_catsplash_multiclient.sh down
```

Cada corrida de `run` guarda los datos crudos en `results/<timestamp>/results.csv`
(`client_id,http_code,time_total_s`), útil para análisis propio o para
comparar contra corridas futuras tras cambios de código.

## Contribuir
Consulta [CONTRIBUTING.md](CONTRIBUTING.md).

## Licencia
MIT
