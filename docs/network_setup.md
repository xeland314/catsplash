# Configuración de Infraestructura de Red (Hostapd + Dnsmasq)

Para que el portal cautivo `catsplash` funcione correctamente, el sistema operativo debe actuar como un punto de acceso (Access Point) y proporcionar servicios de red básicos (DHCP y DNS). A continuación se detalla cómo configurar estos componentes en Debian.

## 1. Instalación de Requisitos

Ejecuta el siguiente comando para instalar las herramientas necesarias:

```bash
sudo apt update
sudo apt install hostapd dnsmasq iptables-go-flags
```

*Nota: `catsplash` se encarga de las reglas de iptables, pero estas herramientas son la base del tráfico.*

---

## 2. Configuración de la Interfaz de Red

Debes asignar una IP estática a la interfaz WiFi que emitirá la señal. En tu caso es `wlx1cbfce41183a`.

Edita `/etc/network/interfaces` o usa `ip addr` (temporal):

```bash
# Temporal (se pierde al reiniciar)
sudo ip addr add 192.168.10.1/24 dev wlx1cbfce41183a
sudo ip link set wlx1cbfce41183a up
```

## 3. Configuración de hostapd (Punto de Acceso)

Crea o edita el archivo `/etc/hostapd/hostapd.conf`.

*Para una configuración básica abierta, usa el ejemplo de abajo. Si deseas añadir una contraseña a la red (Recomendado), consulta la guía de [Seguridad WiFi](wifi_security.md).*

```conf
# Interfaz de red WiFi
interface=wlx1cbfce41183a
...
```

driver=nl80211
ssid=Mi_Red_WiFi
hw_mode=g
channel=6
auth_algs=1

# Estabilidad y compatibilidad
wmm_enabled=0
macaddr_acl=0
ignore_broadcast_ssid=0
```

Para activarlo, edita `/etc/default/hostapd` y apunta al archivo:
`DAEMON_CONF="/etc/hostapd/hostapd.conf"`

---

## 4. Configuración de dnsmasq (DHCP y DNS)

Edita `/etc/dnsmasq.conf`. Esta configuración asegura que los clientes reciban una IP y sepan que tú eres la puerta de enlace.

```conf
# Escuchar solo en la interfaz WiFi
interface=wlx1cbfce41183a
bind-interfaces

# Rango de IPs para los clientes (DHCP)
dhcp-range=192.168.10.10,192.168.10.100,255.255.255.0,12h

# Opción 3: Puerta de enlace (Gateway) - Tu IP local
dhcp-option=3,192.168.10.1

# Opción 6: Servidores DNS
# Entregamos tu IP primero para que el portal pueda interceptar nombres si fuera necesario,
# y un servidor externo de respaldo.
dhcp-option=6,192.168.10.1,8.8.8.8

# Servidor DNS que usará el propio dnsmasq para resolver peticiones externas
server=8.8.8.8
```

---

## 5. Activación de Servicios

Asegúrate de que los servicios arranquen correctamente:

```bash
sudo systemctl unmask hostapd
sudo systemctl enable hostapd
sudo systemctl restart hostapd

sudo systemctl enable dnsmasq
sudo systemctl restart dnsmasq
```

---

## 6. Flujo de catsplash con esta red

Una vez que los servicios anteriores estén corriendo, los clientes podrán conectarse al WiFi pero **no tendrán internet** (porque `catsplash` aún no ha liberado sus MACs y porque el Firewall está en modo restrictivo).

1.  **Arranque:** Ejecutas `sudo ./catsplash`.
2.  **Intercepción:** El cliente intenta entrar a `http://ejemplo.com`.
3.  **Redirección:** `catsplash` intercepta la petición y muestra el portal.
4.  **Autenticación:** El usuario acepta, `catsplash` habilita el NAT y el reenvío hacia `enp1s0`.

---

## Solución de Problemas Comunes

1.  **Interfaz "Busy":** Si `hostapd` falla, asegúrate de que NetworkManager no esté intentando controlar la tarjeta WiFi. Puedes añadirla a la sección `[keyfile] unmanaged-devices` en `/etc/NetworkManager/NetworkManager.conf`.
2.  **DNS no responde:** Verifica que el puerto 53 no esté bloqueado por otro servicio (como `systemd-resolved`). Si es así, detén `systemd-resolved` o configúralo para no escuchar en el puerto 53.
