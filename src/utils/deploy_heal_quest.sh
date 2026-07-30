#!/usr/bin/env bash
#   ./deploy_heal_quest.sh <container_name> <student_login>
#   ./deploy_heal_quest.sh sigma-stony-shaft stony-shaft
## Доп-/гроб-задание занятия 4 ("Взбесившийся Смотритель"): у ученика уже
# должны быть на месте /var/www/demo (занятие 2, root:root) и его
# собственные backup.sh/cleanup.sh в ~/ (занятие 4, основная программа) -
# без них записка Володи не будет иметь смысла.
## Скрипт разворачивается и запускается ОТ ИМЕНИ САМОГО УЧЕНИКА
# (docker exec -u <login>), не root - чтобы kill/ps потом работали
# предсказуемо, без путаницы с правами. Права root ему для этого не
# нужны - NOPASSWD sudo уже настроен курсом в entrypoint.sh, self_heal.sh
# сам вызывает sudo точечно там, где нужно достучаться до /var/www/demo.

set -euo pipefail

CONTAINER="${1:?Использование: $0 <container_name> <student_login>}"
LOGIN="${2:?Использование: $0 <container_name> <student_login>}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
HEALCHECK_BIN="$SCRIPT_DIR/heal-check/heal-check"

STUDENT_HOME="/home/$LOGIN"
SELF_HEAL_PATH="/opt/self_heal/self_heal.sh"

echo "==> Проверяю, что контейнер $CONTAINER запущен"
docker inspect -f '{{.State.Running}}' "$CONTAINER" | grep -q true

echo "==> Проверяю, что /var/www/demo и ~/backups уже существуют (основная программа занятия 4)"
docker exec "$CONTAINER" test -d /var/www/demo
docker exec -u "$LOGIN" "$CONTAINER" bash -c "[ -d ~/backups ]" \
  || docker exec -u "$LOGIN" "$CONTAINER" mkdir -p "$STUDENT_HOME/backups"

echo "==> Кладу self_heal.sh в /opt/self_heal (не в домашнюю папку - это не ученический файл, это то, что оставил Володя)"
docker exec -u root "$CONTAINER" mkdir -p /opt/self_heal
docker cp "$TEMPLATES_DIR/self_heal.sh" "$CONTAINER:$SELF_HEAL_PATH"
docker exec -u root "$CONTAINER" chmod 755 "$SELF_HEAL_PATH"
docker exec -u root "$CONTAINER" chown "$LOGIN:$LOGIN" "$SELF_HEAL_PATH"

echo "==> Готовлю ~/var/www/demo/backups заранее, чтобы tar на первой итерации не падал по банальной причине 'папки нет'"
docker exec -u root "$CONTAINER" mkdir -p /var/www/demo/backups
docker exec -u root "$CONTAINER" chown -R root:root /var/www/demo

echo "==> Запускаю self_heal.sh от имени $LOGIN в фоне (nohup, как и было бы у самого Володи)"
docker exec -u "$LOGIN" -d "$CONTAINER" bash -c \
  "nohup bash $SELF_HEAL_PATH > /dev/null 2>&1 & disown"

echo "==> Кладу heal-check в PATH"
docker cp "$HEALCHECK_BIN" "$CONTAINER:/usr/local/bin/heal-check"
docker exec -u root "$CONTAINER" chmod 755 /usr/local/bin/heal-check

echo "==> Кладу личную записку от Володи в домашнюю папку ученика"
docker cp "$TEMPLATES_DIR/note_4b.txt" "$CONTAINER:$STUDENT_HOME/note_4b.txt"
docker exec -u root "$CONTAINER" chown "$LOGIN:$LOGIN" "$STUDENT_HOME/note_4b.txt"

echo "==> Жду 35 секунд, чтобы прошла хотя бы одна полная итерация цикла"
sleep 35

echo "==> Готово. Быстрая проверка:"
docker exec "$CONTAINER" bash -c "ps aux | grep -i 'self_heal.sh' | grep -v grep"
echo
docker exec "$CONTAINER" lsattr -R /var/www/demo 2>/dev/null | head -5
echo
echo "Если процесс выше есть, а lsattr показывает флаг i на /var/www/demo -"
echo "квест развёрнут для $LOGIN в $CONTAINER. Диск дальше будет расти сам,"
echo "пока ученик не найдёт и не остановит процесс."
