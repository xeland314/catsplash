#!/usr/bin/env bash
# LOPDP compliance check: ensures no plaintext MAC/IP in log statements.
# Fails if a log.Printf/log.Fatalf/fmt.Printf contains raw mac/ip variables
# without using maskMAC()/maskIP() for anonymization.
#
# Whitelist exceptions with "// lopdp:ignore" comment on the same line.
set -euo pipefail

VIOLATIONS=0

check_line() {
  local line="$1"
  local file="$2"
  local lineno="$3"

  # Skip lopdp:ignore
  if echo "$line" | grep -q '// lopdp:ignore'; then
    return
  fi

  # Skip if not a log/fmt print statement
  if ! echo "$line" | grep -qE '(log\.Printf|log\.Fatalf|fmt\.Printf|fmt\.Println|fmt\.Fatalf)'; then
    return
  fi

  # Extract only the format arguments (after the format string)
  local args
  args=$(echo "$line" | sed 's/.*"[^"]*"[[:space:]]*,[[:space:]]*//' 2>/dev/null || true)

  # If no args extracted (no comma after string), nothing to check
  if [ -z "$args" ]; then
    return
  fi

  # --- MAC check ---
  # Detect raw mac/MAC references NOT inside a maskMAC call
  # Step 1: Remove all maskMAC(...) calls from args
  local cleaned
  cleaned=$(echo "$args" | sed -E 's/maskMAC\([^)]*\)//g')
  # Step 2: Also remove state.MaskMAC(...) calls
  cleaned=$(echo "$cleaned" | sed -E 's/[a-zA-Z_]+\.MaskMAC\([^)]*\)//g')
  # Step 3: Check if raw mac/MAC/addr remains
  if echo "$cleaned" | grep -qE '(^|[^a-zA-Z_.])(mac|\.MAC|addr|\.Addr)([^a-zA-Z]|$)'; then
    echo "VIOLATION: Potential plaintext MAC in log at $file:$lineno"
    echo "  -> $line"
    VIOLATIONS=$((VIOLATIONS + 1))
    return
  fi

  # --- IP check ---
  # Detect raw ip/IP/RemoteAddr references NOT inside a maskIP call
  cleaned=$(echo "$args" | sed -E 's/maskIP\([^)]*\)//g')
  cleaned=$(echo "$cleaned" | sed -E 's/[a-zA-Z_]+\.MaskIP\([^)]*\)//g')
  if echo "$cleaned" | grep -qE '(^|[^a-zA-Z_.])(ip|\.IP)([^a-zA-Z]|$)'; then
    echo "VIOLATION: Potential plaintext IP in log at $file:$lineno"
    echo "  -> $line"
    VIOLATIONS=$((VIOLATIONS + 1))
    return
  fi
  if echo "$cleaned" | grep -qE 'RemoteAddr'; then
    echo "VIOLATION: Potential plaintext IP in log at $file:$lineno"
    echo "  -> $line"
    VIOLATIONS=$((VIOLATIONS + 1))
    return
  fi
}

# Find all log/print statements in Go files, excluding test files and vendor
while IFS= read -r match; do
  file=$(echo "$match" | cut -d: -f1)
  lineno=$(echo "$match" | cut -d: -f2)
  content=$(echo "$match" | cut -d: -f3-)
  check_line "$content" "$file" "$lineno"
done < <(grep -rn -E '(log\.Printf|log\.Fatalf|fmt\.Printf|fmt\.Println|fmt\.Fatalf)' --include='*.go' . | grep -v '_test\.go' | grep -v 'vendor/')

if [ $VIOLATIONS -gt 0 ]; then
  echo ""
  echo "FAIL: Found $VIOLATIONS LOPDP compliance violation(s)."
  echo "Use maskMAC() for MAC addresses and maskIP() for IP addresses."
  echo "Add '// lopdp:ignore' comment if the line is intentional."
  exit 1
fi

echo "PASS: No plaintext MAC/IP found in log statements."
