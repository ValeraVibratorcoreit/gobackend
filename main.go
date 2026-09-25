package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	port := getEnv("PORT", "8080")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := pingWithRetry(db, 10, time.Second); err != nil {
		log.Fatalf("db unreachable: %v", err)
	}

	store := NewStore(db)
	app := NewApp(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("GET /lessons", app.listLessons)
	mux.HandleFunc("GET /lessons/{id}", app.getLesson)
	mux.HandleFunc("POST /lessons", app.createLesson)
	mux.HandleFunc("PATCH /lessons/{id}", app.updateLessonStatus)
	mux.HandleFunc("DELETE /lessons/{id}", app.deleteLesson)

	mux.HandleFunc("GET /groups", app.listGroups)
	mux.HandleFunc("GET /teachers", app.listTeachers)
	mux.HandleFunc("GET /subjects", app.listSubjects)
	mux.HandleFunc("GET /classrooms", app.listClassrooms)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS разрешает запросы фронтенда с другого origin (пара 3):
// вёрстка может открываться отдельно от API, поэтому preflight (OPTIONS)
// для PATCH/DELETE нужно обрабатывать явно.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func pingWithRetry(db *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		log.Printf("waiting for db... (%d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return err
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
