#!/usr/bin/env python3
"""
generate_mbox.py - генерирует правдоподобную историю писем в формате mbox
для /var/mail/<login>: десятки одинаковых "command not found" писем от
cron за последние недели, плюс одно другое по содержанию "canary"-письмо,
которое Володя отправил себе, когда только тестировал, что почта вообще
доходит.

Использование:
    python3 generate_mbox.py <login> <hostname> > seed_mailbox.mbox

Формат сообщений соответствует тому, что реально генерирует cron+sendmail
на Debian, чтобы файл выглядел подлинно при просмотре через cat/less/grep.
"""
import sys
from datetime import datetime, timedelta, timezone

WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun",
          "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]


def fmt_asctime(dt: datetime) -> str:
    # формат, который использует сама строка "From " в mbox (asctime-like)
    return f"{WEEKDAYS[dt.weekday()]} {MONTHS[dt.month-1]} {dt.day:2d} {dt.hour:02d}:{dt.minute:02d}:{dt.second:02d} {dt.year}"


def fmt_rfc822(dt: datetime) -> str:
    return f"{WEEKDAYS[dt.weekday()]}, {dt.day:02d} {MONTHS[dt.month-1]} {dt.year} {dt.hour:02d}:{dt.minute:02d}:{dt.second:02d} +0000"


def escape_body(body: str) -> str:
    # mbox-конвенция: строки тела, начинающиеся с "From ", экранируются ">From "
    lines = body.split("\n")
    return "\n".join(("&gt;" + l if l.startswith("From ") else l) for l in lines).replace("&gt;", ">")


def error_message(dt: datetime, login: str, host: str, msg_id: int) -> str:
    date_hdr = fmt_rfc822(dt)
    subject = f"Cron <{login}@{host}> disk-watch --threshold 80"
    body = (
        "/bin/sh: 1: disk-watch: not found\n"
    )
    return (
        f"From {login}@{host}  {fmt_asctime(dt)}\n"
        f"Return-Path: <{login}@{host}>\n"
        f"X-Original-To: {login}\n"
        f"Delivered-To: {login}@{host}\n"
        f"Received: by {host} (Postfix, from userid 0)\n"
        f"\tid {msg_id:08X}; {date_hdr}\n"
        f"From: root@{host} (Cron Daemon)\n"
        f"To: {login}@{host}\n"
        f"Subject: {subject}\n"
        f"Content-Type: text/plain; charset=UTF-8\n"
        f"X-Cron-Env: <SHELL=/bin/sh>\n"
        f"X-Cron-Env: <PATH=/usr/bin:/bin>\n"
        f"X-Cron-Env: <LOGNAME={login}>\n"
        f"Date: {date_hdr}\n"
        f"\n"
        f"{escape_body(body)}\n"
    )


def canary_message(dt: datetime, login: str, host: str, msg_id: int) -> str:
    date_hdr = fmt_rfc822(dt)
    subject = f"Cron <{login}@{host}> echo проверка почты"
    body = (
        "eto pervoe testovoe pismo, chtoby proverit chto pochta voobsche\n"
        "dohodit. esli ty eto chitaesh cherez mesyats - znachit ya tak i\n"
        "ne udosuzhilsya proverit, prishlo li ono.\n"
        "\n"
        "nomer proverki: 214.\n"
    )
    return (
        f"From {login}@{host}  {fmt_asctime(dt)}\n"
        f"Return-Path: <{login}@{host}>\n"
        f"X-Original-To: {login}\n"
        f"Delivered-To: {login}@{host}\n"
        f"Received: by {host} (Postfix, from userid 1000)\n"
        f"\tid {msg_id:08X}; {date_hdr}\n"
        f"From: {login}@{host} (Cron Daemon)\n"
        f"To: {login}@{host}\n"
        f"Subject: {subject}\n"
        f"Content-Type: text/plain; charset=UTF-8\n"
        f"X-Cron-Env: <SHELL=/bin/sh>\n"
        f"X-Cron-Env: <PATH=/usr/bin:/bin>\n"
        f"X-Cron-Env: <LOGNAME={login}>\n"
        f"Date: {date_hdr}\n"
        f"\n"
        f"{escape_body(body)}\n"
    )


def main():
    if len(sys.argv) < 3:
        print("Использование: generate_mbox.py <login> <hostname>", file=sys.stderr)
        sys.exit(1)

    login, host = sys.argv[1], sys.argv[2]

    now = datetime.now(timezone.utc).replace(tzinfo=None)
    # canary - самое раннее письмо, 5 недель назад, Володя тестировал почту
    # ДО того как поставил настоящую задачу с disk-watch
    canary_time = now - timedelta(weeks=5, hours=2)

    messages = []
    messages.append(canary_message(canary_time, login, host, 1))

    # дальше - недели молчаливых падений, примерно раз в 4 минуты, но чтобы
    # не генерировать буквально тысячи писем, представим репрезентативную
    # выборку - несколько писем в день на протяжении 5 недель
    msg_id = 2
    start = canary_time + timedelta(hours=6)
    t = start
    end = now - timedelta(hours=1)
    step = timedelta(hours=8)  # несколько штук в день, не буквально каждые 4 минуты
    while t < end:
        messages.append(error_message(t, login, host, msg_id))
        msg_id += 1
        t += step

    sys.stdout.write("".join(messages))


if __name__ == "__main__":
    main()
