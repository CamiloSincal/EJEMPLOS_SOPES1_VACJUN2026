package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "grpc/proto"

	"google.golang.org/grpc"
)

// -------------------------------------------------------------------
// Implementación del servidor
// -------------------------------------------------------------------

type servidor struct {
	pb.UnimplementedTweetServiceServer
	mu       sync.Mutex
	tweets   map[int32]*pb.Tweet
	contador int32
}

// CrearTweet guarda un nuevo tweet y lo devuelve con su ID asignado
func (s *servidor) CrearTweet(ctx context.Context, req *pb.CrearTweetRequest) (*pb.CrearTweetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.contador++
	tweet := &pb.Tweet{
		Id:      s.contador,
		Autor:   req.Autor,
		Mensaje: req.Mensaje,
	}
	s.tweets[s.contador] = tweet

	log.Printf("[CrearTweet] ID=%d | @%s: %s", tweet.Id, tweet.Autor, tweet.Mensaje)

	return &pb.CrearTweetResponse{
		Tweet:  tweet,
		Status: "creado",
	}, nil
}

// ObtenerTweet busca un tweet por su ID
func (s *servidor) ObtenerTweet(ctx context.Context, req *pb.ObtenerTweetRequest) (*pb.ObtenerTweetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tweet, existe := s.tweets[req.Id]
	if !existe {
		return &pb.ObtenerTweetResponse{
			Status: fmt.Sprintf("tweet con ID %d no encontrado", req.Id),
		}, nil
	}

	log.Printf("[ObtenerTweet] ID=%d encontrado", req.Id)

	return &pb.ObtenerTweetResponse{
		Tweet:  tweet,
		Status: "encontrado",
	}, nil
}

// -------------------------------------------------------------------
// Main
// -------------------------------------------------------------------

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al abrir puerto: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTweetServiceServer(s, &servidor{
		tweets: make(map[int32]*pb.Tweet),
	})

	log.Println("Servidor gRPC escuchando en :50051")
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
