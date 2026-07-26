package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"pos_bos/internal/auth"
	"pos_bos/internal/category"
	"pos_bos/internal/product"
	"pos_bos/internal/rbac"
	"pos_bos/internal/transaction"
	"pos_bos/internal/user"
	customMiddleware "pos_bos/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql" // Import mysql driver
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, relying on environment variables")
	}

	// 1. Connect to Database
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("Database connected successfully!")

	// 2. Initialize Dependency Injection (Wiring)
	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(userRepo, authRepo)
	authHandler := auth.NewAuthHandler(authService)

	rbacRepo := rbac.NewRBACRepository(db)
	rbacService := rbac.NewRBACService(rbacRepo)
	rbacHandler := rbac.NewRBACHandler(rbacService)

	categoryRepo := category.NewCategoryRepository(db)
	categoryService := category.NewCategoryService(categoryRepo)
	categoryHandler := category.NewCategoryHandler(categoryService)

	productRepo := product.NewProductRepository(db)
	productService := product.NewProductService(productRepo, categoryRepo)
	productHandler := product.NewProductHandler(productService)

	txRepo := transaction.NewTransactionRepository(db)
	txService := transaction.NewTransactionService(txRepo)
	txHandler := transaction.NewTransactionHandler(txService)

	// 3. Initialize Router
	router := chi.NewRouter()

	// Add CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Diperketat khusus untuk frontend Nuxt
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Add common middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	// 4. Register Routes
	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes
	router.Group(func(r chi.Router) {
		// Middleware ini akan mem-verifikasi token JWT dari header Authorization.
		// Semua route yang didaftarkan di dalam r.Group ini tidak akan bisa diakses jika token tidak valid/kosong.
		r.Use(customMiddleware.JWTMiddleware)

		// Daftarkan route permissions di sini (otomatis dilindungi JWT)
		rbacHandler.RegisterRoutes(r)

		// Daftarkan route users di sini (dilindungi JWT)
		userHandler.RegisterRoutes(r)

		categoryHandler.RegisterRoutes(r)

		productHandler.RegisterRoutes(r)

		txHandler.RegisterRoutes(r)
	})

	// 5. Start Server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
