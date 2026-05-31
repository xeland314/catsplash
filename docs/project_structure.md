# Portal Cautivo — Estructura del Proyecto

## Árbol de carpetas

```
captive-portal/
│
├── main.go                     # Entrypoint: arranca config, DB, firewall y servidor HTTP
│
├── go.mod                      # Módulo: github.com/xeland314/captive-portal
│                               # Dependencias: mattn/go-sqlite3 (CGO, driver SQLite)
│
├── Makefile                    # build, install, clean, run (requiere root para iptables)
├── README.md
│
├── config/
│   ├── config.go               # Struct Config + loader: lee config.toml y flags de CLI
│   └── config.toml             # Archivo de configuración del operador
│                               #   iface = "wlan0"          (interfaz de red vigilada)
│                               #   portal_port = 8080        (puerto del servidor HTTP)
│                               #   session_timeout = 3600    (segundos, expiración absoluta)
│                               #   idle_timeout = 600        (segundos, expiración por inactividad)
│                               #   db_path = "/var/lib/captive-portal/captive.db"
│                               #   redirect_url = "http://192.168.1.1:8080/portal"
│
├── firewall/
│   ├── firewall.go             # Init(): aplica el conjunto completo de reglas al arrancar
│   │                           # Teardown(): limpia TODAS las reglas al salir (fail-secure)
│   │                           # Ejecuta iptables/nftables vía os/exec con rutas absolutas
│   │
│   ├── rules.go                # Reglas base (aplicadas una sola vez al arrancar):
│   │                           #   - Acepta DHCP (udp 67/68) y DNS local (udp/tcp 53)
│   │                           #   - Bloquea FORWARD por defecto para la interfaz vigilada
│   │                           #   - Inserta cadena personalizada CAPTIVE en PREROUTING
│   │
│   ├── redirect.go             # SetupRedirect(): redirige HTTP (tcp 80) de clientes
│   │                           # no autenticados a la IP:puerto del portal local
│   │                           # (DNAT en la cadena CAPTIVE)
│   │
│   ├── allow.go                # AllowClient(mac, ip string): agrega regla ACCEPT
│   │                           # para el par MAC+IP específico; permite FORWARD
│   │                           # hacia la red externa
│   │
│   ├── block.go                # BlockClient(mac, ip string): elimina la regla ACCEPT;
│   │                           # el tráfico vuelve a quedar bloqueado por defecto
│   │
│   └── cleanup.go              # FlushChain(): elimina la cadena CAPTIVE completa;
│                               # usado en Teardown() y en arranques limpios
│
├── state/
│   ├── db.go                   # Open(path): abre la BD SQLite, aplica PRAGMA WAL,
│   │                           # ejecuta schema.sql si las tablas no existen
│   │                           # Close(): cierra la conexión limpiamente
│   │
│   ├── client.go               # Struct Client { MAC, IP, State, ConnectedAt, LastSeen }
│   │                           # State ∈ { "pending", "authenticated" }
│   │                           # UpsertClient(), GetClient(), ListAuthenticated()
│   │
│   ├── session.go              # Authenticate(mac, ip): cambia State → "authenticated",
│   │                           # registra ConnectedAt = NOW()
│   │                           # Deauthenticate(mac): cambia State → "pending"
│   │                           # UpdateLastSeen(mac): actualiza timestamp de actividad
│   │
│   ├── reaper.go               # Goroutine que corre cada 30s:
│   │                           #   - ExpireAbsolute(): desautentica clientes que
│   │                           #     superaron session_timeout desde ConnectedAt
│   │                           #   - ExpireIdle(): desautentica clientes que
│   │                           #     superaron idle_timeout desde LastSeen
│   │                           #   Llama a firewall.BlockClient() por cada expirado
│   │
│   └── schema.sql              # CREATE TABLE IF NOT EXISTS clients (
│                               #   mac TEXT PRIMARY KEY,
│                               #   ip  TEXT NOT NULL,
│                               #   state TEXT NOT NULL DEFAULT 'pending',
│                               #   connected_at INTEGER,
│                               #   last_seen INTEGER
│                               # );
│
├── server/
│   ├── server.go               # New(cfg, db, fw): construye http.Server con mux propio
│   │                           # Start(): escucha en portal_port, maneja señal SIGTERM
│   │
│   ├── handler_redirect.go     # GET /*  (catch-all para clientes no autenticados):
│   │                           # detecta si el cliente está autenticado;
│   │                           # si no → 302 a /portal
│   │                           # también responde a captive portal detection (CNA):
│   │                           #   /generate_204, /hotspot-detect.html, /ncsi.txt
│   │
│   ├── handler_portal.go       # GET /portal: renderiza templates/portal.html
│   │                           # Extrae IP del RemoteAddr, busca MAC en ARP (/proc/net/arp)
│   │
│   ├── handler_auth.go         # POST /auth: procesa el formulario de aceptación
│   │                           # Valida token CSRF simple (nonce en cookie + campo hidden)
│   │                           # Llama state.Authenticate() + firewall.AllowClient()
│   │                           # Redirige a /success o devuelve error
│   │
│   ├── middleware.go           # LogMiddleware: registra IP, método, path, duración
│   │                           # AuthCheckMiddleware: comprueba estado del cliente
│   │                           # antes de servir cualquier ruta que no sea /portal o /auth
│   │
│   └── session.go              # getMACFromIP(ip): lee /proc/net/arp para obtener MAC
│                               # buildNonce(): genera token CSRF con crypto/rand
│                               # validateNonce(): compara cookie vs campo de formulario
│
├── templates/
│   │   # Embebidos en el binario con //go:embed (no requieren disco en producción)
│   │   # HTML puro, sin JS. CSS embebido en <style> dentro de cada archivo.
│   │   # Paleta: fondo oscuro (similar al dark de C-Slides), texto claro.
│   │   # Compatible con mini-navegadores CNA (sin JS, sin externos, sin webfonts).
│   │
│   ├── portal.html             # Página principal: nombre de red, términos de uso,
│   │                           # botón "Conectar" → POST /auth
│   │                           # Campo hidden con nonce CSRF
│   │
│   ├── success.html            # Confirmación de acceso concedido
│   │                           # Muestra tiempo de sesión disponible
│   │                           # Meta-refresh a la URL destino del usuario (si se guardó)
│   │
│   └── error.html              # Error genérico: mensaje + link de vuelta a /portal
│
└── docs/
    ├── project_structure.md    # Este archivo
    ├── requirements.md         # Requerimientos del sistema
    └── network_setup.md        # Guía de hostapd + dnsmasq
```

