package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/cwshea/localization/internal/config"
	"github.com/cwshea/localization/internal/database"
	"github.com/cwshea/localization/internal/handlers"
	"github.com/cwshea/localization/internal/llm"
	"github.com/cwshea/localization/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	llmFactory := llm.NewClientFactory(cfg)
	svc := service.NewTranslationService(pool, llmFactory)
	h := handlers.NewHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", h.Health)

	r.Route("/api/source-strings", func(r chi.Router) {
		r.Get("/", h.ListSourceStrings)
		r.Post("/", h.CreateSourceString)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetSourceString)
			r.Put("/", h.UpdateSourceString)
			r.Delete("/", h.DeleteSourceString)
			r.Post("/retranslate", h.Retranslate)
		})
	})

	r.Route("/api/translations", func(r chi.Router) {
		r.Put("/{id}", h.UpdateTranslation)
		r.Delete("/{id}", h.DeleteTranslation)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
