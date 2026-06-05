# Grafana
Grafana es una plataforma open source escrita en Go que es utilizada normalmente para procesos de monitorear infreastructuras y/o aplicaciones IT.

## Como crear un contenedor de Grafana

Descargamos la imagen con:
```bash
docker pull grafana/grafana
```

Creamos un volumen para evitar perder datos y dashboards al reiniciar el contenedor:

```bash
docker volume create grafana-storage
```

Levantamos el contendor:

```bash
docker run -d --name grafana -p 3000:3000 -v grafana-storage:/var/lib/grafana grafana/grafana
```

| Parámetro | Descripción |
|---|---|
| `-d` | Corre en segundo plano  |
| `--name grafana` | Nombre del contenedor |
| `-p 3000:3000` | Expone el puerto 3000 del host |
| `-v grafana-storage:/var/lib/grafana` | Monta el volumen persistente |

Para verificar que está corriendo:

```bash
docker ps
```

Considerando los parámetros utilizados, se accede al dashboard con:
```
http://localhost:3000
```

- **Usuario:** `admin`
- **Contraseña:** `admin`

> Al primer inicio les pedirá cambiar la contraseña.

---

## Comandos útiles

```bash
# Detener el contenedor
docker stop grafana

# Volver a iniciarlo
docker start grafana

# Ver los logs
docker logs -f grafana

# Eliminar el contenedor (el volumen se conserva)
docker rm -f grafana

# Eliminar el volumen
docker volume rm grafana-storage
```

## Usando Docker Compose (alternativa)

Creamos un archivo `docker-compose.yml`:

```yaml
services:
  grafana:
    image: grafana/grafana
    container_name: grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana-storage:/var/lib/grafana
    restart: unless-stopped

volumes:
  grafana-storage:
```

Luego se ejecuta:

```bash
docker compose up -d
```

# Valkey
Valkey es un fork de Redis y es una base de datos en memoria y caché de código abierto. 

## Visualizar datos de Valkey en Grafana

### Arquitectura

```
Valkey → redis_exporter → Prometheus → Grafana
```

Grafana no lee Valkey directamente. El flujo estándar es:

1. **redis_exporter** expone las métricas de Valkey en formato Prometheus
2. **Prometheus** recolecta las métricas
3. **Grafana** consulta Prometheus como fuente de datos

---

Iniciamos creando la red Docker compartida

```bash
docker network create monitoring
```

Luego configuramos el Docker Compose completo:

```yaml
services:

  valkey:
    image: valkey/valkey:latest
    container_name: valkey
    ports:
      - "6379:6379"
    networks:
      - monitoring

  redis_exporter:
    image: oliver006/redis_exporter:latest
    container_name: redis_exporter
    environment:
      REDIS_ADDR: "valkey:6379"
    ports:
      - "9121:9121"
    depends_on:
      - valkey
    networks:
      - monitoring

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"
    depends_on:
      - redis_exporter
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana-storage:/var/lib/grafana
    depends_on:
      - prometheus
    networks:
      - monitoring

volumes:
  grafana-storage:

networks:
  monitoring:
    external: true
```

---

Se Configura Prometheus

Crea el archivo `prometheus.yml` en el mismo directorio:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "valkey"
    static_configs:
      - targets: ["redis_exporter:9121"]
```

---

Levantamos todo con:

```bash
docker compose up -d
```

Verifica que los cuatro contenedores estén corriendo:

```bash
docker ps
```

---

Agregamos Prometheus como fuente de datos en Grafana

1. Abrimos `http://localhost:3000`:
2. Vamos a **Connections → Data sources → Add new data source**
3. Seleccionamos **Prometheus**
4. En el campo URL ingresamos:
   ```
   http://prometheus:9090
   ```
5. Guardamos con **Save & test**

# Escritura en Valkey
En el contexto del proyecto, es necesario leer el archivo "/proc" para escribir en valkey, todo con Golang. Considerando esto, es necesario correr los siguientes comandos en el directorio de nuestro archivo de golang que lee y escribe en proc.

```bash
# Inicializar módulo e instalar dependencia
go mod init lector
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go run lector.go

# Para verificar si funciona
curl http://localhost:9200/metrics | grep sysinfo
```

# Queries y gráficas recomendadas para sysinfo en Grafana


### RAM usada vs libre — Time series

```promql
sysinfo_ram_used_kb / 1024
sysinfo_ram_free_kb / 1024
```

Dos líneas apiladas (`Stack: Normal`). Muestra la evolución de memoria a lo largo del tiempo.

---

### % RAM usada — Gauge

```promql
sysinfo_ram_used_kb / sysinfo_ram_total_kb * 100
```

Gauge con umbrales: verde < 60%, amarillo < 85%, rojo > 85%.

---

### Total de procesos — Stat

```promql
sysinfo_process_count
```

Panel tipo `Stat` con un número grande. 

---

### Top 10 por CPU — Bar chart horizontal

```promql
topk(10, max by (name) (sysinfo_process_cpu_percent))
```

`max by (name)` colapsa los PIDs en un solo valor por nombre de proceso, eliminando el
problemas de mostrar mas del top necesario. Tipo `Bar chart`, orientación horizontal, Query options → **Instant**.

---

### Top 10 por Memoria — Bar chart horizontal

```promql
topk(10, max by (name) (sysinfo_process_memory_percent))
```

Mismo patrón que CPU. Query options → **Instant**.

---

### Tabla de todos los procesos — Table

```promql
max by (pid, name, cmdline) (sysinfo_process_rss_kb)
```

Tipo `Table`. En *Transform* agregar `Join by field` y `Organize fields` para mostrar
PID, Name, RSS, VSZ, CPU% y Mem% en columnas ordenables.

---

## Nota importante

Siempre activar **Instant** en los bar charts en las opciones de la query, así Prometheus devuelve un solo valor por serie en lugar de un rango, y `topk` muestra solo el top necesario sin agregar más.