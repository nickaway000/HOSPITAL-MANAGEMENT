// appointment_server.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "shubam/proto"
)

type appointmentServer struct {
	pb.UnimplementedHospitalServiceServer
}

func (s *appointmentServer) Appointment(ctx context.Context, req *pb.AppointmentRequest) (*pb.AppointmentResponse, error) {
	return &pb.AppointmentResponse{Message: "Appointment scheduled successfully"}, nil
}

func startAppointmentServer() {
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
