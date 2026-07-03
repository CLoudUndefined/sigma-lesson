#!/usr/bin/env bash

set -e

HOME_DIR="/home/student"
SYSADMIN_DIR="$HOME_DIR/sysadmin"
HIDDEN_DIR="/tmp/.hidden-secret"
WEBROOT="/var/www/demo"

mkdir -p "$WEBROOT"
cat >"$WEBROOT/index.html" <<'EOF'
<!doctype html><html lang=ru><meta charset=UTF-8><meta name=viewport content="width=device-width,initial-scale=1"><title>Очень важный сервер</title>
<link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500&display=swap" rel=stylesheet><style>body{background:#f5f6fa;color:#202124;font-family:roboto,sans-serif;display:flex;flex-direction:column;align-items:center;justify-content:center;height:100vh;margin:0}.box{background:#fff;padding:32px 40px;border-radius:16px;box-shadow:0 4px 20px rgba(0,0,0,5%);text-align:center;margin-bottom:24px;max-width:90%}h1{color:#1a73e8;font-size:2.2rem;margin:0 0 8px;font-weight:500}p{color:#5f6368;margin:0 0 24px;font-size:1.1rem}.banner{background:#fff3e0;color:#e65100;padding:12px 24px;border-radius:12px;font-size:1rem;font-weight:500;display:inline-block;border:1px solid #ffe0b2}.fire-gif{width:64px;height:auto;image-rendering:pixelated}</style><div class=box><h1>Очень важный сервер</h1><p>Добро пожаловать<div class=banner>Я восстал из пепла и ядерного огня</div></div><img class=fire-gif src=https://media.tenor.com/TpZ5dQKpk7YAAAAi/fire-minecraft.gif alt="Пиксельный огонь">
EOF

mkdir -p "$HIDDEN_DIR"
cat >"$HIDDEN_DIR/nginx.conf" <<'EOF'
server {
    listen 80 default_server;
    server_name demo-lecture.local localhost;

    root /var/www/demo;
    index index.html;

    access_log /var/log/nginx/access.log;
    error_log  /var/log/nginx/error.log;

    location / {
        try_files $uri $uri/ =404;
    }
}
EOF
chmod 644 "$HIDDEN_DIR/nginx.conf"

rm -f /etc/nginx/sites-enabled/default

mkdir -p "$SYSADMIN_DIR"

cat >"$SYSADMIN_DIR/note.txt" <<'EOF'
Привет, подаван.

Ночью был подозрительный вход. Кажется к нам снова завалился тот тип,
который кошмарит наш сервер запросами уже вторую неделю.

Я собрал список странных IP за последнее время в scammer.txt - проверь, может кто-то
из них таки смог пробраться. Логи nginx лежат тут же, в access.log.

Если найдёшь виновника - сохрани отчёт для меня в report.txt,
я вернусь с конференции вечером.

Удачи тебе
EOF

cat >"$SYSADMIN_DIR/scammer.txt" <<'EOF'
# Список подозрительных IP - собрал ©Володя (спасибо мне)
# Эти адреса замечены в попытках взлома наших серверов

185.220.101.45
45.142.212.100
77.88.44.242
194.165.16.77
91.108.4.0
103.124.106.25
162.142.125.0
EOF

python3 <<'PYEOF'
import random
from datetime import datetime, timedelta

NORMAL_IPS = [
    "95.165.12.44", "188.243.56.71", "46.0.234.12",
    "213.87.145.33", "178.65.99.201", "109.252.44.18",
    "5.138.201.77",  "31.173.82.119", "176.59.44.200",
]
BAD_IP = "77.88.44.242"

NORMAL_PATHS = [
    "/", "/index.html", "/about", "/contact",
    "/static/style.css", "/static/logo.png",
    "/favicon.ico", "/robots.txt",
]
NGINX_PATHS = [
    "/etc/nginx/sites-enabled/demo.conf",
    "/etc/nginx/nginx.conf",
    "/etc/nginx/sites-available/",
    "/tmp/.hidden-secret/nginx.conf",
]
BAD_PATHS = NGINX_PATHS + [
    "/admin", "/.env", "/wp-login.php", "/config.php",
]

METHODS  = ["GET"] * 8 + ["POST"] * 2
STATUSES = [200, 200, 200, 200, 301, 304, 404, 500]
UA_NORMAL = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
]
UA_BAD = "python-requests/2.28.0"

start = datetime.now() - timedelta(hours=24)
lines = []

# 19 800 обычных запросов за 24 часа
for i in range(19800):
    t   = start + timedelta(seconds=i * 4 + random.randint(0, 3))
    ip  = random.choice(NORMAL_IPS)
    m   = random.choice(METHODS)
    p   = random.choice(NORMAL_PATHS)
    s   = random.choice(STATUSES)
    sz  = random.randint(200, 8000)
    ua  = random.choice(UA_NORMAL)
    ts  = t.strftime("%d/%b/%Y:%H:%M:%S +0300")
    lines.append(f'{ip} - - [{ts}] "{m} {p} HTTP/1.1" {s} {sz} "-" "{ua}"')

# ~200 запросов от взломщика - в конце лога (последние 2 часа)
attack_start = datetime.now() - timedelta(hours=2)
for i, path in enumerate(BAD_PATHS * 30):
    t  = attack_start + timedelta(seconds=i * 22 + random.randint(0, 10))
    m  = "GET"
    s  = random.choice([200, 403, 404])
    sz = random.randint(100, 2000)
    ts = t.strftime("%d/%b/%Y:%H:%M:%S +0300")
    lines.append(f'{BAD_IP} - - [{ts}] "{m} {path} HTTP/1.1" {s} {sz} "-" "{UA_BAD}"')

lines.sort(key=lambda l: l.split("[")[1].split("]")[0])

with open("/home/student/sysadmin/access.log", "w") as f:
    f.write("\n".join(lines) + "\n")

print(f"access.log: {len(lines)} строк")
PYEOF

chown -R student:student "$SYSADMIN_DIR"
chown -R student:student "$HIDDEN_DIR"
chown student:student "$WEBROOT/index.html"

echo "Среда для занятия 'Ищейка' готова"
echo "  ~/sysadmin/note.txt"
echo "  ~/sysadmin/scammer.txt"
echo "  ~/sysadmin/access.log  (~20 000 строк)"
echo "  /tmp/.hidden-secret/nginx.conf"
echo "  /var/www/demo/index.html"
