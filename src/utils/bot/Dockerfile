FROM python:3.14-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY src ./src

# Чтобы БД переживала пересоздание контейнера
VOLUME ["/app/data"]

CMD ["python", "-m", "src.sigma_bot.main"]
