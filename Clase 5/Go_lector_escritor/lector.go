package main

// Importamos los paquetes necesarios:
// - "encoding/json": para convertir el JSON del archivo /proc/sysinfo a estructuras Go
// - "fmt": para formatear strings (ej: convertir el PID a texto)
// - "log": para imprimir mensajes de error/info en la consola
// - "net/http": para crear el servidor web que Prometheus consultará
// - "os": para leer el archivo /proc/sysinfo del sistema
// - "time": para configurar tiempos límite del servidor HTTP
// - "github.com/prometheus/...": librería oficial de Prometheus para Go
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Ruta del archivo donde el módulo del kernel expone la información del sistema
// con datos de RAM y procesos activos.
const procPath = "/proc/sysinfo"

// Process representa un proceso del sistema operativo con sus métricas de uso de recursos.
// Las etiquetas `json:"..."` le dicen a Go cómo mapear cada campo del JSON a esta estructura.
type Process struct {
	PID         int     `json:"PID"`          // ID único del proceso
	Name        string  `json:"Name"`         // Nombre del ejecutable (ej: "nginx", "bash")
	Cmdline     string  `json:"Cmdline"`      // Comando completo con argumentos
	VSZ         uint64  `json:"vsz"`          // Memoria virtual total reservada (en KB)
	RSS         uint64  `json:"rss"`          // Memoria RAM física que realmente usa (en KB)
	MemoryUsage float64 `json:"Memory_Usage"` // Porcentaje de RAM total que consume
	CPUUsage    float64 `json:"CPU_Usage"`    // Porcentaje de CPU que está usando
}

// SysInfo representa el JSON completo que expone el módulo del kernel.
// Contiene métricas globales del sistema más una lista de todos los procesos activos.
type SysInfo struct {
	TotalRAM  uint64    `json:"Totalram"`   // RAM total del sistema en KB
	FreeRAM   uint64    `json:"Freeram"`    // RAM disponible en KB
	Procs     int       `json:"Procs"`      // Cantidad total de procesos corriendo
	Processes []Process `json:"Processes"`  // Lista detallada de cada proceso
}

// SysInfoCollector es nuestra estructura principal que actúa como "exportador" de Prometheus.
//
// ¿Cómo funciona Prometheus?
// Prometheus es una base de datos de series de tiempo y en lugar de que nosotros
// le enviemos datos, Prometheus hace "scraping": nos consulta periódicamente (ej: cada 15s)
// pidiendo las métricas actuales. Para responder esas consultas, necesitamos implementar
// la interfaz prometheus.Collector, que exige dos métodos: Describe y Collect.
//
// Cada campo aquí es un *prometheus.Desc (descriptor), que es como el "molde" o
// "definición" de una métrica: su nombre, descripción y etiquetas. El valor real
// se asigna en el método Collect().
type SysInfoCollector struct {
	totalRAM   *prometheus.Desc // Descriptor para la RAM total del sistema
	freeRAM    *prometheus.Desc // Descriptor para la RAM libre
	usedRAM    *prometheus.Desc // Descriptor para la RAM usada (calculada: total - libre)
	procCount  *prometheus.Desc // Descriptor para el conteo total de procesos
	procVSZ    *prometheus.Desc // Descriptor para la memoria virtual por proceso
	procRSS    *prometheus.Desc // Descriptor para la memoria física (RSS) por proceso
	procMemPct *prometheus.Desc // Descriptor para el % de memoria por proceso
	procCPUPct *prometheus.Desc // Descriptor para el % de CPU por proceso
}

// NewSysInfoCollector crea e inicializa el collector con todos sus descriptores de métricas.
// Esta función actúa como "constructor"
//
// prometheus.NewDesc(nombre, descripción, etiquetas_variables, etiquetas_constantes)
//   - nombre: cómo aparecerá la métrica en Prometheus/Grafana (ej: sysinfo_ram_total_kb)
//   - etiquetas_variables (labels): permiten filtrar por proceso específico en Grafana
//     Por ejemplo: sysinfo_process_rss_kb{pid="1234", name="nginx"}
func NewSysInfoCollector() *SysInfoCollector {
	// Etiquetas que identifican a qué proceso pertenece cada métrica.
	// En Grafana se podrá filtrar por ejemplo como: "muéstrame solo el proceso con name='nginx'"
	labels := []string{"pid", "name", "cmdline"}

	// Prefijo común para todas las métricas de este exportador.
	// Todas aparecerán como "sysinfo_..." en Prometheus y Grafana.
	ns := "sysinfo"

	return &SysInfoCollector{
		// Métricas globales del sistema (sin etiquetas, son valores únicos)
		totalRAM: prometheus.NewDesc(
			ns+"_ram_total_kb",
			"RAM total del sistema en KB", nil, nil,
		),
		freeRAM: prometheus.NewDesc(
			ns+"_ram_free_kb",
			"RAM libre del sistema en KB", nil, nil,
		),
		usedRAM: prometheus.NewDesc(
			ns+"_ram_used_kb",
			"RAM usada del sistema en KB", nil, nil,
		),
		procCount: prometheus.NewDesc(
			ns+"_process_count",
			"Número total de procesos", nil, nil,
		),

		// Métricas por proceso (con etiquetas pid, name, cmdline para identificar cada proceso)
		procVSZ: prometheus.NewDesc(
			ns+"_process_vsz_kb",
			"Memoria virtual del proceso en KB", labels, nil,
		),
		procRSS: prometheus.NewDesc(
			ns+"_process_rss_kb",
			"RSS del proceso en KB", labels, nil,
		),
		procMemPct: prometheus.NewDesc(
			ns+"_process_memory_percent",
			"Porcentaje de memoria del proceso", labels, nil,
		),
		procCPUPct: prometheus.NewDesc(
			ns+"_process_cpu_percent",
			"Porcentaje de CPU del proceso", labels, nil,
		),
	}
}

