#!/usr/bin/env bash
# ==============================================================================
# Catsplash 🐱 - Script de Configuración Automática y Control
# Portal Cautivo para Linux (Debian/Ubuntu)
# ==============================================================================

# Colores para la interfaz
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

INSTALL_DIR="/opt/catsplash"
CONFIG_FILE="$INSTALL_DIR/config.toml"
DB_FILE="$INSTALL_DIR/catsplash.db"

# Dibujar logotipo
print_logo() {
    clear
    echo -e "${CYAN}${BOLD}"
    echo "   /\_/\  "
    echo "  ( o.o )  Catsplash 🐱"
    echo "   > ^ <   Portal Cautivo Ligero y Eficiente"
    echo -e "================================================${NC}\n"
}

# Verificar que se ejecuta como root
check_root() {
    if [[ "$EUID" -ne 0 ]]; then
        echo -e "${RED}${BOLD}[ERROR] Este script debe ejecutarse como root (sudo).${NC}"
        exit 1
    fi
}

# Listar interfaces de red
list_interfaces() {
    ip -o link show | awk -F': ' '{print $2}' | grep -v "lo"
}

# Detectar interfaz WAN (Internet) por defecto
detect_default_wan() {
    ip route show default | awk '/default/ {print $5}' | head -n1
}

# Menú principal
show_menu() {
    print_logo
    echo -e "${BOLD}Selecciona una opción:${NC}"
    echo -e "1) ${GREEN}Configurar e Instalar todo el sistema (Wizard)${NC}"
    echo -e "2) ${BLUE}Iniciar Servicios${NC}"
    echo -e "3) ${YELLOW}Detener Servicios${NC}"
    echo -e "4) ${CYAN}Reiniciar Servicios${NC}"
    echo -e "5) ${BOLD}Ver Estado del Sistema${NC}"
    echo -e "6) ${PURPLE}Ver Logs de Catsplash${NC}"
    echo -e "7) ${RED}Desinstalar / Limpiar Configuración${NC}"
    echo -e "8) Salir"
    echo
    read -rp "Opción [1-8]: " opt
    case $opt in
        1) wizard ;;
        2) start_services ;;
        3) stop_services ;;
        4) restart_services ;;
        5) show_status ;;
        6) view_logs ;;
        7) uninstall ;;
        8) exit 0 ;;
        *) echo -e "${RED}Opción no válida.${NC}"; sleep 1; show_menu ;;
    esac
}

