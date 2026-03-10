# Лабораторная работа 3 — запуск и тестирование API

## Что изменилось по сравнению с последним коммитом (Lab2)

Раньше приложение отдавало только **HTML-страницы** (главная, карточка конструкции, страница заявки) и обрабатывало формы (добавить в заявку, удалить заявку). Теперь добавлен **REST API** для тех же сущностей: все операции можно вызывать через HTTP-запросы и получать ответы в JSON. Это нужно для будущего SPA-фронтенда.

### Добавленный функционал (кратко)

| Категория | Что появилось |
|-----------|----------------|
| **Услуги (constructions)** | GET список с фильтром по названию, GET одна запись, POST создание с загрузкой изображения и видео в MinIO |
| **Связь заявка–услуга (м-м)** | POST «добавить услугу в черновик», DELETE убрать из заявки, PUT изменить количество/дату рубки/поправку |
| **Заявки (dendrochronologies)** | GET иконка корзины (id черновика + кол-во услуг), GET список с фильтром по датам и статусу, GET одна заявка с услугами, PUT изменить описание, PUT сформировать (с расчётом даты постройки), PUT завершить/отклонить модератором, DELETE удалить черновик |
| **Пользователи** | POST регистрация, POST вход (заглушка), POST выход (заглушка) |
| **Инфраструктура** | Подключение MinIO (загрузка файлов), отдельный пакет сериализаторов (JSON), фиксированный «текущий пользователь» (userID=1) для всех API-запросов |

Старые маршруты Lab2 (**/** , **/construction/:id** , **/dendrochronology** и т.д.) сохранены — в браузере по-прежнему можно открывать страницы и пользоваться формами.

---

## Что нужно для запуска

1. **Go** (уже использовался в Lab2).
2. **Docker и Docker Compose** — для PostgreSQL, Adminer, MinIO, Nginx.
3. **Файл `.env`** в корне проекта с переменными для БД и MinIO.

### Проверка .env

В корне проекта должен быть файл `.env` примерно такого вида:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=lexa
DB_PASS=qwerty123
DB_NAME=dendro_db

MINIO_HOST=localhost
MINIO_PORT=9090
MINIO_USER=lexa
MINIO_PASS=qwerty123
```

Порт MinIO **9090** — это Nginx перед MinIO (в docker-compose проброс 9090:9000). Если подключаетесь к MinIO без Nginx, укажите порт 9000.

---

## Как запустить

### 1. Поднять контейнеры

В корне проекта:

```bash
docker-compose up -d postgres adminer minio1 nginx
```

Проверка: Adminer — http://localhost:8888 , MinIO Console — http://localhost:9091 (логин/пароль из .env).

### 2. Миграции (если ещё не выполняли после добавления Lab3)

Если в БД нет таблицы `dendrochronologies` с полем `build_year`:

```bash
go run ./cmd/migrate/
```

При ошибке подключения к БД проверьте, что контейнер postgres запущен и в `.env` верные данные.

### 3. Запуск приложения

```bash
go run ./cmd/dendroAnalysis/
```

Сервер слушает **http://localhost:8080** (порт из `config/config.toml`).

### 4. Проверка, что Lab2 и Lab3 работают

- В браузере: http://localhost:8080/ — главная с карточками услуг.
- В браузере: http://localhost:8080/dendrochronology — страница заявки (черновик).
- В Postman: **GET** `http://localhost:8080/api/constructions` — в ответе JSON-массив услуг.

Если всё так и есть — можно переходить к тестам API в Postman.

---

## Тестирование в Postman

Базовый URL для всех запросов: **http://localhost:8080**

Во всех методах «текущий пользователь» зафиксирован (id=1). Модератор для завершения/отклонения — пользователь с `is_moderator = true` (например id=2); для проверки finish в Postman можно временно подменить userID в коде или использовать уже существующего модератора в БД.

---

### Домен: Услуги (Constructions)

#### 1. GET список услуг (с опциональным фильтром)

- **Метод:** GET  
- **URL:** `http://localhost:8080/api/constructions`  
- **С фильтром:** `http://localhost:8080/api/constructions?query=сруб`  
- **Тело:** не нужно  
- **Ожидание:** 200, JSON-массив объектов с полями `id`, `construction_title`, `use_life`, `description`, `image_url`, `video_url`, `is_delete`.

#### 2. GET одна услуга

- **Метод:** GET  
- **URL:** `http://localhost:8080/api/constructions/1` (подставьте реальный id из БД)  
- **Ожидание:** 200 и один объект услуги в JSON. При несуществующем id — 404.

#### 3. POST создание услуги (с опциональными файлами)

