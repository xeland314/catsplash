#!/bin/bash
#
# test_catsplash_multiclient.sh
#
# Simula N clientes Wi-Fi contra Catsplash usando network namespaces.
# Dividido en 3 fases independientes para evitar condiciones de carrera
# con el proceso de catsplash:
#
#   up    -> crea la topologia (router, wan, N clientes) y la deja viva
#   run   -> dispara autenticaciones en paralelo contra la topologia
#            YA EXISTENTE y guarda resultados en CSV
#   down  -> destruye la topologia
#
# Flujo correcto:
#   sudo ./test_catsplash_multiclient.sh up -n 50
#   # en otra terminal, DESPUES de que "up" termine:
#   sudo ip netns exec ns_router ./bin/catsplash
#   # de vuelta en la primera terminal:
#   sudo ./test_catsplash_multiclient.sh run
#   sudo ./test_catsplash_multiclient.sh down
#
set -euo pipefail

INT_AP="wlx1cbfce41183a"
INT_WAN="enp1s0"
ROUTER_AP_IP="192.168.6.1"
ROUTER_WAN_IP="192.168.1.100"
WAN_SRV_IP="192.168.1.1"
FAKE_INTERNET_IP="8.8.8.8"
PORTAL_PORT="8080"
CLIENT_SUBNET_BASE="192.168.6"
MAC_PREFIX="d2:d8:37:12"
RESULTS_ROOT="./results"

usage() {
  cat <<EOF
Uso: sudo $0 <comando> [opciones]

Comandos:
  up -n N       Crea router + wan + N clientes simulados (queda vivo)
  run           Autentica en paralelo TODOS los clientes ya creados por "up"
                y guarda resultados en $RESULTS_ROOT/<timestamp>/results.csv
  down          Elimina todos los namespaces de la simulacion

Flujo:
  sudo $0 up -n 50
  sudo ip netns exec ns_router ./bin/catsplash     # en otra terminal
  sudo $0 run
  sudo $0 down
EOF
}

if [ "$EUID" -ne 0 ]; then
  echo "Por favor, ejecuta este script con sudo."
  exit 1
fi

CMD="${1:-}"
shift || true

