# GraphQL Forum

Современный форум, построенный с использованием GraphQL, Go и PostgresSQL. Проект предоставляет API для создания постов, комментариев и подписок на обновления в реальном времени.

## Функциональность

- Создание и получение постов
- Комментирование постов
- Древовидная структура комментариев
- Пагинация для постов и комментариев
- Подписки на новые комментарии в реальном времени
- Поддержка in-memory или PostgresSQL хранилища
- Тестирование работоспособности приложения

## Установка и запуск

### Локальный запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/Headliner38/graphql-forum.git
cd graphql-forum
```

2. Установите зависимости:
```bash
go mod download
```

3. Запустите приложение:
```bash
go run server.go
```

### Запуск с Docker

#### Продакшен окружение

1. Соберите и запустите контейнеры:
```bash
docker-compose up --build
```
Важно: сейчас контейнер запускается с использованием PostgresSQL хранилища.

Чтобы запустить контейнер с in-memory хранилищем:

- Перейдите в Dockerfile и измените 16 строку на `CMD ["./main"]`

Приложение будет доступно по адресу: http://localhost:8080

#### Тестовое окружение

1. Запустите тесты в Docker:
```bash
docker-compose -f docker-compose.test.yml up --build
```

## API

### Основные операции

#### Создание поста
```graphql
mutation {
  createPost(input: {
    title: "Заголовок поста"
    content: "Содержание поста"
    commentsEnabled: true # false, если не хотите, чтобы недоброжелатели оставляли комментарии под постом
  }) {
    id
    title
    content
    comments
  }
}
```

#### Создание комментария
```graphql
mutation {
  createComment(input: {
    postID: "post-id"
    text: "Текст комментария"
    parentCommId: "parent-comment-id" # если это ответ на существующий комментарий
  }) {
    id
    text
    postID
    parentCommID
  }
}
```

#### Получение постов
```graphql
query {
  posts {
    id
    title
    content
    comments
  }
}
```

#### Получение комментариев
```graphql
query {
  comments(postId: "post-id", limit: n, offset: n) {
    id
    text
    postID
    parentCommID
  }
}
```

#### Подписка на новые комментарии
```graphql
subscription {
  newComment(postId: "post-id") {
    id
    text
    postID
    parentCommID
  }
}
```

## Тестирование

### Запуск тестов в Docker

```bash
docker-compose -f docker-compose.test.yml up --build
```

## Конфигурация

### Переменные окружения

#### Продакшен
- `DB_HOST` - хост базы данных
- `DB_PORT` - порт базы данных
- `DB_USER` - пользователь базы данных
- `DB_PASSWORD` - пароль базы данных
- `DB_NAME` - имя базы данных
- `DB_SSLMODE` - режим SSL для подключения к БД

#### Тесты
- `TEST_DB_HOST` - хост тестовой базы данных
- `TEST_DB_PORT` - порт тестовой базы данных
- `TEST_DB_USER` - пользователь тестовой базы данных
- `TEST_DB_PASSWORD` - пароль тестовой базы данных
- `TEST_DB_NAME` - имя тестовой базы данных
- `TEST_DB_SSLMODE` - режим SSL для подключения к тестовой БД
