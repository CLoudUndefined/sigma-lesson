#!/usr/bin/env bash
# deploy_mail_quest.sh
#
# Разворачивает побочный квест "Тень в расписании" (занятие 5, cron + >>)
# в ОДНОМ конкретном контейнере.
#
# ВАЖНО: этот скрипт устанавливает Postfix внутри контейнера. Предполагается,
# что .deb-пакеты Postfix со всеми зависимостями уже скачаны на хосте через:
#
#   mkdir -p /tmp/postfix-payload && cd /tmp/postfix-payload
#   apt-get update
#   apt-get install --download-only -y --no-install-recommends postfix
#   cp /var/cache/apt/archives/*.deb .
#
# Использование:
#   ./deploy_mail_quest.sh <container_name> <student_login> <path_to_postfix_debs>
#
# Пример:
#   ./deploy_mail_quest.sh sigma-stony-shaft stony-shaft /tmp/postfix-payload

set -euo pipefail

CONTAINER="${1:?Использование: $0 <container_name> <student_login> <postfix_debs_dir>}"
LOGIN="${2:?Использование: $0 <container_name> <student_login> <postfix_debs_dir>}"
POSTFIX_DEBS="${3:?Использование: $0 <container_name> <student_login> <postfix_debs_dir>}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
MAILAUDIT_BIN="$SCRIPT_DIR/mail-audit/mail-audit"

echo "==> Проверяю, что контейнер $CONTAINER запущен"
docker inspect -f '{{.State.Running}}' "$CONTAINER" | grep -q true

HOSTNAME_IN_CONTAINER=$(docker exec "$CONTAINER" hostname)

echo "==> Устанавливаю Postfix (Local only, неинтерактивно)"
docker cp "$POSTFIX_DEBS" "$CONTAINER:/tmp/postfix-payload"
docker exec -u root "$CONTAINER" bash -c \
  "echo 'postfix postfix/main_mailer_type select Local only' | debconf-set-selections"
docker exec -u root "$CONTAINER" bash -c \
  "echo 'postfix postfix/mailname string ${HOSTNAME_IN_CONTAINER}' | debconf-set-selections"
docker exec -u root "$CONTAINER" bash -c \
  "DEBIAN_FRONTEND=noninteractive dpkg -i /tmp/postfix-payload/*.deb || true"
# второй проход - доконфигурировать пакеты, оставшиеся 'unconfigured' из-за
# порядка зависимостей (та же история, что мы уже проходили с strace/lsof)
docker exec -u root "$CONTAINER" bash -c \
  "DEBIAN_FRONTEND=noninteractive dpkg -i /tmp/postfix-payload/*.deb"

echo "==> Добавляю Postfix в supervisord"
docker exec -u root "$CONTAINER" bash -c 'cat > /etc/supervisor/conf.d/postfix.conf << "EOF"
[program:postfix]
command=/usr/sbin/postfix start-fg
autostart=true
autorestart=true
EOF'
docker exec -u root "$CONTAINER" supervisorctl reread
docker exec -u root "$CONTAINER" supervisorctl update

echo "==> Проверяю, что Postfix слушает только loopback (не 0.0.0.0)"
sleep 2
docker exec "$CONTAINER" ss -tlnp 2>/dev/null | grep ':25' || echo "    (не увидел строку с :25 - проверь вручную)"

echo "==> Кладу disk-watch в личную папку ученика"
docker exec -u root "$CONTAINER" mkdir -p "/home/$LOGIN/.local/bin"
docker cp "$TEMPLATES_DIR/disk-watch" "$CONTAINER:/home/$LOGIN/.local/bin/disk-watch"
docker exec -u root "$CONTAINER" chmod 755 "/home/$LOGIN/.local/bin/disk-watch"
docker exec -u root "$CONTAINER" chown -R "$LOGIN:$LOGIN" "/home/$LOGIN/.local"

echo "==> Патчу .bashrc (добавляю ~/.local/bin в PATH)"
docker cp "$TEMPLATES_DIR/bashrc_snippet.sh" "$CONTAINER:/tmp/bashrc_snippet.sh"
docker exec -u root "$CONTAINER" bash -c "cat /tmp/bashrc_snippet.sh >> /home/$LOGIN/.bashrc"

echo "==> Добавляю сломанную строку в crontab ученика"
docker exec -u "$LOGIN" "$CONTAINER" bash -c \
  "(crontab -l 2>/dev/null; echo '*/4 * * * * disk-watch --threshold 80') | crontab -"

echo "==> Генерирую и сею историю писем в /var/mail/$LOGIN"
python3 "$TEMPLATES_DIR/generate_mbox.py" "$LOGIN" "$HOSTNAME_IN_CONTAINER" > /tmp/seed_mailbox.mbox
docker cp /tmp/seed_mailbox.mbox "$CONTAINER:/var/mail/$LOGIN"
docker exec -u root "$CONTAINER" chown "$LOGIN:mail" "/var/mail/$LOGIN"
docker exec -u root "$CONTAINER" chmod 660 "/var/mail/$LOGIN"
rm -f /tmp/seed_mailbox.mbox

echo "==> Кладу mail-audit в PATH"
docker cp "$MAILAUDIT_BIN" "$CONTAINER:/usr/local/bin/mail-audit"
docker exec -u root "$CONTAINER" chmod 755 /usr/local/bin/mail-audit

echo "==> Кладу личную записку от Володи"
docker cp "$TEMPLATES_DIR/note_mail.txt" "$CONTAINER:/home/$LOGIN/note_mail.txt"
docker exec -u root "$CONTAINER" chown "$LOGIN:$LOGIN" "/home/$LOGIN/note_mail.txt"

echo
echo "==> Готово. Быстрая проверка:"
echo "    docker exec $CONTAINER wc -l /var/mail/$LOGIN"
echo "    docker exec -u $LOGIN $CONTAINER crontab -l"
echo "    docker exec $CONTAINER ss -tlnp | grep :25   (должен быть 127.0.0.1:25, не 0.0.0.0:25)"
