# Занятие 3 - «Ищейка»
## Быстрый старт (NixOS / Arch / любой Linux с Docker)

```bash
chmod +x start.sh
./start.sh
```

Все, контейнер соберётся, файлы разложатся, nginx поднимется (но сайт
будет недоступен до тех пор, пока ученик не вернёт конфиг на место).

## Подключение

```bash
ssh student@localhost -p 2222
# пароль: sigma2025
```

## Что внутри контейнера

```
~/sysadmin/
    note.txt        - записка от Володи
    scammer.txt     - список подозрительных IP
    access.log      - ~20 000 строк, следы взломщика в конце

/tmp/.hidden-secret/
    nginx.conf      - спрятанный конфиг (ученик должен найти и вернуть)

/etc/nginx/sites-enabled/
    (пусто)         - сюда нужно переместить nginx.conf

/var/www/demo/
    index.html      - сайт сервера
```

## Сценарий финала

Ученик выполняет:
```bash
mv /tmp/.hidden-secret/nginx.conf /etc/nginx/sites-enabled/
```

Watcher внутри контейнера замечает новый файл и автоматически делает
`nginx -s reload`. Через секунду сайт открывается на:

```
http://demo-lecture.local:8080
```

Никаких sudo, никаких дополнительных команд - просто mv и магия от Тимы.

## Остановить контейнер

```bash
docker compose down
```

## Пересобрать с нуля (если что-то поломалось)

```bash
docker compose down
docker compose up --build -d
```

## Злоумышленник в логах

IP: `77.88.44.242` (есть в scammer.txt - IP Яндекса, могут
проверить через whois если кому интересно)

Следы в access.log:
- ~200 запросов в конце лога (последние 2 часа)
- лазил по /etc/nginx/*, /tmp/.hidden-secret/
- user-agent: python-requests (отсылка на то, что делал не человек, а скрипт)

Ключевая строка которую найдёт grep:
```
77.88.44.242 ... "GET /tmp/.hidden-secret/nginx.conf HTTP/1.1" ...
```
