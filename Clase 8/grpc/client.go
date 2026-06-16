package main

import (
	"context"
	"log"
	"time"

	pb "grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Conectarse al servidor
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error al conectar con el servidor: %v", err)
	}
	defer conn.Close()

	cliente := pb.NewTweetServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// -------------------------------------------------------------------
	// 1. Crear varios tweets
	// -------------------------------------------------------------------
	log.Println("=== Creando tweets ===")

	tweets := []struct{ autor, mensaje string }{
		{"gopher", "Go + gRPC es una combinación poderosa"},
		{"camilo", "Aprendiendo gRPC desde cero"},
		{"linux", "El kernel es la base de todo"},
	}

	for _, t := range tweets {
		resp, err := cliente.CrearTweet(ctx, &pb.CrearTweetRequest{
			Autor:   t.autor,
			Mensaje: t.mensaje,
		})
		if err != nil {
			log.Fatalf("Error al crear tweet: %v", err)
		}
		log.Printf("  ✓ Tweet creado — ID=%d | @%s: %s [status: %s]",
			resp.Tweet.Id, resp.Tweet.Autor, resp.Tweet.Mensaje, resp.Status)
	}

	// -------------------------------------------------------------------
	// 2. Obtener tweets por ID
	// -------------------------------------------------------------------
	log.Println("\n=== Obteniendo tweets ===")

	for _, id := range []int32{1, 2, 3, 99} {
		resp, err := cliente.ObtenerTweet(ctx, &pb.ObtenerTweetRequest{Id: id})
		if err != nil {
			log.Fatalf("Error al obtener tweet: %v", err)
		}
		if resp.Tweet != nil {
			log.Printf("  ✓ ID=%d | @%s: %s", resp.Tweet.Id, resp.Tweet.Autor, resp.Tweet.Mensaje)
		} else {
			log.Printf("  ✗ ID=%d — %s", id, resp.Status)
		}
	}
}
