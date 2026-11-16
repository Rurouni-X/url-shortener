# URL Shortener

Минималистичный REST API сервис для сокращения ссылок на Go с использованием PostgreSQL, роутера `chi` и структурированного логирования `slog`.

## Стек технологий

* **Язык:** Go
* **HTTP-роутер:** [go-chi/chi](https://github.com)
* **База данных:** PostgreSQL (`sqlx` + `lib/pq`)
* **Конфигурация:** [ilyakaznacheev/cleanenv](https://github.com)
* **Валидация:** [go-playground/validator](https://github.com)
* **Логирование:** стандартный пакет `log/slog`
* **Инфраструктура:** Docker, Docker Compose

---

## Структура проекта

```text
url-shortener/
├── cmd/
│   └── url-shortener/
│       └── main.go                 # Точка входа, инициализация роутера и HTTP-сервера
├── config/
│   └── local.yaml                  # Конфигурационный файл приложения
├── sql/
│   └── init.sql                    # SQL-скрипт инициализации таблицы при старте контейнера
├── internal/
│   ├── config/                     # Загрузка и парсинг конфигурации
│   ├── http-server/
│   │   └── handlers/
│   │       ├── redirect/           # Редирект по короткому алиасу
│   │       └── url/
│   │           ├── delete/         # Удаление ссылки
│   │           ├── getAll/         # Получение списка всех ссылок
│   │           └── save/           # Создание новой короткой ссылки
│   ├── lib/
│   │   ├── api/response/           # Общие структуры HTTP-ответов
│   │   ├── logger/sl/              # Хелперы для slog
│   │   └── random/                 # Генератор случайных алиасов
│   └── storage/
│       └── database/               # Реализация работы с PostgreSQL
└── docker-compose.yml              # Конфигурация контейнера PostgreSQL
```

---

## База данных

При запуске базы данных через Docker Compose таблица `url` и индекс создаются автоматически из файла `sql/init.sql`:

```sql
CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
```

Если вы используете локально установленный PostgreSQL без Docker, выполните этот скрипт вручную.

---

## Конфигурация

Приложение считывает настройки из YAML-файла, путь к которому передается через переменную окружения `CONFIG_PATH`. Пароль для Basic Auth задается через переменную `HTTP_SERVER_PASSWORD`.

Пример файла `config/local.yaml` (настроен под параметры из `docker-compose.yml`):

```yaml
env: "local" # Доступные значения: local, dev, prod
storage_path: "postgres://url-shortener:pass@localhost:5433/url?sslmode=disable"
http_server:
  address: "localhost:8080"
  timeout: 4s
  idle_timeout: 60s
  user: "admin"
```

### Переменные окружения

| Переменная | Обязательная | Описание |
| :--- | :---: | :--- |
| `CONFIG_PATH` | Да | Путь до YAML-файла конфигурации |
| `HTTP_SERVER_PASSWORD` | Да | Пароль для Basic Auth (переопределяет значение из конфига) |

---

## Запуск проекта

1. **Клонируйте репозиторий и установите зависимости:**
   ```bash
   git clone https://github.com/Rurouni-X/url-shortener
   cd url-shortener
   go mod download
   ```

2. **Поднимите базу данных PostgreSQL в Docker:**
   ```bash
   docker-compose up -d
   ```
   *База данных запустится на порту `5433`, а скрипт `sql/init.sql` автоматически создаст нужную таблицу.*

3. **Задайте переменные окружения и запустите сервер:**
   ```bash
   export CONFIG_PATH="./config/local.yaml"
   export HTTP_SERVER_PASSWORD="secret_password"
   go run ./cmd/url-shortener/main.go
   ```

---

## API Эндпоинты

### Публичные маршруты

#### 1. Переход по короткой ссылке
* **Метод:** `GET /{alias}`
* **Описание:** Выполняет HTTP-редирект (`302 Found`) на исходный URL.
* **Пример:**
  ```bash
  curl -i http://localhost:8080/google
  ```

#### 2. Получение списка всех ссылок
* **Метод:** `GET /`
* **Описание:** Возвращает JSON-массив всех сохраненных ссылок и их алиасов.
* **Пример ответа (`200 OK`):**
  ```json
  [
    {
      "url": "https://google.com",
      "alias": "google"
    }
  ]
  ```

---

### Защищенные маршруты (Basic Auth)

Для доступа к эндпоинтам `/url` требуется базовая HTTP-аутентификация (`user` и `password` из конфигурации).

#### 3. Создание короткой ссылки
* **Метод:** `POST /url`
* **Тело запроса (JSON):**
  * `url` *(обязательно)* — исходная ссылка (должна быть валидным URL).
  * `alias` *(опционально)* — желаемый короткий алиас. Если не указан, генерируется автоматически (6 символов).
* **Пример запроса:**
  ```bash
  curl -X POST http://localhost:8080/url \
    -u admin:secret_password \
    -H "Content-Type: application/json" \
    -d '{"url": "https://github.com", "alias": "gh"}'
  ```
* **Пример ответа (`201 Created`):**
  ```json
  {
    "status": "OK",
    "alias": "gh"
  }
  ```

#### 4. Удаление ссылки
* **Метод:** `DELETE /url/{alias}`
* **Описание:** Удаляет запись по указанному алиасу.
* **Пример запроса:**
  ```bash
  curl -X DELETE http://localhost:8080/url/gh \
    -u admin:secret_password
  ```
* **Ответы:** `204 No Content` при успешном удалении, `404 Not Found` если алиас не найден.
