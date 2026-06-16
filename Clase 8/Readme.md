# gRPC en Golang

## 1. Instalar Protocol Buffers (protoc)

### Linux

Se abre una terminal y actualiza los paquetes del sistema:

```bash
sudo apt update && sudo apt upgrade -y
```

Se instalan las herramientas necesarias para compilar protoc:

```bash
sudo apt install -y build-essential libtool pkg-config protobuf-compiler
```

Se puede verificar la instalación ejecutando:

```bash
protoc --version
```

Se debería ver la versión instalada de `protoc`.

### Windows

Descargamos los binarios de protoc en:

```
https://github.com/protocolbuffers/protobuf/releases
```

Nota: se debe descargar el archivo con terminación win según la arquitectura de cada computadora.

- Con el zip descargado se extraen los archivos.
- Ubicamos el folder en donde queramos.
- Abrimos la carpeta de binario y copiamos el path del protoc (aplicación)
- Abrimos las variables de entorno de windows
- En la sección de variables del sistema buscamos y seleccionamos path.
- Agregamos el path de protoc.

## 2. Instalar plugins de Go en gRPC

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 3. Generar código Go desde `.proto`

El archivo .proto es el corazón de cualquier proyecto gRPC. Es un archivo de definición de contrato que describe los servicios, métodos y mensajes que usarán tanto el cliente como el servidor.

- Se coloca el archivo `.proto` (por ejemplo, tweet.proto) en el directorio del proyecto.
- Generamos el código necesario ejecutando:

```bash
protoc --go_out=. --go-grpc_out=. tweet.proto
```

Esto generará dos archivos en /proto:
  - `tweet.pb.go`
  - `tweet_grpc.pb.go`

## 4. Configurar el código de Go

En el directorio del server y client:

```bash
go mod init grpc
# luego
go mod tidy
```

## 5. Se ejecuta el proyecto

Iniciamos el servidor en una terminal:

```bash
go run server.go
```

En otra terminal, ejecutamos el cliente:

```bash
go run client.go
```

# Ejemplo API REST en Golang — LibrosService


## Estructura del proyecto

```
rest-ejemplo/
├── main.go      # Servidor REST completo
└── README.md
```

## Endpoints

| Método   | URL            | Descripción              |
|----------|----------------|--------------------------|
| GET      | /libros        | Lista todos los libros   |
| POST     | /libros        | Crea un libro nuevo      |
| GET      | /libros/{id}   | Obtiene un libro por ID  |
| PUT      | /libros/{id}   | Actualiza un libro       |
| DELETE   | /libros/{id}   | Elimina un libro         |

---

## Pasos para ejecutar

### 1. Inicializar el módulo

```bash
go mod init rest
```

### 2. Ejecutar el servidor

```bash
go run main.go
```

Salida esperada:
```
Servidor REST escuchando en http://localhost:8080
Endpoints disponibles:
  GET    /libros
  POST   /libros
  GET    /libros/{id}
  PUT    /libros/{id}
  DELETE /libros/{id}
```

---

## Probar con curl

### Listar todos los libros
```bash
curl http://localhost:8080/libros
```

### Obtener un libro por ID
```bash
curl http://localhost:8080/libros/1
```

### Crear un libro nuevo
```bash
curl -X POST http://localhost:8080/libros \
  -H "Content-Type: application/json" \
  -d '{"titulo":"Clean Code","autor":"Robert C. Martin","anio":2008}'
```

### Actualizar un libro
```bash
curl -X PUT http://localhost:8080/libros/4 \
  -H "Content-Type: application/json" \
  -d '{"titulo":"Clean Code (Edicion revisada)","autor":"Robert C. Martin","anio":2022}'
```

### Eliminar un libro
```bash
curl -X DELETE http://localhost:8080/libros/4
```

---

## Salidas esperadas

**GET /libros/1**
```json
{
  "id": 1,
  "titulo": "El Señor de los Anillos",
  "autor": "J.R.R. Tolkien",
  "anio": 1954
}
```

**POST /libros**
```json
{
  "id": 4,
  "titulo": "Clean Code",
  "autor": "Robert C. Martin",
  "anio": 2008
}
```

**DELETE /libros/4**
```json
{
  "mensaje": "libro 4 eliminado"
}
```

**GET /libros/99 (no existe)**
```json
{
  "error": "libro 99 no encontrado"
}
```

---

## REST vs gRPC — Comparación

| Característica        | REST                              | gRPC                                      |
|-----------------------|-----------------------------------|-------------------------------------------|
| **Protocolo**         | HTTP/1.1                          | HTTP/2                                    |
| **Formato de datos**  | JSON (texto legible)              | Protocol Buffers (binario, compacto)      |
| **Contrato**          | Informal (docs, OpenAPI)          | Formal y obligatorio (archivo `.proto`)   |
| **Generación código** | Manual                            | Automática con `protoc`                   |
| **Rendimiento**       | Menor (JSON más pesado)           | Mayor (binario ~5-10x más compacto)       |
| **Streaming**         | No nativo                         | Nativo (unidireccional y bidireccional)   |
| **Facilidad uso**     | Alta (curl, browser, Postman)     | Media (requiere cliente gRPC)             |
| **Dependencias**      | Ninguna (stdlib de Go)            | protoc + plugins + librería grpc          |
| **Ideal para**        | APIs públicas, frontends, móviles | Microservicios internos, alto rendimiento |
