#!/bin/bash
#
# monitor_catsplash.sh
#
# Script externo y dedicado para espiar el uso de CPU y RAM de Catsplash
# mientras corre dentro del namespace de red 'ns_router'.

set -euo pipefail

# Función para descubrir el PID dinámicamente en el namespace
get_catsplash_pid() {
  ip netns pids ns_router 2>/dev/null | while read -r pid; do
    if [ -r "/proc/$pid/comm" ] && grep -q catsplash "/proc/$pid/comm" 2>/dev/null; then
      echo "$pid"
      return 0
    fi
  done
  echo ""
}

echo "🔍 Buscando proceso 'catsplash' dentro de ns_router..."
PID=$(get_catsplash_pid)

if [ -z "$PID" ]; then
  echo "❌ No se detectó Catsplash corriendo."
  echo "Asegúrate de que está activo en su terminal: sudo ip netns exec ns_router ./bin/catsplash"
  exit 1
fi

echo "🐱 ¡Proceso detectado con éxito!"
echo "📍 PID Monitoreado: $PID"
echo "--------------------------------------------------------"
printf "%-10s %-10s %-15s %-15s\n" "HORA" "% CPU" "RAM RSS (KB)" "RAM VIRT (KB)"
echo "--------------------------------------------------------"

# Ciclo de monitoreo en tiempo real
while true; do
  # Si Catsplash se muere, cerramos el monitor limpiamente
  if ! kill -0 "$PID" 2>/dev/null; then
    echo -e "\n💀 El proceso Catsplash (PID: $PID) se ha detenido."
    exit 0
  fi

  # 1. Medir CPU instantáneo (2 iteraciones con top separadas por 200ms)
  # LC_NUMERIC=C evita problemas si tu sistema usa comas para decimales
  CPU=$(LC_NUMERIC=C top -b -n 2 -d 0.2 -p "$PID" 2>/dev/null | awk -v pid="$PID" '$1 == pid {cpu=$9} END {print cpu ? cpu : "0.0"}')

  # 2. Medir RAM real y virtual directo desde /proc
  if [ -r "/proc/$PID/status" ]; then
    RAM_RSS=$(awk '/VmRSS/{print $2}' "/proc/$PID/status" 2>/dev/null || echo 0)
    RAM_VIRT=$(awk '/VmSize/{print $2}' "/proc/$PID/status" 2>/dev/null || echo 0)
  else
    RAM_RSS=0
    RAM_VIRT=0
  fi

  # Imprimir fila con formato limpio
  printf "%-10s %-10s %-15s %-15s\n" "$(date +%H:%M:%S)" "${CPU}%" "$RAM_RSS" "$RAM_VIRT"

  # Frecuencia de actualización en segundos (ajustable)
  sleep 2
done
