#!/usr/bin/env bash
# Команды для разработки и деплоя. Справка: ./scripts/env.sh help
set -euo pipefail

# Переходим в корень проекта, откуда бы ни запустили скрипт
cd "$(dirname "$0")/.."

PROD_FILE=docker-compose.prod.yml
PROD_ENV=.env.prod
WEB_DIR=/var/www/tasklist   # куда nginx на сервере смотрит за статикой

# ---------- обёртки над docker compose ----------

# окружение для разработки: docker-compose.yml и .env
dc() { docker compose "$@"; }

# боевое окружение: docker-compose.prod.yml и .env.prod
dcp() {
  [[ -f $PROD_ENV ]] || { echo "Нет файла $PROD_ENV" >&2; exit 1; }
  docker compose -f "$PROD_FILE" --env-file "$PROD_ENV" "$@"
}

# psql внутри контейнера базы; имя функции compose передаётся первым аргументом
psql_in() {
  local compose=$1; shift
  "$compose" exec postgres sh -c 'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" "$@"' sh "$@"
}

load_env() {
  [[ -f .env ]] || { echo "Нет файла .env" >&2; exit 1; }
  set -a; source .env; set +a
}

# ---------- боевые операции ----------

backup() {
  mkdir -p backups
  local file="backups/backup_$(date +%F_%H%M).sql.gz"
  dcp exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' \
    | gzip > "$file" || { rm -f "$file"; echo "Бэкап не удался" >&2; exit 1; }
  echo "Готово: $file"
}

# собирает фронтенд в Docker и выкладывает файлы туда, откуда их раздаёт nginx
web_publish() {
  local tmp
  tmp=$(mktemp -d)
  docker build -f Dockerfile.web --target dist --output "type=local,dest=$tmp" .
  rsync -a --delete "$tmp"/ "$WEB_DIR"/
  rm -rf "$tmp"
  echo "Фронтенд выложен в $WEB_DIR"
}

deploy() {
  git pull --ff-only
  dcp up -d --build
  web_publish
}

# ---------- справка ----------

usage() {
  cat <<'EOF'
Использование: ./scripts/env.sh <команда> [аргументы]

Разработка:
  up                поднять Postgres и применить миграции
  down              остановить контейнеры (данные остаются)
  clean             удалить контейнеры И данные базы (спросит подтверждение)
  migrate           применить миграции
  psql [аргументы]  консоль базы, например: psql -c '\dt'
  logs [сервис]     логи
  ps                состояние контейнеров
  api               запустить Go-сервис (переменные берутся из .env)
  web               запустить Vite (интерфейс на http://localhost:5173)
  test              go test ./...

Боевое окружение (prod <команда>):
  prod up           собрать образы и запустить
  prod down         остановить
  prod ps           состояние контейнеров
  prod logs [сервис] логи
  prod psql [арг.]  консоль базы
  prod backup       дамп базы в backups/
  prod web          собрать и выложить фронтенд (на сервере)
  prod deploy       git pull + prod up + prod web (на сервере)
EOF
}

# ---------- разбор команд ----------

prod_cmd() {
  local sub=${1:-help}; shift || true
  case $sub in
    up)     dcp up -d --build ;;
    down)   dcp down ;;
    ps)     dcp ps -a ;;
    logs)   dcp logs -f --tail=100 "$@" ;;
    psql)   psql_in dcp "$@" ;;
    backup) backup ;;
    web)    web_publish ;;
    deploy) deploy ;;
    *)      usage; exit 1 ;;
  esac
}

cmd=${1:-help}; shift || true

case $cmd in
  up)      dc up -d postgres && dc run --rm migrate ;;
  down)    dc down ;;
  clean)
    read -r -p "Удалить контейнеры и ВСЕ данные базы? [y/N] " answer
    [[ $answer == [yY] ]] || { echo "Отменено"; exit 1; }
    dc down -v
    ;;
  migrate) dc run --rm migrate ;;
  psql)    psql_in dc "$@" ;;
  logs)    dc logs -f --tail=100 "$@" ;;
  ps)      dc ps ;;
  api)     load_env; exec go run ./cmd/api-server ;;
  web)     (cd web && exec npm run dev) ;;
  test)    go test ./... ;;
  prod)    prod_cmd "$@" ;;
  help|-h|--help) usage ;;
  *)       echo "Неизвестная команда: $cmd" >&2; usage; exit 1 ;;
esac