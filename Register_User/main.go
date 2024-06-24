package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

type PageData struct {
	UserID   string
	UserEmail string
}

type Doctor struct {
    ID            int      `json:"id"`
    Name          string   `json:"name"`
    Specialty     string   `json:"specialty"`
    Experience    string   `json:"experience"`
    PhotoURL      string   `json:"photo_url"`
    AvailableSlots []string `json:"available_slots"`
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
	appointmentConn, err := grpc.Dial("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("did not connect to appointment service: %w", err)
	}
	appointmentClient = pb.NewHospitalServiceClient(appointmentConn)
	log.Println("Successfully connected to the appointment gRPC server")

	pharmacyConn, err := grpc.Dial("localhost:5002", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("Static/service.html")
		if err != nil {
			http.Error(w, "Error loading service page", http.StatusInternalServerError)
			log.Printf("Error loading service page: %v\n", err)
			return
		}
		tmpl.Execute(w, nil)
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
    log.Println("AppointmentHandler called")  // Debug log

    userIDCookie, err := r.Cookie("userID")
    if err != nil {
        http.Error(w, "Missing user ID", http.StatusBadRequest)
        return
    }
    userEmailCookie, err := r.Cookie("userEmail")
    if err != nil {
        http.Error(w, "Missing user email", http.StatusBadRequest)
        return
    }

    log.Println("Cookies:", userIDCookie, userEmailCookie)  // Debug log

    if r.Method == http.MethodGet {
        tmpl, err := template.ParseFiles("Static/appointment.html")
        if err != nil {
            log.Printf("Error parsing appointment template: %v\n", err)
            http.Error(w, "Error loading appointment page", http.StatusInternalServerError)
            return
        }

        data := struct {
            UserID    string
            UserEmail string
        }{
            UserID:    userIDCookie.Value,
            UserEmail: userEmailCookie.Value,
        }

        tmpl.Execute(w, data)
        return
    }

    if r.Method == http.MethodPost {
        err := r.ParseForm()
        if err != nil {
            http.Error(w, "Parse form error", http.StatusInternalServerError)
            return
        }

        doctorName := r.FormValue("doctor")
        date := r.FormValue("date")
        time := r.FormValue("time")

        if date == "" {
            http.Error(w, "Date is required", http.StatusBadRequest)
            return
        }

        if time == "" {
            http.Error(w, "Time is required", http.StatusBadRequest)
            return
        }

        // Add debug log for form values
        log.Println("Form values:", doctorName, date, time)

        req := &pb.AppointmentRequest{
            DoctorName: doctorName,
            UserId:     userIDCookie.Value,
            Email:      userEmailCookie.Value,
            Date:       date,
            Time:       time,
        }

        log.Println("AppointmentClient:", appointmentClient)  // Debug log

        resp, err := appointmentClient.Appointment(context.Background(), req)
        if err != nil {
            log.Printf("Failed to create appointment: %v", err)
            http.Error(w, "Failed to create appointment", http.StatusInternalServerError)
            return
        }

        log.Println(resp.Message)
        http.Redirect(w, r, "/service", http.StatusSeeOther)
    }
}

func bookedSlotsHandler(w http.ResponseWriter, r *http.Request) {
    doctorName := r.URL.Query().Get("doctor")
    if doctorName == "" {
        http.Error(w, "Doctor name is required", http.StatusBadRequest)
        return
    }

    query := "SELECT date, time FROM appointments WHERE doctor_name = $1 AND status = 'BOOKED'"
    rows, err := db.Query(query, doctorName)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to fetch booked slots: %v", err), http.StatusInternalServerError)
        log.Printf("Error fetching booked slots: %v", err)
        return
    }
    defer rows.Close()

    bookedSlots := make(map[string][]string)
    for rows.Next() {
        var slotDate, time string
        err := rows.Scan(&slotDate, &time)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
            log.Printf("Failed to scan row: %v", err)
            return
        }
        bookedSlots[slotDate] = append(bookedSlots[slotDate], time)
    }

    log.Println("Booked slots fetched")

    w.Header().Set("Content-Type", "application/json")
    err = json.NewEncoder(w).Encode(bookedSlots)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error encoding booked slots: %v", err), http.StatusInternalServerError)
        log.Printf("Error encoding booked slots: %v", err)
        return
    }

    log.Println("Booked slots sent")
}


func pharmacyHandler(w http.ResponseWriter, r *http.Request) {
	userIDCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	userEmailCookie, err := r.Cookie("userEmail")
	if err != nil {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		return
	}

	data := PageData{
		UserID:   userIDCookie.Value,
		UserEmail: userEmailCookie.Value,
	}

	tmpl, err := template.ParseFiles("Static/appointment.html")
if err != nil {
    log.Printf("Error parsing pharmacy template: %v\n", err)
    http.Error(w, "Error loading pharmacy page", http.StatusInternalServerError)
    return
}
log.Println("Pharmacy called")

	tmpl.Execute(w, data)
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
	http.HandleFunc("/service", serviceHandler)
	http.HandleFunc("/appointment", appointmentHandler)
	http.HandleFunc("/pharmacy", pharmacyHandler)
	http.HandleFunc("/bookedSlots", bookedSlotsHandler)

	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}