# go-repo-deps-checker

CLI для проверки обновлений зависимостей в Go репозиториях. Клонирует репозиторий, парсит `go.mod` и показывает список зависимостей, для которых доступны новые версии.

## Установка

```bash
go install github.com/YOUR_USERNAME/go-repo-deps-checker@latest
```

Или сборка из исходников:

```bash
git clone https://github.com/YOUR_USERNAME/go-repo-deps-checker
cd go-repo-deps-checker
go build -ldflags "-X repoDepsCheckerCLI/cmd.Version=1.0.0" -o go-repo-deps-checker .
```

## Использование

```bash
go-repo-deps-checker <url-репозитория>
```

### Примеры

```bash
# Публичный репозиторий
go-repo-deps-checker https://github.com/user/repo

# С токеном для приватного репо (или GITHUB_TOKEN в env)
go-repo-deps-checker https://github.com/user/repo -t ghp_xxx

# JSON в файл
go-repo-deps-checker https://github.com/user/repo -f json -o report.json

# Без кеша
go-repo-deps-checker https://github.com/user/repo --no-cache

# Без progress bar (для CI)
go-repo-deps-checker https://github.com/user/repo --no-progress
```

## Флаги

| Флаг | Описание |
|------|----------|
| `-t, --token` | GitHub токен для приватных репозиториев |
| `-f, --format` | Формат вывода: table, json, simple (default: table) |
| `-o, --output` | Файл для сохранения результата |
| `--no-cache` | Игнорировать кеш, всегда запрашивать с proxy |
| `--no-progress` | Отключить progress bar |
| `--no-color` | Отключить цветной вывод |
| `--concurrency` | Количество одновременных запросов (default: 10) |
| `--retries` | Количество повторных попыток при ошибке (default: 3) |
| `-v, --version` | Показать версию |

## Кеширование

Ответы proxy кешируются в `~/.cache/repo-deps-checker/` (или `~/Library/Caches/repo-deps-checker/` на macOS) на 1 час. Повторные запуски для тех же зависимостей выполняются значительно быстрее.

## Требования

- Go 1.21+
- Доступ к proxy.golang.org (или GOPROXY)

## Лицензия

MIT
