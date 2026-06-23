package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Librería oficial de RabbitMQ para Go.
	// El alias "amqp" nos permite escribir amqp.Dial(...) en lugar del nombre completo del paquete.
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// --- Configuración desde variables de entorno ---
	// Se leen variables de entorno para no hardcodear valores sensibles en el código.
	// Esto es fundamental en contenedores/Kubernetes: puedes cambiar el comportamiento
	// sin recompilar, solo ajustando las env vars en el manifiesto del pod.
	rabbitURL := envOrDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq-cluster.rabbitmq-system.svc.cluster.local:5672/")
	queueName := envOrDefault("RABBITMQ_QUEUE", "hello")
	message := envOrDefault("RABBITMQ_MESSAGE", "Hola desde Kubernetes RabbitMQ publisher")
	intervalSeconds := envOrDefault("PUBLISH_INTERVAL_SECONDS", "15")

	// Convertir el string "15" a una duración real de Go (15s).
	// time.ParseDuration entiende sufijos como "s", "m", "h", por eso concatenamos la "s".
	// Si el valor es inválido o negativo, se usa 15 segundos como fallback seguro.
	publishInterval, err := time.ParseDuration(intervalSeconds + "s")
	if err != nil || publishInterval <= 0 {
		publishInterval = 15 * time.Second
	}

	// --- Conexión TCP a RabbitMQ ---
	// amqp.Dial establece la conexión física (TCP) con el broker.
	// La URL tiene el formato: amqp://usuario:contraseña@host:puerto/
	// El protocolo AMQP (Advanced Message Queuing Protocol) opera por defecto en el puerto 5672.
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	// defer garantiza que la conexión se cierre cuando main() termine,
	// incluso si ocurre un error más adelante en el programa.
	defer conn.Close()

	// --- Apertura del Channel ---
	// Un channel es una sesión de comunicación lógica dentro de la conexión TCP.
	// Una sola conexión puede tener múltiples channels simultáneos, lo cual es
	// eficiente porque abrir conexiones TCP es costoso. Aquí solo usamos uno,
	// pero en apps más grandes cada goroutine tendría el suyo.
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	// --- Declaración de la cola ---
	// Le indica a RabbitMQ que necesitamos esta cola. Si ya existe con los mismos
	// parámetros, no pasa nada; si no existe, la crea. Esto permite que el publisher
	// arranque sin importar si el consumer ya está corriendo o no.
	_, err = ch.QueueDeclare(
		queueName, // Nombre de la cola
		true,      // durable: la cola sobrevive si RabbitMQ se reinicia (se guarda en disco)
		false,     // auto-delete: NO se borra automáticamente cuando no hay consumidores
		false,     // exclusive: NO es exclusiva de esta conexión (otros pueden usarla)
		false,     // no-wait: esperar confirmación del broker antes de continuar
		nil,       // arguments: argumentos adicionales (TTL, límite de mensajes, etc.)
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// --- Servidor de salud en goroutine ---
	// "go" lanza startHTTPServer() en paralelo (goroutine) sin bloquear main().
	// Kubernetes usa el endpoint /health para saber si el pod está vivo (liveness probe).
	// Si el pod no responde, Kubernetes lo reinicia automáticamente.
	go startHTTPServer()

	// --- Ticker: temporizador periódico ---
	// time.NewTicker crea un reloj que genera un "tick" cada publishInterval segundos.
	// ticker.C es un canal de Go que recibe un valor en cada tick.
	ticker := time.NewTicker(publishInterval)
	defer ticker.Stop()

	// El for range bloquea aquí y ejecuta el cuerpo cada vez que llega un tick.
	// Esto mantiene al publisher corriendo indefinidamente.
	for range ticker.C {
		// --- Publicación del mensaje ---
		// ch.Publish envía un mensaje al broker a través del exchange.
		err = ch.Publish(
			"",        // exchange: cadena vacía = default exchange de RabbitMQ.
			//           El default exchange enruta el mensaje directamente a la cola
			//           cuyo nombre coincide con la routing key.
			queueName, // routing key: con el default exchange, es el nombre de la cola destino
			false,     // mandatory: si no hay destino, NO retornar el mensaje al publisher
			false,     // immediate: NO requerir que haya un consumidor activo en este momento
			amqp.Publishing{
				ContentType: "text/plain",    // tipo de contenido del mensaje
				Body:        []byte(message), // cuerpo del mensaje como slice de bytes
			},
		)
		if err != nil {
			// Si falla la publicación, se registra el error pero se continúa con el
			// siguiente tick en lugar de matar el programa. Esto hace al publisher robusto
			// ante fallos temporales de red o del broker.
			log.Printf("failed to publish message: %v", err)
			continue
		}
		log.Printf("Message published to queue %q: %s", queueName, message)
	}
}

// startHTTPServer expone un endpoint /health que responde "ok".
// Kubernetes lo consulta periódicamente como liveness/readiness probe:
// si no responde, considera el pod como no saludable y lo reinicia.
func startHTTPServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Println("Publisher health server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("health server failed: %v", err)
	}
}

// envOrDefault lee una variable de entorno por su nombre.
// Si la variable no existe o está vacía, devuelve defaultValue.
// Centralizar esta lógica evita repetir os.Getenv + chequeo en cada variable.
func envOrDefault(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}