# Asistente de configuración (Wizard)
wizard() {
    print_logo
    echo -e "${YELLOW}${BOLD}--- Asistente de Configuración ---${NC}\n"

    # 1. Instalación de Dependencias
    echo -e "${BLUE}[1/6] Instalando dependencias del sistema...${NC}"
    if [ -f /etc/debian_version ]; then
        apt-get update -qq
        apt-get install -y hostapd dnsmasq iptables gcc make golang sqlite3 >/dev/null 2>&1
        echo -e "${GREEN}[OK] Dependencias instaladas.${NC}\n"
    else
        echo -e "${YELLOW}[ADVERTENCIA] Sistema operativo no basado en Debian. Instale hostapd, dnsmasq, iptables, gcc, make, golang y sqlite3 manualmente.${NC}\n"
    fi

    # 2. Selección de Interfaces de Red
    echo -e "${BLUE}[2/6] Configuración de Interfaces de Red:${NC}"
    echo -e "Interfaces disponibles:"
    list_interfaces
    echo

    # Interfaz AP
    while true; do
        read -rp "Introduce la interfaz WiFi para el Hotspot (ej. wlan0): " AP_IFACE
        if [ -n "$AP_IFACE" ] && ip link show "$AP_IFACE" >/dev/null 2>&1; then
            break
        else
            echo -e "${RED}Interfaz no válida o no encontrada.${NC}"
        fi
    done

    # Interfaz WAN
    DEFAULT_WAN=$(detect_default_wan)
    read -rp "Introduce la interfaz de Internet/WAN [Por defecto: $DEFAULT_WAN]: " WAN_IFACE
    WAN_IFACE=${WAN_IFACE:-$DEFAULT_WAN}
    while true; do
        if [ -n "$WAN_IFACE" ] && ip link show "$WAN_IFACE" >/dev/null 2>&1; then
            break
        else
            echo -e "${RED}Interfaz WAN no válida o no encontrada.${NC}"
            read -rp "Introduce la interfaz de Internet/WAN: " WAN_IFACE
        fi
    done
    echo -e "${GREEN}[OK] Hotspot: $AP_IFACE | Internet: $WAN_IFACE${NC}\n"

    # 3. Configuración WiFi (Hostapd)
    echo -e "${BLUE}[3/6] Configuración del Punto de Acceso WiFi:${NC}"
    read -rp "Nombre de la red WiFi (SSID) [Por defecto: Catsplash_WiFi]: " SSID
    SSID=${SSID:-Catsplash_WiFi}

    echo -e "Seguridad de la red:"
    echo "1) Red Abierta (Sin contraseña)"
    echo "2) WPA2-PSK (Con contraseña)"
    read -rp "Opción [1-2]: " SEC_OPT
    
    WPA_PASS=""
    if [[ "$SEC_OPT" == "2" ]]; then
        while true; do
            read -rsp "Introduce la contraseña WiFi (mínimo 8 caracteres): " WPA_PASS
            echo
            if [[ ${#WPA_PASS} -ge 8 ]]; then
                break
            else
                echo -e "${RED}La contraseña debe tener al menos 8 caracteres.${NC}"
            fi
        done
    fi

    read -rp "Canal WiFi (1-11) [Por defecto: 6]: " CHANNEL
    CHANNEL=${CHANNEL:-6}
    echo -e "${GREEN}[OK] Configuración WiFi completada.${NC}\n"

    # 4. Configuración de IP y DHCP
    echo -e "${BLUE}[4/6] Configuración de direccionamiento IP (DHCP):${NC}"
    read -rp "IP del Portal/Puerta de Enlace [Por defecto: 192.168.10.1]: " GATEWAY_IP
    GATEWAY_IP=${GATEWAY_IP:-192.168.10.1}

    # Extraer los primeros 3 octetos de la IP
    IP_PREFIX=$(echo "$GATEWAY_IP" | cut -d'.' -f1-3)

    read -rp "Rango DHCP inicio [Por defecto: $IP_PREFIX.10]: " DHCP_START
    DHCP_START=${DHCP_START:-"$IP_PREFIX.10"}
    
    read -rp "Rango DHCP fin [Por defecto: $IP_PREFIX.100]: " DHCP_END
    DHCP_END=${DHCP_END:-"$IP_PREFIX.100"}

    read -rp "Puerto del servidor del portal [Por defecto: 8080]: " PORTAL_PORT
    PORTAL_PORT=${PORTAL_PORT:-8080}

    read -rp "Tiempo de sesión absoluto (segundos) [Por defecto: 3600]: " SESSION_TIMEOUT
    SESSION_TIMEOUT=${SESSION_TIMEOUT:-3600}

    read -rp "Tiempo de inactividad máximo (segundos) [Por defecto: 900]: " IDLE_TIMEOUT
    IDLE_TIMEOUT=${IDLE_TIMEOUT:-900}
    echo -e "${GREEN}[OK] Direccionamiento configurado.${NC}\n"

    # 5. Compilación del binario Catsplash y Catsctl
    echo -e "${BLUE}[5/6] Compilando Catsplash y Catsctl...${NC}"
    if [ -f Makefile ]; then
        make build
        if [ $? -eq 0 ] && [ -f catsplash ] && [ -f catsctl ]; then
            echo -e "${GREEN}[OK] Binarios compilados con éxito.${NC}\n"
        else
            echo -e "${RED}[ERROR] Falló la compilación. Verifique la instalación de Go y gcc.${NC}"
            exit 1
        fi
    else
        echo -e "${RED}[ERROR] No se encontró el archivo Makefile. Asegúrese de ejecutar el script en el directorio del proyecto.${NC}"
        exit 1
    fi

    # 6. Escribir configuraciones y crear servicios
    echo -e "${BLUE}[6/6] Aplicando configuraciones en el sistema...${NC}"

    # Crear directorio de instalación
    mkdir -p "$INSTALL_DIR"
    cp catsplash "$INSTALL_DIR/"
    cp catsctl "$INSTALL_DIR/"
    ln -sf "$INSTALL_DIR/catsctl" /usr/local/bin/catsctl

    # Crear archivo de configuración config.toml
    cat > "$CONFIG_FILE" <<EOF
iface = "$AP_IFACE"
wan_iface = "$WAN_IFACE"
portal_port = $PORTAL_PORT
session_timeout = $SESSION_TIMEOUT
idle_timeout = $IDLE_TIMEOUT
db_path = "$DB_FILE"
redirect_url = "http://$GATEWAY_IP:$PORTAL_PORT/portal"
EOF

    # Configurar Hostapd
    # Respaldar original
    [ -f /etc/hostapd/hostapd.conf ] && cp /etc/hostapd/hostapd.conf /etc/hostapd/hostapd.conf.bak
    
    cat > /etc/hostapd/hostapd.conf <<EOF
interface=$AP_IFACE
driver=nl80211
ssid=$SSID
hw_mode=g
channel=$CHANNEL
auth_algs=1
wmm_enabled=0
macaddr_acl=0
ignore_broadcast_ssid=0
EOF

    if [[ -n "$WPA_PASS" ]]; then
        cat >> /etc/hostapd/hostapd.conf <<EOF
wpa=2
wpa_passphrase=$WPA_PASS
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
EOF
    fi

    # Configurar archivo default de hostapd
    [ -f /etc/default/hostapd ] && cp /etc/default/hostapd /etc/default/hostapd.bak
    sed -i 's|#DAEMON_CONF=""|DAEMON_CONF="/etc/hostapd/hostapd.conf"|g' /etc/default/hostapd
    sed -i 's|DAEMON_CONF=""|DAEMON_CONF="/etc/hostapd/hostapd.conf"|g' /etc/default/hostapd

    # Configurar Dnsmasq
    # Respaldar original
    [ -f /etc/dnsmasq.conf ] && cp /etc/dnsmasq.conf /etc/dnsmasq.conf.bak
    
    cat > /etc/dnsmasq.conf <<EOF
# Escuchar solo en la interfaz WiFi del Hotspot
interface=$AP_IFACE
bind-interfaces

# Rango de direcciones IP para clientes
dhcp-range=$DHCP_START,$DHCP_END,255.255.255.0,12h

# Opción 3: Puerta de enlace (Gateway)
dhcp-option=3,$GATEWAY_IP

# Opción 6: Servidor DNS (Nosotros mismos y Google de respaldo)
dhcp-option=6,$GATEWAY_IP,8.8.8.8

# Servidor DNS upstream
server=8.8.8.8
EOF

    # Resolver conflicto con NetworkManager si está activo
    if systemctl is-active --quiet NetworkManager; then
        echo -e "${YELLOW}[INFO] NetworkManager detectado. Configurando $AP_IFACE como no gestionado...${NC}"
        mkdir -p /etc/NetworkManager/conf.d
        cat > /etc/NetworkManager/conf.d/99-catsplash.conf <<EOF
[keyfile]
unmanaged-devices=interface-name:$AP_IFACE
EOF
        systemctl restart NetworkManager
    fi

    # Encontrar ruta absoluta de 'ip'
    IP_BIN=$(which ip)
    IP_BIN=${IP_BIN:-/sbin/ip}

    # Crear servicio catsplash-ip
    cat > /etc/systemd/system/catsplash-ip.service <<EOF
[Unit]
Description=Configurar IP Estatica para Interfaz de Portal Cautivo
Before=hostapd.service dnsmasq.service
After=network.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=-$IP_BIN addr add $GATEWAY_IP/24 dev $AP_IFACE
ExecStart=$IP_BIN link set $AP_IFACE up
ExecStop=-$IP_BIN addr flush dev $AP_IFACE
ExecStop=-$IP_BIN link set $AP_IFACE down

[Install]
WantedBy=multi-user.target
EOF

    # Crear servicio catsplash
    cat > /etc/systemd/system/catsplash.service <<EOF
[Unit]
Description=Servicio Catsplash Captive Portal
After=network.target hostapd.service dnsmasq.service catsplash-ip.service
Wants=hostapd.service dnsmasq.service catsplash-ip.service

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/catsplash
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # Habilitar servicios en systemd
    systemctl daemon-reload
    systemctl unmask hostapd >/dev/null 2>&1
    systemctl enable catsplash-ip.service >/dev/null 2>&1
    systemctl enable hostapd >/dev/null 2>&1
    systemctl enable dnsmasq >/dev/null 2>&1
    systemctl enable catsplash.service >/dev/null 2>&1

    echo -e "${GREEN}${BOLD}[OK] Configuración completada con éxito!${NC}"
    read -rp "¿Deseas iniciar los servicios ahora? (s/n): " start_now
    if [[ "$start_now" =~ ^[sS]$ ]]; then
        start_services
    else
        echo -e "${YELLOW}Puedes iniciar los servicios más tarde desde este menú o con: systemctl start catsplash${NC}"
        sleep 2
        show_menu
    fi
}

# Iniciar servicios
start_services() {
    print_logo
    echo -e "${BLUE}Iniciando servicios del portal cautivo...${NC}"
    
    echo -n "1. Configurando IP en interfaz... "
    systemctl start catsplash-ip.service
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "2. Iniciando hostapd (Hotspot WiFi)... "
    systemctl restart hostapd
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "3. Iniciando dnsmasq (DHCP/DNS)... "
    systemctl restart dnsmasq
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "4. Iniciando Catsplash (Portal Cautivo)... "
    systemctl restart catsplash.service
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi

    echo -e "\n${GREEN}${BOLD}¡Portal cautivo en marcha!${NC}"
    sleep 2
    show_menu
}

# Detener servicios
stop_services() {
    print_logo
    echo -e "${YELLOW}Deteniendo servicios del portal cautivo...${NC}"
    
    echo -n "1. Deteniendo Catsplash (Portal Cautivo)... "
    systemctl stop catsplash.service
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "2. Deteniendo dnsmasq (DHCP/DNS)... "
    systemctl stop dnsmasq
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "3. Deteniendo hostapd (Hotspot WiFi)... "
    systemctl stop hostapd
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi
    
    echo -n "4. Limpiando IP en interfaz... "
    systemctl stop catsplash-ip.service
    if [ $? -eq 0 ]; then echo -e "${GREEN}[OK]${NC}"; else echo -e "${RED}[FALLÓ]${NC}"; fi

    echo -e "\n${YELLOW}${BOLD}Servicios detenidos.${NC}"
    sleep 2
    show_menu
}

# Reiniciar servicios
restart_services() {
    print_logo
    echo -e "${BLUE}Reiniciando servicios del portal cautivo...${NC}"
    systemctl restart catsplash-ip.service
    systemctl restart hostapd
    systemctl restart dnsmasq
    systemctl restart catsplash.service
    echo -e "${GREEN}${BOLD}Servicios reiniciados.${NC}"
    sleep 2
    show_menu
}

# Mostrar estado
show_status() {
    print_logo
    echo -e "${BOLD}--- Estado del Sistema ---${NC}\n"

    # Verificar estado de servicios
    for service in catsplash-ip hostapd dnsmasq catsplash; do
        if systemctl is-active --quiet "$service"; then
            echo -e "Servicio ${BOLD}$service${NC}: ${GREEN}ACTIVO (Corriendo)${NC}"
        else
            echo -e "Servicio ${BOLD}$service${NC}: ${RED}INACTIVO (Detenido)${NC}"
        fi
    done
    echo

    # Leer clientes activos desde base de datos
    if [ -f "$DB_FILE" ] && command -v sqlite3 >/dev/null 2>&1; then
        echo -e "${BOLD}Clientes registrados en base de datos:${NC}"
        # Hacer query de sesiones activas
        active_sessions=$(sqlite3 "$DB_FILE" "SELECT mac, ip, expires_at, idle_at FROM sessions;" 2>/dev/null)
        if [ -n "$active_sessions" ]; then
            echo -e "${CYAN}MAC                | IP             | Expira en (seg) / Hora Expiración${NC}"
            echo "----------------------------------------------------------------------------"
            while IFS='|' read -r mac ip expires_at idle_at; do
                now=$(date +%s)
                remaining=$((expires_at - now))
                if [ $remaining -gt 0 ]; then
                    echo -e "$mac  | $ip  | Expirará en ${remaining}s"
                else
                    echo -e "$mac  | $ip  | Expirado"
                fi
            done <<< "$active_sessions"
        else
            echo -e "${YELLOW}No hay sesiones activas en la base de datos.${NC}"
        fi
    else
        echo -e "${YELLOW}Base de datos no inicializada o sqlite3 no disponible.${NC}"
    fi

    echo
    read -rp "Presiona [Enter] para volver al menú."
    show_menu
}

# Ver logs
view_logs() {
    print_logo
    echo -e "${PURPLE}${BOLD}Mostrando últimos logs de Catsplash (Ctrl+C para salir):${NC}\n"
    journalctl -u catsplash.service -n 50 -f
    show_menu
}

# Desinstalar y limpiar
uninstall() {
    print_logo
    echo -e "${RED}${BOLD}¿Estás seguro de que deseas desinstalar Catsplash y restaurar las configuraciones originales? (s/n)${NC}"
    read -rp "Respuesta: " confirm
    if [[ ! "$confirm" =~ ^[sS]$ ]]; then
        echo -e "${GREEN}Desinstalación cancelada.${NC}"
        sleep 1
        show_menu
    fi

    echo -e "\n${YELLOW}Deteniendo y deshabilitando servicios...${NC}"
    systemctl stop catsplash.service >/dev/null 2>&1
    systemctl disable catsplash.service >/dev/null 2>&1
    systemctl stop dnsmasq >/dev/null 2>&1
    systemctl disable dnsmasq >/dev/null 2>&1
    systemctl stop hostapd >/dev/null 2>&1
    systemctl disable hostapd >/dev/null 2>&1
    systemctl stop catsplash-ip.service >/dev/null 2>&1
    systemctl disable catsplash-ip.service >/dev/null 2>&1

    # Eliminar archivos del servicio
    echo -e "${YELLOW}Eliminando archivos de servicios systemd...${NC}"
    rm -f /etc/systemd/system/catsplash.service
    rm -f /etc/systemd/system/catsplash-ip.service
    systemctl daemon-reload

    # Restaurar configuraciones de hostapd y dnsmasq
    echo -e "${YELLOW}Restaurando archivos de configuración de respaldo...${NC}"
    if [ -f /etc/hostapd/hostapd.conf.bak ]; then
        mv /etc/hostapd/hostapd.conf.bak /etc/hostapd/hostapd.conf
    else
        rm -f /etc/hostapd/hostapd.conf
    fi

    if [ -f /etc/default/hostapd.bak ]; then
        mv /etc/default/hostapd.bak /etc/default/hostapd
    fi

    if [ -f /etc/dnsmasq.conf.bak ]; then
        mv /etc/dnsmasq.conf.bak /etc/dnsmasq.conf
    else
        rm -f /etc/dnsmasq.conf
    fi

    # Eliminar regla de NetworkManager
    rm -f /etc/NetworkManager/conf.d/99-catsplash.conf
    if systemctl is-active --quiet NetworkManager; then
        systemctl restart NetworkManager
    fi

    # Eliminar directorio de instalación y enlaces
    echo -e "${YELLOW}Eliminando directorio de instalación $INSTALL_DIR y enlaces...${NC}"
    rm -f /usr/local/bin/catsctl
    rm -rf "$INSTALL_DIR"

    echo -e "${GREEN}${BOLD}Desinstalación completada con éxito.${NC}"
    sleep 2
    exit 0
}

# Ejecutar script
check_root

# Si se pasa un argumento por comando (start/stop/restart/status/logs/uninstall)
if [[ -n "$1" ]]; then
    case "$1" in
        start) start_services ;;
        stop) stop_services ;;
        restart) restart_services ;;
        status) show_status ;;
        logs) view_logs ;;
        uninstall) uninstall ;;
        *) echo "Uso: $0 [start|stop|restart|status|logs|uninstall]" ;;
    esac
else
    show_menu
fi
