# Под капотом сервера: Bash и системное администрирование

Telegram-бот для работы учеников на курсе

### Развертывание

1. Создать `.env` файл:

```
BOT_TOKEN=YOUR_BOT_TOKEN_HERE
ADMIN_IDS=YOUR_ADMIN_IDS_HERE
COURSE_CHAT_LINK=YOUR_INVITE_LINK_HERE
```

2. Подготовить список даных для входа

Создать файл `data/students.csv` со следующими столбцами:

```csv
login, password, port, domain
```
### Запуск

```bash
python main.py
```
