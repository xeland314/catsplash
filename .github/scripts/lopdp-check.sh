#!/usr/bin/env bash
# LOPDP compliance check: ensures no plaintext MAC/IP in log statements.
# Fails if a log.Printf/log.Fatalf/fmt.Printf contains raw mac/ip variables
# without using maskMAC()/maskIP() for anonymization.
#
# Whitelist exceptions with "// lopdp:ignore" comment on the same line.
set -euo pipefail

VIOLATIONS=0

# Find all log/print statements in Go files, excluding test files and vendor
while IFS= read -r line; do
  # Skip test files
  if echo "$line" | grep -qE '_test\.go:'; then
    continue
  fi

  # Check if line has lopdp:ignore comment
  if echo "$line" | grep -qE '// lopdp:ignore'; then
    continue
  fi

  # Extract file:line
  file=$(echo "$line" | cut -d: -f1)
  lineno=$(echo "$line" | cut -d: -f2)

  # Check for log/fmt statements
  if ! echo "$line" | grep -qE '(log\.Printf|log\.Fatalf|fmt\.Printf|fmt\.Println|fmt\.Fatalf)'; then
    continue
  fi

  # Extract just the arguments after the format string to avoid false positives
  # from variable names inside function calls (e.g., maskIP(getIPFromRemoteAddr(...)))
  args=$(echo "$line" | sed 's/.*fmt\.\(Printf\|Println\|Fatalf\)\(.*\)/\2/' | sed 's/.*log\.\(Printf\|Fatalf\)\(.*\)/\2/')

  # MAC leak detection: look for raw mac/MAC/addr variable references
  # Exclude when preceded by maskMAC( or inside string literals
  mac_raw=false
  if echo "$args" | grep -qE '(^|[^a-zA-Z])(mac|\.MAC)([^a-zA-Z]|$)'; then
    # Check it's not inside maskMAC() call
    if ! echo "$args" | grep -qE 'maskMAC\('; then
      mac_raw=true
    fi
  fi
  if echo "$args" | grep -qE '(^|[^a-zA-Z])(addr|\.Addr)([^a-zA-Z]|$)'; then
    if ! echo "$args" | grep -qE 'maskMAC\('; then
      mac_raw=true
    fi
  fi

  # IP leak detection: look for raw ip/IP/RemoteAddr variable references
  # Exclude when preceded by maskIP( or inside string literals
  ip_raw=false
  if echo "$args" | grep -qE '(^|[^a-zA-Z])(ip|\.IP)([^a-zA-Z]|$)'; then
    if ! echo "$args" | grep -qE 'maskIP\('; then
      ip_raw=true
    fi
  fi
  if echo "$args" | grep -qE 'RemoteAddr'; then
    if ! echo "$args" | grep -qE 'maskIP\('; then
      ip_raw=true
    fi
  fi

  if [ "$mac_raw" = true ]; then
    echo "VIOLATION: Potential plaintext MAC in log at $file:$lineno"
    echo "  -> $line"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi

  if [ "$ip_raw" = true ]; then
    echo "VIOLATION: Potential plaintext IP in log at $file:$lineno"
    echo "  -> $line"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(grep -rn -E '(log\.Printf|log\.Fatalf|fmt\.Printf|fmt\.Println|fmt\.Fatalf)' --include='*.go' . | grep -v '_test\.go' | grep -v 'vendor/')

if [ $VIOLATIONS -gt 0 ]; then
  echo ""
  echo "FAIL: Found $VIOLATIONS LOPDP compliance violation(s)."
  echo "Use maskMAC() for MAC addresses and maskIP() for IP addresses."
  echo "Add '// lopdp:ignore' comment if the line is intentional."
  exit 1
fi

echo "PASS: No plaintext MAC/IP found in log statements."
