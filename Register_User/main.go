package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	errPing := db.Ping()
	if errPing != nil {
		return fmt.Errorf("error pinging the database: %w", errPing)
	}

	log.Println("Successfully connected to the database")
	return nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Parse form error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, password)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func servicehandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		tmpl2, err := template.ParseFiles("Static/service.html")
		if err != nil {
			http.Error(w, "Error loading login page", http.StatusInternalServerError)
			log.Printf("Error loading login page: %v\n", err)
			return
		}
		tmpl2.Execute(w, nil)

		return

	}

}
func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		tmpl, err := template.ParseFiles("Static/login.html")
		if err != nil {
			http.Error(w, "Error loading login page", http.StatusInternalServerError)
			log.Printf("Error loading login page: %v\n", err)
			return
		}
		tmpl.Execute(w, nil)

		return

	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var dbPassword string
	err := db.QueryRow("SELECT password FROM users WHERE email = $1", email).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with email: %s\n", email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			log.Printf("Error querying database: %v\n", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	if password != dbPassword {
		log.Printf("Password mismatch for user: %s\n", email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/service", http.StatusSeeOther)

}

func main() {

	err := initDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	fileServer := http.FileServer(http.Dir("Static"))
	http.Handle("/", fileServer)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/service", servicehandler)

	fmt.Printf("Starting server at 8080 port\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
