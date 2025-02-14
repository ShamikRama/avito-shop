# ВНИМАНИЕ - собрать файл находясь в корне проекта с помощью docker build -t avitoservice .
# Используем базовый образ Go
FROM golang:1.24-alpine

# Устанавливаем рабочую директорию
WORKDIR /avito-shop

# Копируем только go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем содержимое папки cmd/ в рабочую директорию
COPY cmd/ ./cmd/

# Копируем остальные необходимые файлы (например, internal/)
COPY internal/ ./internal/

# Копируем содержимое папки pkg/ в рабочую директорию
COPY migrations/ ./migrations/

# Копируем .env файл в контейнер
COPY .env .

# Собираем приложение
RUN go build -o avitoservice ./cmd/

# Устанавливаем goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Устанавливаем wait-for-it
RUN apk add --no-cache curl && \
    curl -L https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh -o /usr/local/bin/wait-for-it && \
    chmod +x /usr/local/bin/wait-for-it

# Устанавливаем bash
RUN apk add --no-cache bash

# Добавляем /usr/local/bin в PATH
ENV PATH="/usr/local/bin:${PATH}"

# Указываем переменную окружения для пути к конфигу
ENV CONFIG_PATH=/avito-shop/.env

# Команда для запуска приложения
CMD ["./avitoservice"]