// Describe es el primer método requerido por la interfaz prometheus.Collector.
// Prometheus lo llama al inicio para saber qué métricas va a exponer este collector.
// Solo enviamos los descriptores (definiciones), no los valores todavía.
// El canal `ch` es como una "bandeja de salida" donde depositamos los descriptores.
func (c *SysInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalRAM
	ch <- c.freeRAM
	ch <- c.usedRAM
	ch <- c.procCount
	ch <- c.procVSZ
	ch <- c.procRSS
	ch <- c.procMemPct
	ch <- c.procCPUPct
}

// Collect es el segundo método requerido por prometheus.Collector.
// Prometheus lo llama cada vez que hace un "scrape" (consulta periódica).
// Aquí es donde leemos /proc/sysinfo, parseamos el JSON y enviamos
// los valores reales de cada métrica al canal `ch`.
func (c *SysInfoCollector) Collect(ch chan<- prometheus.Metric) {
	// Leemos el archivo /proc/sysinfo que expone nuestro módulo del kernel.
	// os.ReadFile devuelve el contenido como bytes crudos.
	raw, err := os.ReadFile(procPath)
	if err != nil {
		// Si no podemos leer el archivo (ej: el módulo del kernel no está cargado),
		// solo logueamos el error y salimos sin enviar métricas.
		// Prometheus marcará este scrape como fallido automáticamente.
		log.Printf("[ERROR] leyendo %s: %v", procPath, err)
		return
	}

	// Convertimos el JSON crudo (bytes) a nuestra estructura SysInfo.
	// Si el JSON tiene un formato incorrecto, logueamos el error y salimos.
	var info SysInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		log.Printf("[ERROR] parseando JSON: %v", err)
		return
	}

	// ── Métricas globales del sistema ──────────────────────────────────────────
	// prometheus.MustNewConstMetric crea un valor de métrica listo para enviar.
	// GaugeValue significa que es un valor que puede subir o bajar (vs CounterValue
	// que solo sube, usado para contadores como "total de requests").
	ch <- prometheus.MustNewConstMetric(c.totalRAM, prometheus.GaugeValue, float64(info.TotalRAM))
	ch <- prometheus.MustNewConstMetric(c.freeRAM, prometheus.GaugeValue, float64(info.FreeRAM))
	// La RAM usada no viene directamente en el JSON, la calculamos aquí.
	ch <- prometheus.MustNewConstMetric(c.usedRAM, prometheus.GaugeValue, float64(info.TotalRAM-info.FreeRAM))
	ch <- prometheus.MustNewConstMetric(c.procCount, prometheus.GaugeValue, float64(info.Procs))

	// ── Métricas por proceso ───────────────────────────────────────────────────
	// Iteramos sobre cada proceso del sistema y enviamos sus métricas individuales.
	// Cada métrica lleva las etiquetas (pid, name, cmdline) para identificar el proceso.
	for _, p := range info.Processes {
		// Convertimos el PID (número) a string porque las etiquetas de Prometheus son texto.
		pid := fmt.Sprintf("%d", p.PID)

		// Los argumentos finales (pid, p.Name, p.Cmdline) son los valores de las etiquetas
		// definidas en labels = []string{"pid", "name", "cmdline"} — deben ir en el mismo orden.
		ch <- prometheus.MustNewConstMetric(c.procVSZ, prometheus.GaugeValue, float64(p.VSZ), pid, p.Name, p.Cmdline)
		ch <- prometheus.MustNewConstMetric(c.procRSS, prometheus.GaugeValue, float64(p.RSS), pid, p.Name, p.Cmdline)
		ch <- prometheus.MustNewConstMetric(c.procMemPct, prometheus.GaugeValue, p.MemoryUsage, pid, p.Name, p.Cmdline)
		ch <- prometheus.MustNewConstMetric(c.procCPUPct, prometheus.GaugeValue, p.CPUUsage, pid, p.Name, p.Cmdline)
	}
}

func main() {
	// Creamos nuestro collector y lo registramos en el registro global de Prometheus.
	// A partir de este momento, Prometheus sabe que debe consultar este collector
	// cada vez que alguien pida las métricas.
	collector := NewSysInfoCollector()
	prometheus.MustRegister(collector)

	// Configuramos las rutas HTTP del servidor:

	// GET /metrics → endpoint principal que Prometheus consulta periódicamente.
	// promhttp.Handler() genera automáticamente la respuesta en el formato de texto
	// que Prometheus espera (una métrica por línea, con sus etiquetas y valor).
	http.Handle("/metrics", promhttp.Handler())

	// GET /health → endpoint simple para verificar que el servicio está vivo.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Responde con código HTTP 200
		w.Write([]byte("ok"))
	})

	addr := ":9200"
	log.Printf("Exponiendo métricas en http://0.0.0.0%s/metrics", addr)
	log.Printf("Leyendo %s en cada scrape de Prometheus", procPath)

	// Creamos el servidor HTTP con timeouts de seguridad:
	// - ReadTimeout: máximo 5s para leer una request entrante (evita conexiones colgadas)
	// - WriteTimeout: máximo 10s para escribir la respuesta (las métricas pueden ser grandes)
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Iniciamos el servidor. ListenAndServe bloquea el programa indefinidamente.
	// Si falla (ej: el puerto 9200 ya está en uso), log.Fatal imprime el error y termina.
	log.Fatal(srv.ListenAndServe())
}