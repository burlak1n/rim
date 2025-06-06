package main

import (
	"context"
	"fmt"
	"log/slog"

	"rim/internal/config"
	"rim/pkg/database"
	"rim/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	authDelivery "rim/internal/auth/delivery"
	authRepo "rim/internal/auth/repository"
	authUseCase "rim/internal/auth/usecase"

	contactDelivery "rim/internal/contact/delivery"
	contactRepo "rim/internal/contact/repository"
	contactUseCase "rim/internal/contact/usecase"

	groupDelivery "rim/internal/group/delivery"
	groupRepo "rim/internal/group/repository"
	groupUseCase "rim/internal/group/usecase"

	systemDelivery "rim/internal/system/delivery"
	systemRepo "rim/internal/system/repository"
	systemUseCase "rim/internal/system/usecase"
)

// initSystemSettings инициализирует системные настройки при первом запуске
func initSystemSettings(sysUseCase systemUseCase.UseCase, log *slog.Logger) {
	ctx := context.Background()

	// Проверяем, существует ли настройка debug_mode
	_, err := sysUseCase.GetDebugMode(ctx)
	if err != nil {
		// Если настройка не найдена, создаем её со значением false
		log.Info("Debug mode setting not found, creating with default value (false)")
		err = sysUseCase.SetDebugMode(ctx, false)
		if err != nil {
			log.Error("Failed to initialize debug_mode setting", slog.Any("error", err))
		} else {
			log.Info("Debug mode setting initialized successfully")
		}
	} else {
		log.Info("Debug mode setting already exists")
	}
}

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

	// Добавляем middleware безопасности в начале
	app.Use(authDelivery.SecurityMiddleware())

	// Настройка CORS с поддержкой cookies
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost, http://localhost:80, http://localhost.local, http://localhost.local:80",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token",
		AllowCredentials: true, // Важно для cookies
	}))

	// Инициализация зависимостей для модуля Group
	grpRepo := groupRepo.NewSQLiteRepository(sqliteDB, log)
	grpUseCase := groupUseCase.NewGroupUseCase(grpRepo, log)
	grpHandler := groupDelivery.NewHandler(grpUseCase, log)

	// Инициализация зависимостей для модуля Contact
	// contactRepo используется в auth, поэтому создается раньше
	cntRepo := contactRepo.NewSQLiteRepository(sqliteDB, log)

	// Инициализация зависимостей для модуля Auth
	authRepository := authRepo.NewAuthRepository(sqliteDB, redisClient, log)
	authUseCaseInstance := authUseCase.NewAuthUseCase(authRepository, cntRepo, log)

	// Инициализация зависимостей для модуля System
	sysRepo := systemRepo.NewSQLiteRepository(sqliteDB, log)
	sysUseCase := systemUseCase.NewSystemUseCase(sysRepo, log)
	sysHandler := systemDelivery.NewHandler(sysUseCase, log)

	// Инициализация системных настроек при первом запуске
	initSystemSettings(sysUseCase, log)

	// Завершение инициализации Auth с systemUseCase
	authHandler := authDelivery.NewHandler(authUseCaseInstance, sysUseCase, cfg.BotToken, cfg.ForceDebugMode, log)

	// Завершение инициализации Contact с authUseCase
	cntUseCase := contactUseCase.NewContactUseCase(cntRepo, grpRepo, log)
	cntHandler := contactDelivery.NewHandler(cntUseCase, authUseCaseInstance, log)

	// Группа маршрутов API v1
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Middleware для проверки админских прав с учетом отладочного режима
	requireAdminOrDebug := func(c *fiber.Ctx) error {
		// Сначала проверяем принудительный отладочный режим из переменной окружения
		if cfg.ForceDebugMode {
			log.InfoContext(c.Context(), "Force debug mode is enabled via environment variable, allowing access to authenticated user")
			return c.Next()
		}

		// Затем проверяем отладочный режим из базы данных
		debugMode, err := sysUseCase.GetDebugMode(c.Context())
		if err != nil {
			log.WarnContext(c.Context(), "Failed to get debug mode status", slog.Any("error", err))
		} else if debugMode {
			log.InfoContext(c.Context(), "Debug mode is enabled, allowing access to authenticated user")
			return c.Next()
		}

		// Получаем user_id из контекста (должен быть установлен RequireAuth middleware)
		userID, ok := c.Locals("user_id").(uint)
		if !ok {
			log.WarnContext(c.Context(), "User ID not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Проверяем права администратора
		isAdmin, err := authUseCaseInstance.IsUserAdmin(c.Context(), userID)
		if err != nil {
			log.ErrorContext(c.Context(), "Failed to check admin status", slog.Uint64("user_id", uint64(userID)), slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		if !isAdmin {
			log.WarnContext(c.Context(), "User is not admin and debug mode is off", slog.Uint64("user_id", uint64(userID)))
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin rights required",
			})
		}

		log.InfoContext(c.Context(), "User has admin rights", slog.Uint64("user_id", uint64(userID)))
		return c.Next()
	}

	// Маршруты для Group
	groupRoutes := v1.Group("/groups")
	groupRoutes.Post("/", grpHandler.CreateGroup)
	groupRoutes.Get("/", grpHandler.GetAllGroups)
	groupRoutes.Get("/:id", grpHandler.GetGroupByID)
	groupRoutes.Put("/:id", grpHandler.UpdateGroup)
	groupRoutes.Delete("/:id", grpHandler.DeleteGroup)

	// Маршруты для Contact
	contactRoutes := v1.Group("/contacts")
	// Применяем secure cookie middleware для проверки авторизации
	contactRoutes.Use(authHandler.CookieAuthMiddleware())
	// Добавляем CSRF защиту для всех изменяющих операций
	contactRoutes.Use(authHandler.CSRFMiddleware())

	contactRoutes.Get("/", cntHandler.GetAllContacts) // Доступно без авторизации (ограниченные данные)

	// Защищенные роуты (требуют авторизации)
	contactRoutes.Post("/", authHandler.RequireAuthCookie(), requireAdminOrDebug, cntHandler.CreateContact)
	contactRoutes.Get("/:id", authHandler.RequireAuthCookie(), cntHandler.GetContactByID)
	contactRoutes.Put("/:id", authHandler.RequireAuthCookie(), requireAdminOrDebug, cntHandler.UpdateContact)
	contactRoutes.Delete("/:id", authHandler.RequireAuthCookie(), requireAdminOrDebug, cntHandler.DeleteContact)
	// Маршруты для управления связями контактов и групп (только админ)
	contactRoutes.Post("/:contact_id/groups/:group_id", authHandler.RequireAuthCookie(), requireAdminOrDebug, cntHandler.AddContactToGroup)        // Добавить контакт в группу
	contactRoutes.Delete("/:contact_id/groups/:group_id", authHandler.RequireAuthCookie(), requireAdminOrDebug, cntHandler.RemoveContactFromGroup) // Удалить контакт из группы

	// Маршруты для Auth
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/telegram", authHandler.AuthWithTelegram)
	authRoutes.Get("/me", authHandler.GetMe)
	authRoutes.Get("/csrf-token", authHandler.GetCSRFToken) // Получить CSRF токен

	// Защищенные auth роуты с CSRF защитой
	authRoutes.Use(authHandler.CSRFMiddleware())
	authRoutes.Put("/contact", authHandler.RequireAuthCookie(), authHandler.UpdateMyContact) // Обновить свой контакт
	authRoutes.Post("/logout", authHandler.Logout)

	// Маршруты для System (публичные для получения, только админ для установки)
	systemRoutes := v1.Group("/system")
	systemRoutes.Get("/debug-mode", sysHandler.GetDebugMode) // Получить состояние отладочного режима

	// Защищенные system роуты с CSRF защитой
	systemRoutes.Use(authHandler.CSRFMiddleware())
	systemRoutes.Put("/debug-mode", authHandler.RequireAuthCookie(), requireAdminOrDebug, sysHandler.SetDebugMode) // Установить отладочный режим (только админ)

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
