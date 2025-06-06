# Настройка авторизации через Telegram

## Настройка бота

1. Создайте бота через [@BotFather](https://t.me/botfather)
2. Получите токен бота
3. Настройте домен для Telegram Login Widget:
   ```
   /setdomain
   @your_bot_name
   yourdomain.com
   ```

## Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```env
APP_PORT=3000
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
SQLITE_PATH=./rim.db
BOT_TOKEN=your_telegram_bot_token_here
```

## Настройка frontend

В файле `frontend/src/lib/components/TelegramAuth.svelte` замените:
```typescript
export let botUsername = 'your_bot_username'; // Замените на имя вашего бота
```

В файле `frontend/src/lib/pages/Login.svelte` замените:
```svelte
<TelegramAuth botUsername="your_bot_username" />
```

## Запуск приложения

### Backend
```bash
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm run dev
```

## Использование

1. Неавторизованные пользователи видят только имена контактов
2. Для полного доступа нужно авторизоваться через Telegram
3. После авторизации пользователь получает доступ ко всем функциям
4. Если у пользователя есть контакт с соответствующим Telegram ID, он автоматически связывается

## API Endpoints

### Авторизация
- `POST /api/v1/auth/telegram` - авторизация через Telegram
- `GET /api/v1/auth/me` - получить информацию о пользователе
- `POST /api/v1/auth/logout` - выход из системы

### Контакты
- `GET /api/v1/contacts` - получить список контактов (ограниченный для неавторизованных)
- `POST /api/v1/contacts` - создать контакт (требует авторизации)
- `GET /api/v1/contacts/:id` - получить контакт по ID (требует авторизации)
- `PUT /api/v1/contacts/:id` - обновить контакт (требует авторизации)
- `DELETE /api/v1/contacts/:id` - удалить контакт (требует авторизации)

## Структура базы данных

### Таблица users
- `id` - уникальный идентификатор
- `telegram_id` - ID пользователя в Telegram
- `contact_id` - связь с таблицей contacts (nullable)
- `is_active` - активен ли пользователь
- `created_at`, `updated_at` - временные метки

### Таблица contacts (обновлена)
- Добавлено поле `telegram_id` для связи с пользователями Telegram

### Redis сессии
- Ключ: `session:{session_token}`
- Значение: JSON с данными сессии (user_id, created_at, expired_at)
- TTL: 30 дней 