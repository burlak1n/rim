package main

import (
	"fmt"
	"log/slog"

	"rim/internal/config"
	"rim/pkg/database"
	"rim/pkg/logger"

	"github.com/gofiber/fiber/v2"

	authDelivery "rim/internal/auth/delivery"
	authRepo "rim/internal/auth/repository"
	authUseCase "rim/internal/auth/usecase"

	contactDelivery "rim/internal/contact/delivery"
	contactRepo "rim/internal/contact/repository"
	contactUseCase "rim/internal/contact/usecase"

	groupDelivery "rim/internal/group/delivery"
	groupRepo "rim/internal/group/repository"
	groupUseCase "rim/internal/group/usecase"
)

// @title RIM API
// @version 1.0
// @description Корпоративный портал RIM для управления контактами, группами и ресурсами.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /api/v1
func main() {
	log := logger.NewLogger()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("Failed to load config", slog.Any("error", err))
		return
	}

	log.Info("Config loaded successfully")

	// Подключаемся к SQLite
	sqliteDB, err := database.NewSQLiteConnection(cfg, log)
	if err != nil {
		// Ошибка уже залогирована в NewSQLiteConnection
		return
	}
	// Пока не используем sqliteDB, но он готов
	_ = sqliteDB // Это чтобы компилятор не ругался на неиспользуемую переменную

	// Подключаемся к Redis
	redisClient, err := database.NewRedisClient(cfg, log)
	if err != nil {
		// Ошибка уже залогирована в NewRedisClient
		return
	}
	// Пока не используем redisClient, но он готов
	_ = redisClient // Это чтобы компилятор не ругался на неиспользуемую переменную

	app := fiber.New()

	// Инициализация зависимостей для модуля Group
	grpRepo := groupRepo.NewSQLiteRepository(sqliteDB, log)
	grpUseCase := groupUseCase.NewGroupUseCase(grpRepo, log)
	grpHandler := groupDelivery.NewHandler(grpUseCase, log)

	// Инициализация зависимостей для модуля Contact
	// contactRepo использует grpRepo для проверки существования групп
	cntRepo := contactRepo.NewSQLiteRepository(sqliteDB, log)
	cntUseCase := contactUseCase.NewContactUseCase(cntRepo, grpRepo, log)
	cntHandler := contactDelivery.NewHandler(cntUseCase, log)

	// Инициализация зависимостей для модуля Auth
	authRepository := authRepo.NewAuthRepository(sqliteDB, redisClient, log)
	authUseCaseInstance := authUseCase.NewAuthUseCase(authRepository, cntRepo, log)
	authHandler := authDelivery.NewHandler(authUseCaseInstance, cfg.BotToken, log)

	// Группа маршрутов API v1
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Маршруты для Group
	groupRoutes := v1.Group("/groups")
	groupRoutes.Post("/", grpHandler.CreateGroup)
	groupRoutes.Get("/", grpHandler.GetAllGroups)
	groupRoutes.Get("/:id", grpHandler.GetGroupByID)
	groupRoutes.Put("/:id", grpHandler.UpdateGroup)
	groupRoutes.Delete("/:id", grpHandler.DeleteGroup)

	// Маршруты для Contact
	contactRoutes := v1.Group("/contacts")
	// Применяем middleware для проверки авторизации (но не требуем её)
	contactRoutes.Use(authHandler.AuthMiddleware())

	contactRoutes.Get("/", cntHandler.GetAllContacts) // Доступно без авторизации (ограниченные данные)

	// Защищенные роуты (требуют авторизации)
	contactRoutes.Post("/", authHandler.RequireAuth(), cntHandler.CreateContact)
	contactRoutes.Get("/:id", authHandler.RequireAuth(), cntHandler.GetContactByID)
	contactRoutes.Put("/:id", authHandler.RequireAuth(), cntHandler.UpdateContact)
	contactRoutes.Delete("/:id", authHandler.RequireAuth(), cntHandler.DeleteContact)
	// Маршруты для управления связями контактов и групп
	contactRoutes.Post("/:contact_id/groups/:group_id", authHandler.RequireAuth(), cntHandler.AddContactToGroup)        // Добавить контакт в группу
	contactRoutes.Delete("/:contact_id/groups/:group_id", authHandler.RequireAuth(), cntHandler.RemoveContactFromGroup) // Удалить контакт из группы

	// Маршруты для Auth
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/telegram", authHandler.AuthWithTelegram)
	authRoutes.Get("/me", authHandler.GetMe)
	authRoutes.Post("/logout", authHandler.Logout)

	app.Get("/", func(c *fiber.Ctx) error {
		log.Info("Received request for /", slog.String("ip", c.IP()))
		return c.SendString("Hello, World! Welcome to RIM API.")
	})

	listenAddr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Info("Starting server", slog.String("address", listenAddr))

	if err := app.Listen(listenAddr); err != nil {
		log.Error("Failed to start server", slog.Any("error", err))
	}
}
