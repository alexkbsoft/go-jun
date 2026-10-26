# go-jun

## Команды для постройки и запуска dev окружения

```bash
docker compose up -d postgres        # поднять базу в фоне
docker compose run --rm migrate      # применить миграции
set -a; source .env; set +a          # загрузить переменные из .env в терминал
go run ./cmd/api-server              # запустить сервис
```

## Команды для получения информации
```bash
docker compose ps                    # запущенные контейнеры проекта
docker ps -a                         # вообще все контейнеры, включая остановленные
docker image ls                      # образы на машине
docker volume ls                     # тома (найдите свой *_postgres_data)
docker volume inspect <имя_тома>     # где и как том устроен
# какие compose-проекты есть и где лежат их файлы
docker compose ls

docker ps -a --format '{{.Names}}\t{{.Image}}\t{{.Label "com.docker.compose.project"}}'
```

```bash
docker compose exec postgres psql -U postgres -d api_server_db -c '\dt'
docker compose exec postgres psql -U postgres -d api_server_db -c 'SELECT * FROM lists'
```

## Requests
```bash
curl -X POST  -H "Content-Type: application/json" localhost:8090/v1/lists -d '{"name":"покупки"}'
curl -X POST  -H "Content-Type: application/json" localhost:8090/v1/tasks -d '{"title":"купить молоко", "list_id": "fe309137-84b2-4d15-8962-5a659f025549"}'
curl -X GET  localhost:8090/v1/lists/fe309137-84b2-4d15-8962-5a659f025549
```
## Building prod
```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
docker compose -f docker-compose.prod.yml --env-file .env.prod ps -a

ssh сервер 'cd tasklist && ./scripts/env.sh prod deploy'
```
## Зайти внутрь работающего контейнера
```bash
docker exec -it -u 0 tasklist-prod-api-1 sh
docker compose exec postgres psql -U postgres -d api_server_db
```
-u 0 - открыть суперюзером