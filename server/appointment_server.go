package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	pb "shubam/proto"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var db *sql.DB

type appointmentServer struct {
	pb.UnimplementedHospitalServiceServer
}

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

func (s *appointmentServer) Appointment(ctx context.Context, req *pb.AppointmentRequest) (*pb.AppointmentResponse, error) {
	err := s.saveAppointment(req)
	if err != nil {
		return nil, err
	}

	return &pb.AppointmentResponse{Message: "Appointment scheduled successfully"}, nil
}

func (s *appointmentServer) saveAppointment(req *pb.AppointmentRequest) error {
	_, err := db.Exec("INSERT INTO appointments (doctor_name, user_id, email, date, time, status) VALUES ($1, $2, $3, $4, $5, $6)",
		req.DoctorName, req.UserId, req.Email, req.Date, req.Time, "BOOKED")
	if err != nil {
		log.Printf("Error inserting appointment into database: %v", err)
		return err
	}

	log.Println("Appointment details saved successfully")
	return nil
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	lis, err := net.Listen("tcp", ":5001") // Listening on port 5001
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHospitalServiceServer(s, &appointmentServer{})

	log.Printf("Appointment gRPC server listening on port %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