- **Метод:** POST  
- **URL:** `http://localhost:8080/api/constructions`  
- **Body:** type **form-data**  
  - `construction_title` (text) — обязательно  
  - `use_life` (text) — обязательно  
  - `description` (text) — обязательно  
  - `image` (file) — по желанию, файл изображения  
  - `video` (file) — по желанию, файл видео  
- **Ожидание:** 201, в ответе созданная услуга в JSON, в заголовке `Location` — ссылка на `/api/constructions/{id}`. Файлы сохраняются в MinIO, в БД — имена файлов (латиница).

---

### Домен: Связь заявка–услуга (м-м)

#### 4. POST добавить услугу в черновик заявки

- **Метод:** POST  
- **URL:** `http://localhost:8080/api/constructions/3/add-to-dendrochronology` (3 — id конструкции)  
- **Тело:** не нужно  
- **Ожидание:** 200 или 201 (201, если черновик создали впервые). В ответе — JSON заявки. Если услуга уже в заявке — 409.

Сначала можно вызвать **GET /api/dendrochronologies/cart**, чтобы узнать id черновика (или он появится после первого добавления).

#### 5. DELETE убрать услугу из заявки

- **Метод:** DELETE  
- **URL:** `http://localhost:8080/api/dendrochronology-constructions/3/1`  
  - первый число — `construction_id`, второе — `dendrochronology_id` (черновика).  
- **Ожидание:** 200 и JSON заявки после удаления строки из м-м.

#### 6. PUT изменить в связи (количество, дата рубки, поправка)

- **Метод:** PUT  
- **URL:** `http://localhost:8080/api/dendrochronology-constructions/3/1` (construction_id / dendrochronology_id)  
- **Body:** raw **JSON**  
```json
{
  "samples_count": 2,
  "cutting_date": "2014",
  "date_correction": "5"
}
```
- **Ожидание:** 200 и JSON с обновлёнными полями связи (samples_count, cutting_date, date_correction).

---

### Домен: Заявки (Dendrochronologies)

#### 7. GET иконка корзины

- **Метод:** GET  
- **URL:** `http://localhost:8080/api/dendrochronologies/cart`  
- **Тело:** не нужно  
- **Ожидание:** 200, JSON вида `{"dendrochronology_id": 1, "constructions_count": 2}` или при отсутствии черновика — `{"status": "no_draft", "constructions_count": 0}`.

#### 8. GET список заявок (кроме черновиков и удалённых)

- **Метод:** GET  
- **URL:** `http://localhost:8080/api/dendrochronologies`  
- **Опционально:**  
  - `from_date=2024-01-01`  
  - `to_date=2024-12-31`  
  - `status=сформирован` (или `завершён`, `отклонён`)  
- Пример: `http://localhost:8080/api/dendrochronologies?status=сформирован`  
- **Ожидание:** 200, массив заявок с полями создатель/модератор (логины), `dated_constructions_count` и т.д.

#### 9. GET одна заявка (с её услугами)

- **Метод:** GET  
- **URL:** `http://localhost:8080/api/dendrochronologies/1`  
- **Ожидание:** 200, JSON заявки и вложенный массив `constructions` (услуги с полями из связи: samples_count, cutting_date, date_correction, image_url и т.д.). Удалённые не отдаются.

#### 10. PUT изменить поля заявки (по теме)

- **Метод:** PUT  
- **URL:** `http://localhost:8080/api/dendrochronologies/1`  
- **Body:** raw **JSON**  
```json
{
  "description": "Новое описание заявки"
}
```
- **Ожидание:** 200 и обновлённая заявка. Работает только для черновика.

#### 11. PUT сформировать заявку (создатель)

- **Метод:** PUT  
- **URL:** `http://localhost:8080/api/dendrochronologies/1/form`  
- **Тело:** не нужно  
- **Ожидание:** 200 и заявка со статусом «сформирован», проставлены `date_formed`, `total_samples`, `build_year` (max(cutting_date + date_correction) по строкам м-м). Если черновик пустой — 400.

#### 12. PUT завершить/отклонить заявку (модератор)

- **Метод:** PUT  
- **URL:** `http://localhost:8080/api/dendrochronologies/1/finish`  
- **Body:** raw **JSON**  
```json
{
  "status": "завершён"
}
```
или `"status": "отклонён"`.  
- **Ожидание:** 200 и заявка с `date_completed` и модератором. Вызывать может только пользователь с `is_moderator = true` (в Lab3 userID берётся из синглтона; для теста модератора нужно, чтобы в репозитории был выставлен userID модератора). Заявка должна быть в статусе **«сформирован»** — завершить/отклонить черновик нельзя; сначала выполните **PUT** `.../form` для этой заявки.

