package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config хранит все конфигурационные параметры приложения.
// Значения читаются из переменных окружения или .env файла.
type Config struct {
	AppPort       string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	SQLitePath    string
}

// LoadConfig загружает конфигурацию из переменных окружения.
// Сначала пытается загрузить из .env файла, если он существует.
func LoadConfig() (*Config, error) {
	// Загружаем значения из .env файла, если он существует.
	// Это удобно для локальной разработки.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	appPort := getEnv("APP_PORT", "3000")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDBStr := getEnv("REDIS_DB", "0")
	sqlitePath := getEnv("SQLITE_PATH", "./rim.db")

	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB value: %s. Using default 0. Error: %v", redisDBStr, err)
		redisDB = 0 // Используем значение по умолчанию в случае ошибки
	}

	return &Config{
		AppPort:       appPort,
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
		SQLitePath:    sqlitePath,
	}, nil
}

// getEnv читает переменную окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
