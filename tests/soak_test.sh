#!/bin/bash
#
# soak_test.sh
#
# Corre Catsplash bajo carga ligera y constante durante varias horas
# (por defecto 6) para detectar fugas de memoria, reglas de iptables
# huerfanas tras expiracion de sesiones, o crecimiento anormal del .db.
#
# NO modifica test_catsplash_multiclient.sh: solo invoca sus
# subcomandos 'up' y 'down'.
#
# Diseñado para consumir el minimo de disco posible: un unico archivo
# de log con una linea compacta por muestra (por defecto cada 5 min,
# ~72 lineas en 6 horas -> unos pocos KB en total).
#
# Flujo:
#   sudo ./tests/soak_test.sh -n 10
#   # en otra terminal, cuando el script lo indique:
#   sudo ip netns exec ns_router ./bin/catsplash
#   # el soak test detecta que ya responde y arranca solo
#
set -euo pipefail

MULTICLIENT_SCRIPT="./tests/test_catsplash_multiclient.sh"
N_CLIENTS=10
HOURS=1
INTERVAL_MIN=1
DB_PATH="./catsplash.db"
LOG_FILE="./soak_results.log"
ROUTER_AP_IP="192.168.6.1"
PORTAL_PORT="8080"

usage() {
  cat <<EOF
Uso: sudo $0 [-n N_CLIENTES] [--hours H] [--interval MIN] [--db PATH]

  -n N          Clientes simulados para la topologia (default: 10)
  --hours H     Duracion total del soak test en horas (default: 6)
  --interval M  Minutos entre muestras (default: 5)
  --db PATH     Ruta al catsplash.db a monitorear (default: ./catsplash.db)
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    -n) N_CLIENTS="$2"; shift 2 ;;
    --hours) HOURS="$2"; shift 2 ;;
    --interval) INTERVAL_MIN="$2"; shift 2 ;;
    --db) DB_PATH="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Opcion desconocida: $1"; usage; exit 1 ;;
  esac
done

if [ "$EUID" -ne 0 ]; then
  echo "Por favor, ejecuta este script con sudo."
  exit 1
fi

if [ ! -x "$MULTICLIENT_SCRIPT" ]; then
  echo "No encuentro $MULTICLIENT_SCRIPT (ejecutable). Ajusta la variable MULTICLIENT_SCRIPT."
  exit 1
fi

DURATION_SEC=$((HOURS * 3600))
INTERVAL_SEC=$((INTERVAL_MIN * 60))

echo "Levantando topologia con $N_CLIENTS cliente(s)..."
"$MULTICLIENT_SCRIPT" up -n "$N_CLIENTS"

echo ""
echo "Ahora, en OTRA terminal:"
echo "   sudo ip netns exec ns_router ./bin/catsplash"
echo ""
echo "Esperando a que Catsplash responda..."

while true; do
  if ip netns exec ns_client1 curl -s -o /dev/null --connect-timeout 1 \
      "http://$ROUTER_AP_IP:$PORTAL_PORT/" 2>/dev/null; then
    break
  fi
  sleep 2
done
echo "Catsplash detectado. Iniciando soak test de $HOURS hora(s), muestreo cada $INTERVAL_MIN min."
echo ""

# Cabecera del log (una sola vez, archivo unico que se va agregando)
echo "timestamp,elapsed_min,mem_rss_kb,db_size_bytes,nat_rules,filter_rules,active,pending,total" > "$LOG_FILE"

# Descubre el PID de catsplash dentro del namespace (puede no existir
# aun si el usuario tarda en arrancarlo; se reintenta cada muestra).
get_catsplash_pid() {
  ip netns pids ns_router 2>/dev/null | while read -r pid; do
    if [ -r "/proc/$pid/comm" ] && grep -q catsplash "/proc/$pid/comm" 2>/dev/null; then
      echo "$pid"
      break
    fi
  done
}

