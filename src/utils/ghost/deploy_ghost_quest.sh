#!/usr/bin/env bash
#   ./deploy_ghost_quest.sh <container_name> <student_login>
#   ./deploy_ghost_quest.sh sigma-stony-shaft stony-shaft
#
# Предполагается, что в контейнере уже доустановлены:
#   procps psmisc lsof strace binutils file

set -euo pipefail

CONTAINER="${1:?Использование: $0 <container_name> <student_login>}"
LOGIN="${2:?Использование: $0 <container_name> <student_login>}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
GHOSTD_BIN="$SCRIPT_DIR/ghostd/ghostd"
GHOSTCHECK_BIN="$SCRIPT_DIR/ghost-check/ghost-check"

LEGACY_DIR="/opt/.legacy-diag"

echo "==> Проверяю, что контейнер $CONTAINER запущен"
docker inspect -f '{{.State.Running}}' "$CONTAINER" | grep -q true

echo "==> Создаю $LEGACY_DIR"
docker exec -u root "$CONTAINER" mkdir -p "$LEGACY_DIR"

echo "==> Копирую бинарники призрака (ghostd играет роль и ghostd, и echod - см. main.go, режим переключается флагом --peer)"
docker cp "$GHOSTD_BIN" "$CONTAINER:$LEGACY_DIR/.ghostd"
docker cp "$GHOSTD_BIN" "$CONTAINER:$LEGACY_DIR/.echod"
docker exec -u root "$CONTAINER" chmod 755 "$LEGACY_DIR/.ghostd" "$LEGACY_DIR/.echod"
docker exec -u root "$CONTAINER" chown root:root "$LEGACY_DIR/.ghostd" "$LEGACY_DIR/.echod"

echo "==> Копирую diag.log (акт 1)"
docker cp "$TEMPLATES_DIR/diag.log" "$CONTAINER:$LEGACY_DIR/diag.log"

echo "==> Создаю защищённую папку с черновиком (акт 2)"
docker exec -u root "$CONTAINER" mkdir -p "$LEGACY_DIR/notice"
docker cp "$TEMPLATES_DIR/resignation_draft.txt" "$CONTAINER:$LEGACY_DIR/notice/draft.txt"

echo "==> Подделываю метки времени - вся папка должна выглядеть старше контейнера"
# День 11 в timestamp = фрагмент 1 (см. ghost-check/main.go)
docker exec -u root "$CONTAINER" touch -d "2024-03-11 03:14:00" "$LEGACY_DIR"
docker exec -u root "$CONTAINER" touch -d "2024-03-17 03:14:00" "$LEGACY_DIR/diag.log"
docker exec -u root "$CONTAINER" touch -d "2024-03-11 03:14:00" "$LEGACY_DIR/notice"
docker exec -u root "$CONTAINER" touch -d "2024-03-11 03:14:00" "$LEGACY_DIR/notice/draft.txt"

echo "==> Ставлю immutable-флаг на папку notice (акт 2 - rm/mv откажут без объяснений)"
docker exec -u root "$CONTAINER" chattr +i "$LEGACY_DIR/notice"

echo "==> Запускаю призрака (стартует ghostd, тот сам поднимет echod при первой проверке respawn)"
docker exec -u root -d "$CONTAINER" "$LEGACY_DIR/.ghostd"

echo "==> Настраиваю автозапуск призрака при перезапуске контейнера (@reboot в root-crontab)"
docker exec -u root "$CONTAINER" bash -c \
  "(crontab -l -u root 2>/dev/null; echo '@reboot $LEGACY_DIR/.ghostd') | crontab -u root -"

echo "==> Кладу ghost-check в PATH"
docker cp "$GHOSTCHECK_BIN" "$CONTAINER:/usr/local/bin/ghost-check"
docker exec -u root "$CONTAINER" chmod 755 /usr/local/bin/ghost-check

echo "==> Кладу личную записку от Володи в домашнюю папку ученика"
docker cp "$TEMPLATES_DIR/note_ghost.txt" "$CONTAINER:/home/$LOGIN/note_ghost.txt"
docker exec -u root "$CONTAINER" chown "$LOGIN:$LOGIN" "/home/$LOGIN/note_ghost.txt"

echo "==> Готово. Быстрая проверка:"
docker exec "$CONTAINER" bash -c "ps aux | grep -i 'kworker/R' | grep -v grep"
echo
echo "Если строки выше есть - призрак жив. Квест развёрнут для $LOGIN в $CONTAINER."