**Если получаете 403 «только модератор может завершить или отклонить заявку»:** по умолчанию «текущий пользователь» в API — это `userID = 1`. Нужно, чтобы в таблице `users` у пользователя с `id = 1` было `is_moderator = true`. В Adminer (http://localhost:8888) или через psql выполните:
```sql
UPDATE users SET is_moderator = true WHERE id = 1;
```
После этого PUT `/api/dendrochronologies/:id/finish` с телом `{"status": "завершён"}` или `{"status": "отклонён"}` вернёт 200.

#### 13. DELETE удалить заявку (черновик)

- **Метод:** DELETE  
- **URL:** `http://localhost:8080/api/dendrochronologies/1`  
- **Ожидание:** 200 и `{"status": "deleted"}`. Только для черновика и только создатель.

---

### Домен: Пользователи

#### 14. POST регистрация

- **Метод:** POST  
- **URL:** `http://localhost:8080/api/users/signup`  
- **Body:** raw **JSON**  
```json
{
  "login": "newuser",
  "password": "secret123",
  "is_moderator": false
}
```
- **Ожидание:** 201 и созданный пользователь (без пароля в ответе). При существующем логине — 409.

#### 15. POST вход (заглушка для Lab4)

- **Метод:** POST  
- **URL:** `http://localhost:8080/api/users/signin`  
- **Body:** raw **JSON**  
```json
{
  "login": "user1",
  "password": "qwerty123"
}
```
- **Ожидание:** 200 и данные пользователя. В Lab3 «текущий пользователь» в репозитории можно переключать здесь (заглушка).

#### 16. POST выход (заглушка)

- **Метод:** POST  
- **URL:** `http://localhost:8080/api/users/signout`  
- **Тело:** не нужно  
- **Ожидание:** 200 и `{"status": "signed_out"}`.

---

## Порядок проверки в Postman (сценарий)

1. **GET** `/api/constructions` — убедиться, что список услуг есть.  
2. **GET** `/api/dendrochronologies/cart` — получить id черновика (или 0).  
3. **POST** `/api/constructions/2/add-to-dendrochronology` (подставить реальный id услуги) — добавить в черновик.  
4. **GET** `/api/dendrochronologies/cart` — снова: должен появиться/обновиться `dendrochronology_id` и `constructions_count`.  
5. **GET** `/api/dendrochronologies/{id}` (id черновика) — заявка с массивом `constructions`.  
6. **PUT** `/api/dendrochronology-constructions/2/{dendro_id}` с JSON `{"samples_count":1,"cutting_date":"2012","date_correction":"10"}` — обновить связь.  
7. **PUT** `/api/dendrochronologies/{id}/form` — сформировать заявку; в ответе проверить `build_year`, `total_samples`, `date_formed`.  
8. **GET** `/api/dendrochronologies` — в списке должна быть заявка со статусом «сформирован».  
9. При наличии модератора: **PUT** `/api/dendrochronologies/{id}/finish` с `{"status":"завершён"}`.  
10. **POST** `/api/users/signup` — новый пользователь; **POST** `/api/users/signin` и **POST** `/api/users/signout` — проверить заглушки.

---

## Формат ошибок API

При любой ошибке ответ в JSON вида:

```json
{
  "status": "error",
  "description": "текст ошибки"
}
```

Коды: 400 (неверный запрос), 404 (не найдено), 409 (конфликт, например дубликат), 403 (нет прав), 500 (ошибка сервера).

---

## Сравнение с Lab2 (что где смотреть)

| Действие | Lab2 (браузер) | Lab3 (Postman) |
|----------|----------------|----------------|
| Список услуг | Открыть http://localhost:8080/ | GET /api/constructions |
| Одна услуга | /construction/1 | GET /api/constructions/1 |
| Добавить в заявку | Кнопка на главной, форма POST | POST /api/constructions/:id/add-to-dendrochronology |
| Страница заявки | /dendrochronology | GET /api/dendrochronologies/cart затем GET /api/dendrochronologies/:id |
| Удалить заявку | Кнопка на странице заявки | DELETE /api/dendrochronologies/:id |
| Создать услугу | не было | POST /api/constructions (form-data + файлы) |
| Сформировать / завершить | не было в явном виде | PUT .../form и PUT .../finish |

Файлы для загрузки изображений/видео в Lab3 хранятся в MinIO (бакет `constructions`), отдача по URL через ваш Nginx (например http://localhost:9090/...) в зависимости от конфигурации.

Если что-то из запросов не срабатывает — проверьте, что сервер запущен (`go run ./cmd/dendroAnalysis/`), в Postman выбран правильный метод и URL, а для PUT/POST с телом указан тип Body (JSON или form-data) и корректные значения.
