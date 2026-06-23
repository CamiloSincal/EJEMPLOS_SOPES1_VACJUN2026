package main

import (
	"context"
	"log"
	"net/http"
	"os"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/gin-gonic/gin"
)

type Message struct {
	Data string `json:"data"`
}

var (
	AppPort    = getEnv("APP_PORT", "3500")
	PubsubName = getEnv("PUBSUB_NAME", "mypubsub")
	TopicName  = getEnv("TOPIC_NAME", "mytopic")
)

var daprClient dapr.Client

func init() {
	client, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Error creando cliente Dapr: %v", err)
	}
	daprClient = client
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func publishHandler(c *gin.Context) {
	var message Message

	if err := c.ShouldBind(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("Published message with data: " + message.Data)
	PublishEvent(message.Data)

	c.JSON(http.StatusOK, gin.H{
		"message_data": message.Data,
	})
}

func PublishEvent(data string) error {
	ctx := context.Background()
	if err := daprClient.PublishEvent(ctx, PubsubName, TopicName, []byte(data)); err != nil {
		panic(err)
	}
	log.Printf("Published %s to topic %s\n", data, TopicName)
	return nil
}

func main() {
	defer daprClient.Close()

	router := gin.Default()
	gin.SetMode(gin.DebugMode)
	router.Use(gin.Logger())
	router.POST("/publish", publishHandler)
	router.Run(":" + AppPort)
	log.Println("Running Publisher Service")
}
