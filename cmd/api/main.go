package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	rbacRepo := rbac.NewRBACRepository(db)
	rbacService := rbac.NewRBACService(rbacRepo)
	rbacHandler := rbac.NewRBACHandler(rbacService)

	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(userRepo, authRepo, rbacRepo)
	authHandler := auth.NewAuthHandler(authService)

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

	// Parse allowed origins dari .env
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "http://localhost:3000"
	}
	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	// Add CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins, // Diperketat, baca dari env atau default localhost:3000
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

	// Serve static files for uploaded photos with hotlink protection
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	
	router.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				referer := r.Header.Get("Referer")
				
				isAllowed := false
				for _, origin := range allowedOrigins {
					if strings.HasPrefix(referer, origin) {
						isAllowed = true
						break
					}
				}

				if !isAllowed {
					http.Error(w, "Forbidden: Akses gambar secara langsung tidak diizinkan", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(filesDir)))
	})

	// 4. Register Routes
	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes
	router.Group(func(r chi.Router) {
		// Middleware ini akan mem-verifikasi token JWT dari header Authorization.
		// Semua route yang didaftarkan di dalam r.Group ini tidak akan bisa diakses jika token tidak valid/kosong.
		r.Use(customMiddleware.JWTMiddleware)

		authHandler.RegisterProtectedRoutes(r)
		rbacHandler.RegisterRoutes(r, rbacRepo)
		userHandler.RegisterRoutes(r, rbacRepo)
		categoryHandler.RegisterRoutes(r, rbacRepo)
		productHandler.RegisterRoutes(r, rbacRepo)
		txHandler.RegisterRoutes(r, rbacRepo)
	})

	// 5. Start Server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
