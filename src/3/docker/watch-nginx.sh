#!/usr/bin/env bash

# Как только там появляется файл - перезапускаем nginx, иначе не взлетит

WATCH_DIR="/etc/nginx/sites-enabled"

echo "[watcher] Слежу за $WATCH_DIR ..."

inotifywait -m -e create -e moved_to "$WATCH_DIR" |
  while read -r dir event file; do
    echo "[watcher] Появился файл: $file — перезапускаю nginx"
    nginx -t 2>/dev/null && nginx -s reload
    echo "[watcher] nginx перезапущен"
  done
