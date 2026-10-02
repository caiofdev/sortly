#!/usr/bin/env bash
# Falha se a cobertura de testes dos pacotes em internal/ ficar abaixo do mínimo.
# Uso: scripts/check-coverage.sh [mínimo-em-%]   (padrão: 85)
set -euo pipefail

threshold="${1:-85}"
profile="coverage-internal.out"

packages="$(go list ./internal/... 2>/dev/null || true)"
if [ -z "$packages" ]; then
  echo "Nenhum pacote em internal/ ainda; verificação de cobertura ignorada."
  exit 0
fi

go test -covermode=atomic -coverprofile="$profile" ./internal/...

total="$(go tool cover -func="$profile" | awk '/^total:/ { gsub("%", "", $3); print $3 }')"
echo "Cobertura de internal/: ${total}% (mínimo: ${threshold}%)"

if awk -v total="$total" -v min="$threshold" 'BEGIN { exit !(total + 0 < min + 0) }'; then
  echo "::error::Cobertura de internal/ (${total}%) abaixo do mínimo de ${threshold}%"
  exit 1
fi
