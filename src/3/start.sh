#!/usr/bin/env bash

# Запускать один раз на хосте (NixOS / Arch). Кто сломает - тот лох
# Добавляет demo-lecture.local в /etc/hosts и поднимает контейнер

set -e

HOSTS_ENTRY="127.0.0.1  demo-lecture.local"
HOSTS_FILE="/etc/hosts"

detect_distro() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "$ID"
  else
    echo "unknown"
  fi
}

DISTRO=$(detect_distro)

if grep -q "demo-lecture.local" "$HOSTS_FILE"; then
  echo "✓ demo-lecture.local уже есть в $HOSTS_FILE"
else
  case "$DISTRO" in
  nixos)
    echo ""
    echo "Ты на NixOS. Пшел нахер"
    read -rp "Продолжить без записи в hosts? (Y/n) " yn
    [[ "$yn" =~ ^([Yy]|)$ ]] || exit 1
    ;;
  arch | manjaro | endeavouros)
    echo "$HOSTS_ENTRY" | sudo tee -a "$HOSTS_FILE" >/dev/null
    echo "Добавлено в $HOSTS_FILE: $HOSTS_ENTRY"
    ;;
  *)
    echo "$HOSTS_ENTRY" | sudo tee -a "$HOSTS_FILE" >/dev/null
    echo "Добавлено в $HOSTS_FILE: $HOSTS_ENTRY"
    ;;
  esac
fi

echo "Собираю и запускаю контейнер..."
docker compose up --build -d

echo ""
echo "  Среда готова!"
echo ""
echo "  SSH:    ssh student@localhost -p 2222"
echo "  Пароль: sigma2025"
echo ""
echo "  Сайт (до починки):   http://demo-lecture.local:8080  →  пусто"
echo "  Сайт (после починки): http://demo-lecture.local:8080  →  200"
