FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Установка wait-for-it
RUN apt-get update && apt-get install -y wait-for-it

# Используем wait-for-it для ожидания готовности базы данных
CMD ["wait-for-it", "db:5432", "--", "./main", "--storage=postgres", "--pg-conn=postgres://postgres:3276@db:5432/postgres?sslmode=disable"] 