cleanup_and_summarize() {
  echo ""
  echo "Deteniendo soak test. Bajando topologia..."
  "$MULTICLIENT_SCRIPT" down || true

  if [ -f "$LOG_FILE" ] && [ "$(wc -l < "$LOG_FILE")" -gt 1 ]; then
    echo ""
    echo "Resumen (primera muestra vs ultima):"
    echo "--------------------------------------------------------"
    first=$(sed -n '2p' "$LOG_FILE")
    last=$(tail -n 1 "$LOG_FILE")
    echo "Primera: $first"
    echo "Ultima:  $last"
    echo ""
    echo "Log completo en: $LOG_FILE ($(wc -l < "$LOG_FILE") lineas, $(du -h "$LOG_FILE" | cut -f1))"
    echo "--------------------------------------------------------"
  fi
}
trap cleanup_and_summarize EXIT INT TERM

start_ts=$(date +%s)
end_ts=$((start_ts + DURATION_SEC))

while [ "$(date +%s)" -lt "$end_ts" ]; do
  now=$(date +%s)
  elapsed_min=$(( (now - start_ts) / 60 ))

  # Memoria RSS del proceso catsplash
  pid=$(get_catsplash_pid || true)
  if [ -n "${pid:-}" ] && [ -r "/proc/$pid/status" ]; then
    mem_kb=$(awk '/VmRSS/{print $2}' "/proc/$pid/status" 2>/dev/null || echo 0)
  else
    mem_kb=0
  fi

  # Tamano del archivo de base de datos
  db_size=0
  [ -f "$DB_PATH" ] && db_size=$(stat -c%s "$DB_PATH" 2>/dev/null || echo 0)

  # Conteo de reglas de iptables dentro de ns_router (senal de reglas huerfanas)
  nat_rules=$(ip netns exec ns_router iptables -t nat -S 2>/dev/null | wc -l || echo 0)
  filter_rules=$(ip netns exec ns_router iptables -S 2>/dev/null | wc -l || echo 0)

  # Estado de sesiones via catsctl, si esta disponible
  active=0; pending=0; total=0
  if [ -x "./bin/catsctl" ]; then
    status_out=$(./bin/catsctl status 2>/dev/null || true)
    active=$(echo "$status_out" | awk -F: '/Active Sessions/{gsub(/ /,"",$2); print $2}')
    pending=$(echo "$status_out" | awk -F: '/Pending Clients/{gsub(/ /,"",$2); print $2}')
    total=$(echo "$status_out" | awk -F: '/Total Clients/{gsub(/ /,"",$2); print $2}')
  fi

  echo "$(date -Iseconds),$elapsed_min,$mem_kb,$db_size,$nat_rules,$filter_rules,${active:-0},${pending:-0},${total:-0}" >> "$LOG_FILE"
  echo "[${elapsed_min}min] mem=${mem_kb}KB db=${db_size}B nat_rules=${nat_rules} filter_rules=${filter_rules}"

  # Trickle ligero: un solo cliente re-autentica para mantener actividad
  # real durante la noche usando el flujo completo del portal.
  cookie_jar=$(mktemp)
  portal_page=$(mktemp)

  ip netns exec ns_client1 curl -s -c "$cookie_jar" -b "$cookie_jar" \
    -o "$portal_page" "http://$ROUTER_AP_IP:$PORTAL_PORT/portal" >/dev/null 2>&1 || true

  nonce=$(ip netns exec ns_client1 sed -n 's/.*name="nonce" value="\([^"]*\)".*/\1/p' "$portal_page" | head -n 1 || true)
  if [ -n "$nonce" ]; then
    ip netns exec ns_client1 curl -s -o /dev/null \
      -b "$cookie_jar" -c "$cookie_jar" \
      -d "auth_data=true" -d "nonce=$nonce" \
      "http://$ROUTER_AP_IP:$PORTAL_PORT/auth" >/dev/null 2>&1 || true
  fi

  rm -f "$cookie_jar" "$portal_page"

  sleep "$INTERVAL_SEC"
done