case "$CMD" in
  up)
    N=3
    while [ $# -gt 0 ]; do
      case "$1" in
        -n) N="$2"; shift 2 ;;
        *) echo "Opcion desconocida: $1"; usage; exit 1 ;;
      esac
    done

    if ! [[ "$N" =~ ^[0-9]+$ ]] || [ "$N" -lt 1 ] || [ "$N" -gt 200 ]; then
      echo "N debe ser un entero entre 1 y 200 (recibido: $N)"
      exit 1
    fi

    echo "Limpiando namespaces de corridas anteriores..."
    for ns in $(ip netns list 2>/dev/null | awk '{print $1}' | grep -E '^ns_(client[0-9]+|router|wan)$' || true); do
      ip netns del "$ns" 2>/dev/null || true
    done

    echo "Creando router y wan simulada..."
    ip netns add ns_router
    ip netns add ns_wan

    ip netns exec ns_router ip link add "$INT_AP" type bridge
    ip netns exec ns_router ip addr add "$ROUTER_AP_IP/24" dev "$INT_AP"
    ip netns exec ns_router ip link set "$INT_AP" up

    ip link add veth_rt_wan type veth peer name veth_wan_srv
    ip link set veth_rt_wan netns ns_router
    ip link set veth_wan_srv netns ns_wan

    ip netns exec ns_router ip link set veth_rt_wan name "$INT_WAN"
    ip netns exec ns_router ip addr add "$ROUTER_WAN_IP/24" dev "$INT_WAN"
    ip netns exec ns_router ip link set "$INT_WAN" up
    ip netns exec ns_router ip link set lo up
    ip netns exec ns_router sysctl -w net.ipv4.ip_forward=1 > /dev/null
    ip netns exec ns_router sysctl -w net.ipv4.conf.all.rp_filter=0 > /dev/null
    ip netns exec ns_router sysctl -w net.ipv4.conf."$INT_WAN".rp_filter=0 > /dev/null

    ip netns exec ns_wan ip addr add "$WAN_SRV_IP/24" dev veth_wan_srv
    ip netns exec ns_wan ip link set veth_wan_srv up
    ip netns exec ns_wan ip link set lo up
    ip netns exec ns_wan ip addr add "$FAKE_INTERNET_IP/32" dev lo
    ip netns exec ns_wan ip route add "$CLIENT_SUBNET_BASE.0/24" via "$ROUTER_WAN_IP"

    echo "Creando $N cliente(s)..."
    for i in $(seq 1 "$N"); do
      ns="ns_client${i}"
      veth_cli="vc${i}"
      veth_ap="va${i}"
      ip_octet=$((9 + i))
      client_ip="${CLIENT_SUBNET_BASE}.${ip_octet}"
      mac=$(printf "%s:%02x:%02x" "$MAC_PREFIX" $(((i>>8) & 0xFF)) $((i & 0xFF)))

      ip netns add "$ns"
      ip link add "$veth_cli" type veth peer name "$veth_ap"
      ip link set "$veth_cli" netns "$ns"
      ip link set "$veth_ap" netns ns_router

      ip netns exec "$ns" ip link set "$veth_cli" address "$mac"
      ip netns exec "$ns" ip addr add "$client_ip/24" dev "$veth_cli"
      ip netns exec "$ns" ip link set "$veth_cli" up
      ip netns exec "$ns" ip link set lo up
      ip netns exec "$ns" ip route add default via "$ROUTER_AP_IP"

      ip netns exec ns_router ip link set "$veth_ap" master "$INT_AP"
      ip netns exec ns_router ip link set "$veth_ap" up

      printf "  [%3d/%d] %-14s ip=%-16s mac=%s\n" "$i" "$N" "$ns" "$client_ip" "$mac"
    done

    echo ""
    echo "Topologia lista y VIVA (no se destruye al salir de este comando)."
    echo "--------------------------------------------------------"
    echo "Siguiente paso, en OTRA terminal:"
    echo "   sudo ip netns exec ns_router ./bin/catsplash"
    echo ""
    echo "Cuando este corriendo, vuelve aqui y ejecuta:"
    echo "   sudo $0 run"
    echo "--------------------------------------------------------"
    ;;

  run)
    mapfile -t clients < <(ip netns list 2>/dev/null | awk '{print $1}' | grep -E '^ns_client[0-9]+$' | sort -t t -k2 -n)
    if [ "${#clients[@]}" -eq 0 ]; then
      echo "No hay clientes simulados. Corre primero: sudo $0 up -n N"
      exit 1
    fi
    if ! ip netns list | grep -q '^ns_router'; then
      echo "No existe ns_router. Corre primero: sudo $0 up -n N"
      exit 1
    fi

    echo "Verificando que Catsplash responda a traves de un cliente real..."
    probe_ok=0
    for _ in $(seq 1 10); do
      if ip netns exec "${clients[0]}" curl -s -o /dev/null --connect-timeout 1 \
          "http://$ROUTER_AP_IP:$PORTAL_PORT/"; then
        probe_ok=1
        break
      fi
      sleep 0.5
    done
    if [ "$probe_ok" -eq 0 ]; then
      echo "Catsplash no responde en $ROUTER_AP_IP:$PORTAL_PORT."
      echo "Confirma que lo arrancaste DESPUES de 'up', dentro de ns_router:"
      echo "   sudo ip netns exec ns_router ./bin/catsplash"
      exit 1
    fi

    ts=$(date +%Y%m%d_%H%M%S)
    run_dir="$RESULTS_ROOT/$ts"
    mkdir -p "$run_dir"
    csv="$run_dir/results.csv"
    echo "client_id,http_code,time_total_s" > "$csv"

    echo "Lanzando ${#clients[@]} autenticaciones en paralelo..."
    pids=()
    for ns in "${clients[@]}"; do
      (
        cookie_jar=$(mktemp)
        portal_page=$(mktemp)

        ip netns exec "$ns" curl -s -c "$cookie_jar" -b "$cookie_jar" \
          -o "$portal_page" "http://$ROUTER_AP_IP:$PORTAL_PORT/portal" >/dev/null 2>&1 || true

        nonce=$(ip netns exec "$ns" sed -n 's/.*name="nonce" value="\([^"]*\)".*/\1/p' "$portal_page" | head -n 1 || true)

        if [ -n "$nonce" ]; then
          result=$(ip netns exec "$ns" curl -s -o /dev/null -w '%{http_code},%{time_total}' \
            -b "$cookie_jar" -c "$cookie_jar" \
            -d "auth_data=true" -d "nonce=$nonce" \
            "http://$ROUTER_AP_IP:$PORTAL_PORT/auth" 2>/dev/null || echo "000,0")
        else
          result="000,0"
        fi

        rm -f "$cookie_jar" "$portal_page"
        echo "${ns#ns_client},$result" >> "$csv"
      ) &
      pids+=($!)
    done
    for pid in "${pids[@]}"; do wait "$pid"; done

    echo ""
    echo "Resultados guardados en: $csv"
    echo "--------------------------------------------------------"
    tail -n +2 "$csv" | awk -F',' '
      {
        n++
        codes[$2]++
        times[n] = $3
        sum += $3
        if ($3 > max || n==1) max = $3
        if ($3 < min || n==1) min = $3
      }
      END {
        asort(times)
        p50 = times[int(n*0.50)+1]
        p95 = times[int(n*0.95)+1]
        p99 = times[int(n*0.99)+1]
        printf "  Peticiones:   %d\n", n
        printf "  Promedio:     %.4f s\n", sum/n
        printf "  Min / Max:    %.4f s / %.4f s\n", min, max
        printf "  p50 / p95 / p99: %.4f / %.4f / %.4f s\n", p50, p95, p99
        printf "  Codigos HTTP:\n"
        for (c in codes) printf "    %s -> %d\n", c, codes[c]
      }
    ' 2>/dev/null || echo "  (instala gawk para el resumen con percentiles; el CSV crudo ya esta guardado)"
    echo "--------------------------------------------------------"
    echo "Repite 'up -n' con distinto N (10, 25, 50, 100...) y vuelve a"
    echo "correr 'run' para construir una curva de latencia vs concurrencia."
    echo ""
    echo "Estado de NAT/reglas tras la corrida:"
    echo "   sudo ip netns exec ns_router iptables -t nat -L -n -v"
    ;;

  down)
    echo "Eliminando namespaces de la simulacion..."
    for ns in $(ip netns list 2>/dev/null | awk '{print $1}' | grep -E '^ns_(client[0-9]+|router|wan)$' || true); do
      ip netns del "$ns" 2>/dev/null || true
    done
    echo "Listo."
    ;;

  *)
    usage
    exit 1
    ;;
esac
