# Настройка переменных окружения

## Frontend

Создайте файл `.env` в директории `frontend/` со следующими переменными:

```bash
# Имя Telegram бота (без @)
VITE_BOT_USERNAME=your_bot_username_here

# Базовый URL для API (опционально)
VITE_API_BASE_URL=http://localhost:8080
```

## Backend

Создайте файл `.env` в корневой директории проекта со следующими переменными:

```bash
# Токен Telegram бота
BOT_TOKEN=your_bot_token_here

# Имя Telegram бота (без @)
BOT_USERNAME=your_bot_username_here

# Путь к базе данных
DATABASE_URL=rim.db

# Порт сервера
PORT=8080
```

## Получение данных для Telegram Bot

1. Создайте бота через [@BotFather](https://t.me/BotFather)
2. Отправьте команду `/newbot` и следуйте инструкциям
3. Получите токен бота (BOT_TOKEN)
4. Имя бота (BOT_USERNAME) - это имя, которое вы выбрали при создании (без @)

## Настройка localhost.local

Для использования домена `localhost.local` добавьте в файл `/etc/hosts` (macOS/Linux):

```bash
127.0.0.1 localhost.local
```

Команда для добавления:
```bash
echo "127.0.0.1 localhost.local" | sudo tee -a /etc/hosts
```

После этого приложение будет доступно по адресам:
- http://localhost
- http://localhost.local

## Важно

- Добавьте `.env` в `.gitignore` 
- Никогда не коммитьте файлы с реальными токенами
- Для продакшн окружения используйте соответствующие переменные окружения
- Для запуска на 80 порте может потребоваться sudo: `sudo npm run dev` 