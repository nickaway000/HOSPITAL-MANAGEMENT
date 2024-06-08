package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var db *sql.DB

func initDB() error {
	err := godotenv.Load("/Users/nikhil/HOSPITAL-MANAGEMENT/.env")
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

func formHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "Parse form error: %v", err)
		return
	}
	fmt.Fprintf(w, "Form submitted successfully")
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

	fmt.Fprintf(w, "Registration successful")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
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

	fmt.Fprintf(w, "Login successful")
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	fileServer := http.FileServer(http.Dir("/Users/nikhil/HOSPITAL-MANAGEMENT/Static"))
	http.Handle("/", fileServer)

	http.HandleFunc("/register", formHandler)
	http.HandleFunc("/register-submit", registerHandler)
	http.HandleFunc("/login", loginHandler)

	fmt.Printf("Starting server at 8080 port\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}




