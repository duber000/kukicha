#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ZED_DIR=$(cd "${SCRIPT_DIR}/.." && pwd)

SOURCE="${ZED_DIR}/languages/kukicha/highlights.scm"
TARGET="${ZED_DIR}/grammars/kukicha/queries/highlights.scm"

if cmp -s "${SOURCE}" "${TARGET}"; then
    echo "Highlight queries are in sync."
    exit 0
fi

echo "Highlight queries are out of sync:"
echo "- ${SOURCE}"
echo "- ${TARGET}"
echo "Run: editors/zed/scripts/sync-highlights.sh"
exit 1