## Dependencias

```
# go.mod (módulo principal)
module github.com/xeland314/captive-portal

go 1.2x

require (
    github.com/mattn/go-sqlite3 v1.14.22   # CGO — único driver SQLite puro y maduro
)
```

> `mattn/go-sqlite3` requiere CGO y `gcc`. En Debian: `apt install gcc`.  
> No hay más dependencias externas. La stdlib cubre HTTP, templates, logging y crypto.

## Compilación

```makefile
# Makefile
BINARY  = captive-portal
PREFIX  = /usr/local/bin
CGO_ENABLED = 1

build:
	CGO_ENABLED=1 go build -o $(BINARY) ./...

install: build
	install -m 755 $(BINARY) $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY) captive.db

# Requiere root (iptables)
run: build
	sudo ./$(BINARY) -config config.toml
```

## Flujo de arranque (`main.go`)

```
1. Parsear flags + cargar config.toml
2. Abrir BD SQLite (state.Open)
3. Llamar firewall.Init()         ← inserta reglas base + cadena CAPTIVE + DNAT
4. defer firewall.Teardown()      ← garantiza limpieza al salir (fail-secure)
5. Arrancar state.Reaper()        ← goroutine de expiración
6. Arrancar server.Start()        ← bloquea hasta SIGTERM/SIGINT
```

## Decisiones de diseño

| Decisión | Razón |
| :--- | :--- |
| `net/http` stdlib, sin frameworks | Binario mínimo, sin dependencias, compila rápido |
| `mattn/go-sqlite3` con CGO | Único driver SQLite fiable; go-sqlite3 puro (modernc) tarda mucho más en compilar |
| HTML con CSS embebido, sin JS | Compatible con CNA de iOS/Android; no necesita parser JS |
| `//go:embed` para templates | Binario único, sin archivos sueltos en producción |
| `os/exec` para iptables | Go no tiene bindings de netfilter puros estables; exec es predecible y auditarle |
| Lectura de MAC desde `/proc/net/arp` | Sin dependencias; funciona en cualquier kernel Linux >= 2.6 |
| `defer firewall.Teardown()` en main | Garantiza fail-secure aunque el servidor panic |
