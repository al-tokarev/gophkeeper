# GophKeeper

Менеджер паролей: клиент-серверное приложение для хранения логинов/паролей, текстов, бинарных данных и карт.

## Что нужно

- Docker и Docker Compose

## Запуск сервера

Сервер и PostgreSQL поднимаются одной командой:

```bash
docker compose up --build
```

При старте:
- поднимается PostgreSQL (`gophkeeper`/`gophkeeper`, БД `gophkeeper`) на `localhost:5432`;
- сервер ждёт, пока БД станет доступна, применяет миграции и слушает `http://localhost:8080`.

Остановить:

```bash
docker compose down
```

Остановить и удалить данные БД:

```bash
docker compose down -v
```

## Сборка клиента

Клиент — отдельный бинарник, в Docker не упакован (по ТЗ он распространяется как исполняемый файл под Windows/Linux/macOS).

```bash
go build -o gophkeeper ./cmd/cli
```

Сервер по умолчанию — `http://localhost:8080`. Если нужно другой адрес:

```bash
GOPHKEEPER_SERVER=http://127.0.0.1:9090 ./gophkeeper list
```

## Проверка

Регистрация:

```bash
curl -i -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"secret"}'
```

Ожидаемо: `201 Created` и JSON с `id`, `email`.

Логин:

```bash
curl -s -X POST http://localhost:8080/users/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"secret"}'
```

Ожидаемо: `{"token":"..."}`.

Создание записи (нужно подставить токен):

```bash
curl -i -X POST http://localhost:8080/records \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"type":"login","name":"gmail","data":{"login":"a","password":"b"},"metadata":"mail"}'
```

Ожидаемо: `201 Created`.

Список записей:

```bash
curl -i http://localhost:8080/records -H "Authorization: Bearer <token>"
```

Ожидаемо: `200 OK` и JSON-массив.

## Через клиент

```bash
# регистрация
./gophkeeper register --email=a@b.c --password=secret

# логин
./gophkeeper login --email=a@b.c --password=secret

# добавить запись
./gophkeeper add --type=login --name=gmail \
  --data='{"login":"a","password":"b"}' --metadata=mail

# список
./gophkeeper list

# получить по id
./gophkeeper get --id=<id>

# обновить
./gophkeeper update --id=<id> --name=gmail \
  --data='{"login":"a","password":"NEW"}' --metadata=mail

# удалить
./gophkeeper delete --id=<id>
```

## Переменные окружения сервера

| Переменная | Описание |
|---|---|
| `DATABASE_DSN` | DSN PostgreSQL (обязательна) |
| `HTTP_ADDR` | адрес сервера, по умолчанию `localhost:8080` |
| `SECRET_KEY` | base64 от 32 байт, ключ шифрования AES-256 |
| `JWT_SECRET` | строка ≥ 32 байт для подписи JWT |
| `JWT_EXPIRY` | время жизни токена, например `24h` |

В `docker-compose.yml` значения уже прописаны для локального запуска. Для продакшена их нужно заменить.