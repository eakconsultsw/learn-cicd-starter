package main

import (
	"database/sql"
	"embed"
	"io"
	"log"
        "flag"
	"net/http"
        "time"
	"os"
        "strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/bootdotdev/learn-cicd-starter/internal/database"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var portFlag = flag.String("port", "", "Port on which the HTTP server should listen")

func getPortFromEnvOrFlag() string {
    flag.Parse() // parses -port=... etc.

    // 1️⃣ Flag‑Wert hat Vorrang
    if *portFlag != "" {
        return *portFlag
    }

    // 2️⃣ Dann Umgebungs‑Variable
    if p := os.Getenv("PORT"); p != "" {
        return p
    }

    // 3️⃣ Fallback‑Default
    return "8080"
}
type apiConfig struct {
	DB *database.Queries
}

//go:embed static/*
var staticFiles embed.FS

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

//	port := os.Getenv("PORT")
     // … Port aus Konfiguration/CLI/Env einlesen …
        port := getPortFromEnvOrFlag() // string
    
        if p, err := strconv.Atoi(port); err == nil && p > 0 && p <= 65535 {
        // 2. Nur das geprüfte, numerische Ergebnis loggen
           log.Println("Serving on port:", p)
        } else {
        // 3. Bei ungültigem Wert keinen rohen Input ausgeben
           log.Println("Invalid port configuration – server not started")
           return // oder geeignete Fehlerbehandlung
        }
        
	apiCfg := apiConfig{}

	// https://github.com/libsql/libsql-client-go/#open-a-connection-to-sqld
	// libsql://[your-database].turso.io?authToken=[your-auth-token]
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		apiCfg.DB = dbQueries
		log.Println("Connected to database!")
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := staticFiles.Open("static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	v1Router := chi.NewRouter()

	if apiCfg.DB != nil {
		v1Router.Post("/users", apiCfg.handlerUsersCreate)
		v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))
		v1Router.Get("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesGet))
		v1Router.Post("/notes", apiCfg.middlewareAuth(apiCfg.handlerNotesCreate))
	}

	v1Router.Get("/healthz", handlerReadiness)

	router.Mount("/v1", v1Router)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
                ReadHeaderTimeout: 10 * time.Second,   // max. Zeit, die ein Header‑Teil gelesen werden darf
                ReadTimeout:  30 * time.Second,
                WriteTimeout: 30 * time.Second,
                IdleTimeout:  60 * time.Second,
	}
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
           log.Fatalf("Server error: %v", err)
       }	
}
