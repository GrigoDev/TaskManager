# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o taskmanager

# Финальный этап
FROM ubuntu:latest

WORKDIR /app

# Устанавливаем SQLite
RUN apt-get update && apt-get install -y sqlite3 && rm -rf /var/lib/apt/lists/*

# Копируем бинарный файл и директорию web из этапа сборки
COPY --from=builder /app/taskmanager .
COPY --from=builder /app/web ./web

# Открываем порт
EXPOSE 7540

# Устанавливаем переменные окружения
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

# Создаем директорию для данных и устанавливаем права
RUN mkdir -p /data && chmod 777 /data

# Запускаем приложение
CMD ["./taskmanager"] 