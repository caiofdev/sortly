#!/usr/bin/env bash
# Falha se a cobertura de testes dos pacotes em backend/ ficar abaixo do mínimo.
# Uso: scripts/check-coverage.sh [mínimo-em-%]   (padrão: 85)
set -euo pipefail

threshold="${1:-85}"
profile="coverage-backend.out"

# Erros do go list (ex.: go fora do PATH) interrompem o script via set -e;
# "matched no packages" é só um aviso e retorna vazio.
packages="$(go list ./backend/...)"
if [ -z "$packages" ]; then
  echo "Nenhum pacote em backend/ ainda; verificação de cobertura ignorada."
  exit 0
fi

go test -covermode=atomic -coverprofile="$profile" ./backend/...

total="$(go tool cover -func="$profile" | awk '/^total:/ { gsub("%", "", $3); print $3 }')"
echo "Cobertura de backend/: ${total}% (mínimo: ${threshold}%)"

if awk -v total="$total" -v min="$threshold" 'BEGIN { exit !(total + 0 < min + 0) }'; then
  echo "::error::Cobertura de backend/ (${total}%) abaixo do mínimo de ${threshold}%"
  exit 1
fi
