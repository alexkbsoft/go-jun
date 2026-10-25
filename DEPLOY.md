# Деплой на прод

## Схема

```
интернет → nginx (VPS, порт 80) → 127.0.0.1:8090 → контейнер api → контейнер postgres
```

Всё, кроме nginx, работает в Docker Compose. Наружу открыт только nginx. Порт сервиса доступен лишь локально на сервере, порт Postgres не опубликован.

## Файлы в репозитории

| Файл | Назначение |
|---|---|
| `Dockerfile` | Образ Go-сервиса. Двухэтапная сборка: `golang:1.27-alpine3.23` собирает бинарник, `alpine:3.24` запускает его от пользователя без прав root. Порт 8090. |
| `Dockerfile.migrate` | Образ миграций: `migrate/migrate` плюс SQL-файлы из `migrations/`, скопированные внутрь. |
| `migrations/*.up.sql` | Миграции схемы: `01_init` (таблица `lists`), `02_create_tasks` (таблица `tasks`, связь с `lists`). Применённые файлы не редактируем, только добавляем новые. |
| `docker-compose.prod.yml` | Описание боевого окружения (подробнее ниже). |
| `.dockerignore` | Не пускает в образы `.git`, `.env*` и служебные файлы. |
| `.gitignore` | Исключает из git `.env` и `.env.prod`. |

### docker-compose.prod.yml

Проект называется `tasklist-prod`, поэтому тома и сеть не пересекаются с окружением для разработки. Сервисы стартуют по цепочке:

1. **`postgres`**: образ `postgres:18.6`, данные в томе `postgres_data`, проверка готовности (`pg_isready`), порт наружу не открыт. База создаётся сама из переменной `POSTGRES_DB`.
2. **`migrate`**: одноразовый контейнер, выполняет `migrate up` после готовности базы и завершается. Состояние хранится в таблице `schema_migrations` внутри базы.
3. **`api`**: стартует после успешных миграций, порт `127.0.0.1:8090:8090`, `restart: unless-stopped`. Адрес базы получает через переменную `API_SERVER_DB_URL`.

## Файлы только на сервере

| Файл | Назначение |
|---|---|
| `.env.prod` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`. Пароль генерируется командой `openssl rand -hex 24`. Права `chmod 600`, в git не попадает. |
| `/etc/nginx/sites-available/tasklist` | Конфигурация nginx: `server_name`, `proxy_pass http://127.0.0.1:8090` и заголовки проксирования. |
| `/etc/nginx/sites-enabled/tasklist` | Символическая ссылка на файл выше. Без неё nginx конфигурацию не читает. |

## Как выкатить

Первый запуск и каждое обновление выполняются на сервере одинаково:

```bash
cd tasklist
git pull
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

Проверка:

```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod ps -a
curl localhost:8090/v1/lists
```

Контейнер `migrate` должен быть в статусе `Exited (0)`, остальные работать. Изменения nginx проверяются и применяются так:

```bash
sudo nginx -t && sudo systemctl reload nginx
```

## Резервная копия базы

```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod exec -T postgres \
  pg_dump -U app tasklist | gzip > backup_$(date +%F).sql.gz
```

## Окружение для разработки (на деплой не влияет)

`docker-compose.yml` поднимает только Postgres (порт 5432 открыт) и `migrate` по запросу. Сервис запускается через `go run` с переменными из `.env`.

## Что ещё не сделано

- HTTPS: нужен домен, затем `certbot --nginx`.
- Авторизация: API открыт всем, кто знает адрес (можно закрыть `auth_basic` в nginx).
- Закрыть `/debug/` в nginx.
- Автоматические бэкапы (`cron`) и автодеплой (GitHub Actions).