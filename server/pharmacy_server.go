// pharmacy_server.go
package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "shubam/proto"
)

type pharmacyServer struct {
	pb.UnimplementedHospitalServiceServer
}

func (s *pharmacyServer) Pharmacy(ctx context.Context, req *pb.PharmacyRequest) (*pb.PharmacyResponse, error) {
	return &pb.PharmacyResponse{Message: "Pharmacy request processed successfully"}, nil
}

func startPharmacyServer() {
	lis, err := net.Listen("tcp", ":5002") // Listening on port 5002
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHospitalServiceServer(s, &pharmacyServer{})

	log.Printf("Pharmacy gRPC server listening on port %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
