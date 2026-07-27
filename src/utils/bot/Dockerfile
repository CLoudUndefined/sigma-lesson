FROM python:3.14-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

# Чтобы БД переживала пересоздание контейнера
VOLUME ["/app/data"]
VOLUME ["/src"]

CMD ["python", "-m", "src.sigma_bot.main"]
