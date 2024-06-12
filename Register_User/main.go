package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	pb "shubam/proto"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegisterPageData struct {
	Error string
}

var db *sql.DB
var appointmentClient pb.HospitalServiceClient
var pharmacyClient pb.HospitalServiceClient

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

func initGRPC() error {
	appointmentConn, err := grpc.NewClient("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("did not connect to appointment service: %w", err)
	}
	appointmentClient = pb.NewHospitalServiceClient(appointmentConn)
	log.Println("Successfully connected to the appointment gRPC server")

	// Initialize the pharmacy gRPC client
	pharmacyConn, err := grpc.NewClient("localhost:5002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("did not connect to pharmacy service: %w", err)
	}
	pharmacyClient = pb.NewHospitalServiceClient(pharmacyConn)
	log.Println("Successfully connected to the pharmacy gRPC server")

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

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if exists {
		errorMessage := "Email already exists"
		http.Redirect(w, r, "/register.html?error="+errorMessage, http.StatusSeeOther)
		return
	}

	var id int
	err = db.QueryRow("INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", email, password).Scan(&id)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Store the ID and email in session or cookie for later use
	http.SetCookie(w, &http.Cookie{
		Name:  "userID",
		Value: fmt.Sprintf("%d", id),
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "userEmail",
		Value: email,
		Path:  "/",
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}


func servicehandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
		tmpl2, err := template.ParseFiles("Static/service.html")
		if err != nil {
			http.Error(w, "Error loading service page", http.StatusInternalServerError)
			log.Printf("Error loading service page: %v\n", err)
			return
		}
		tmpl2.Execute(w, nil)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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
	var userID int
	err := db.QueryRow("SELECT id, password FROM users WHERE email = $1", email).Scan(&userID, &dbPassword)
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

	// Set cookies with the user ID and email
	http.SetCookie(w, &http.Cookie{
		Name:  "userID",
		Value: fmt.Sprintf("%d", userID),
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "userEmail",
		Value: email,
		Path:  "/",
	})

	http.Redirect(w, r, "/service", http.StatusSeeOther)
}

func appointmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	emailCookie, err := r.Cookie("userEmail")
	if err != nil {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		return
	}

	id := idCookie.Value
	email := emailCookie.Value

	resp, err := appointmentClient.Appointment(context.Background(), &pb.AppointmentRequest{
		Id:    id,
		Email: email,
	})
	if err != nil {
		http.Error(w, "Error calling appointment gRPC service", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Appointment response: %s", resp.Message)
	log.Println("Redirecting to appointment.html")
	http.Redirect(w, r, "/appointment.html", http.StatusSeeOther)
}

func pharmacyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	emailCookie, err := r.Cookie("userEmail")
	if err != nil {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		return
	}

	id := idCookie.Value
	email := emailCookie.Value

	resp, err := pharmacyClient.Pharmacy(context.Background(), &pb.PharmacyRequest{
		Id:    id,
		Email: email,
	})
	if err != nil {
		http.Error(w, "Error calling pharmacy gRPC service", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Pharmacy response: %s", resp.Message)
	log.Println("Redirecting to pharmacy.html")
	http.Redirect(w, r, "/pharmacy.html", http.StatusSeeOther)
}


func main() {

	err := initDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	err = initGRPC()
	if err != nil {
		log.Fatalf("Error initializing gRPC client: %v", err)
	}

	fileServer := http.FileServer(http.Dir("Static"))
	http.Handle("/", fileServer)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/service", servicehandler)
	http.HandleFunc("/appointment", appointmentHandler)
	http.HandleFunc("/pharmacy", pharmacyHandler)

	fmt.Printf("Starting server at 8080 port\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
