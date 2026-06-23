package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// Librería oficial de RabbitMQ para Go.
	// El alias "amqp" nos permite escribir amqp.Dial(...) en lugar del nombre completo del paquete.
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// --- Configuración desde variables de entorno ---
	// Se leen variables de entorno para no hardcodear valores sensibles en el código.
	// Esto es fundamental en contenedores/Kubernetes: puedes cambiar el comportamiento
	// sin recompilar, solo ajustando las env vars en el manifiesto del pod.
	// A diferencia del publisher, el consumer no necesita intervalo ni mensaje,
	// solo saber a qué broker conectarse y qué cola escuchar.
	rabbitURL := envOrDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq-cluster.rabbitmq-system.svc.cluster.local:5672/")
	queueName := envOrDefault("RABBITMQ_QUEUE", "mensajes")

	// --- Conexión TCP a RabbitMQ ---
	// amqp.Dial establece la conexión física (TCP) con el broker.
	// La URL tiene el formato: amqp://usuario:contraseña@host:puerto/
	// El protocolo AMQP (Advanced Message Queuing Protocol) opera por defecto en el puerto 5672.
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error al conectar con RabbitMQ: %v", err)
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
		log.Fatalf("Error al abrir canal: %v", err)
	}
	defer ch.Close()

	// --- Declaración de la cola ---
	// Le indica a RabbitMQ que necesitamos esta cola. Si ya existe con los mismos
	// parámetros, no pasa nada; si no existe, la crea.
	// IMPORTANTE: tanto el publisher como el consumer declaran la misma cola con los
	// mismos parámetros. Esto garantiza que no importa cuál de los dos arranque primero:
	// el que llegue primero la crea, y el segundo simplemente la reutiliza.
	_, err = ch.QueueDeclare(
		queueName, // Nombre de la cola
		true,      // durable: la cola sobrevive si RabbitMQ se reinicia (se guarda en disco)
		false,     // auto-delete: NO se borra automáticamente cuando no hay consumidores
		false,     // exclusive: NO es exclusiva de esta conexión (otros pueden usarla)
		false,     // no-wait: esperar confirmación del broker antes de continuar
		nil,       // arguments: argumentos adicionales (TTL, límite de mensajes, etc.)
	)

	if err != nil {
		log.Fatalf("Error al declarar cola: %v", err)
	}

	// --- Registro del consumidor ---
	// ch.Consume le dice a RabbitMQ "quiero recibir mensajes de esta cola".
	// No bloquea el programa: simplemente registra al consumidor y devuelve "msgs",
	// que es un canal de Go (<-chan amqp.Delivery) desde donde llegarán los mensajes.
	msgs, err := ch.Consume(
		queueName, // Nombre de la cola a escuchar
		"",        // consumer tag: identificador único de este consumidor.
		//           Vacío = RabbitMQ genera uno automático. Útil para identificar
		//           consumidores específicos en el panel de administración de RabbitMQ.
		true,  // auto-ack: confirmar (acknowledger) mensajes automáticamente al recibirlos.
		//       Con true: RabbitMQ elimina el mensaje en cuanto lo entrega, sin importar
		//       si tu código lo procesó bien o no. Fácil para aprender, pero riesgoso
		//       en producción: si el consumer falla a mitad del procesamiento, el mensaje
		//       se pierde. Con false (manual ack), tú llamas d.Ack(false) cuando terminas.
		false, // exclusive: NO reservar la cola solo para este consumidor.
		//       Con true, ningún otro consumidor podría leer de esta cola.
		false, // no-local: parámetro del estándar AMQP que RabbitMQ ignora en la práctica.
		false, // no-wait: esperar confirmación del broker antes de continuar.
		nil,   // arguments: argumentos adicionales (prioridad de consumidor, etc.)
	)
	if err != nil {
		log.Fatalf("Error al registrar consumidor: %v", err)
	}

	// --- Servidor de salud en goroutine ---
	// "go" lanza startHTTPServer() en paralelo (goroutine) sin bloquear main().
	// Kubernetes usa el endpoint /health para saber si el pod está vivo (liveness probe).
	// Si el pod no responde, Kubernetes lo reinicia automáticamente.
	go startHTTPServer()

	log.Printf("Esperando mensajes en la cola %q", queueName)

	// --- Canal forever: bloqueo permanente de main() ---
	// make(chan bool) crea un canal vacío que nunca recibirá ningún valor.
	// Su único propósito es mantener main() vivo indefinidamente.
	// Sin esto, main() terminaría inmediatamente después de lanzar las goroutines
	// y el programa moriría sin procesar ningún mensaje.
	forever := make(chan bool)

	// --- Goroutine de procesamiento de mensajes ---
	// Se lanza en paralelo para no bloquear el hilo principal.
	// "func() { ... }()" es una función anónima que se ejecuta inmediatamente.
	go func() {
		// for d := range msgs bloquea esta goroutine esperando mensajes del canal msgs.
		// Cada vez que RabbitMQ entrega un mensaje, "d" es un amqp.Delivery que contiene:
		//   - d.Body: el cuerpo del mensaje como []byte
		//   - d.ContentType, d.Headers, d.RoutingKey, y otras propiedades del mensaje
		// El loop corre indefinidamente hasta que el canal msgs se cierre
		// (lo que ocurriría si la conexión con RabbitMQ se pierde).
		for d := range msgs {
			// Aquí iría la lógica de negocio real: parsear el mensaje, guardar en BD, etc.
			// En este ejemplo simplemente se imprime el cuerpo del mensaje.
			log.Printf("Mensaje recibido: %s", d.Body)
		}
	}()

	// <-forever intenta leer del canal forever, que nunca tendrá un valor.
	// Esto bloquea main() para siempre, manteniendo vivas las goroutines
	// del servidor HTTP y del procesamiento de mensajes.
	// Alternativas comunes a este patrón: select{} o esperar señales del SO con signal.Notify.
	<-forever
}

// startHTTPServer expone un endpoint /health que responde "ok".
// Kubernetes lo consulta periódicamente como liveness/readiness probe:
// si no responde, considera el pod como no saludable y lo reinicia.
func startHTTPServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Println("Server del consumidor escuchando en :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error en el servidor de Consumidor: %v", err)
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