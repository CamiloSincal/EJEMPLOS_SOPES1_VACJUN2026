package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// -------------------------------------------------------------------
// Modelo
// -------------------------------------------------------------------

type Libro struct {
	ID     int    `json:"id"`
	Titulo string `json:"titulo"`
	Autor  string `json:"autor"`
	Anio   int    `json:"anio"`
}

// -------------------------------------------------------------------
// "Base de datos" en memoria
// -------------------------------------------------------------------

type almacen struct {
	mu       sync.Mutex
	libros   map[int]Libro
	contador int
}

var db = &almacen{
	libros: map[int]Libro{
		1: {ID: 1, Titulo: "El Señor de los Anillos", Autor: "J.R.R. Tolkien", Anio: 1954},
		2: {ID: 2, Titulo: "Cien Años de Soledad", Autor: "Gabriel García Márquez", Anio: 1967},
		3: {ID: 3, Titulo: "Don Quijote de la Mancha", Autor: "Miguel de Cervantes", Anio: 1605},
	},
	contador: 3,
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func responderJSON(w http.ResponseWriter, status int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(datos)
}

func responderError(w http.ResponseWriter, status int, mensaje string) {
	responderJSON(w, status, map[string]string{"error": mensaje})
}

// Extrae el ID de la URL: /libros/42 → 42
func idDesdeURL(path string) (int, error) {
	partes := strings.Split(strings.Trim(path, "/"), "/")
	if len(partes) < 2 {
		return 0, fmt.Errorf("ID no proporcionado")
	}
	return strconv.Atoi(partes[len(partes)-1])
}

// -------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------

// GET /libros        → lista todos los libros
// POST /libros       → crea un libro nuevo
func handlerLibros(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		db.mu.Lock()
		defer db.mu.Unlock()

		lista := make([]Libro, 0, len(db.libros))
		for _, l := range db.libros {
			lista = append(lista, l)
		}
		log.Printf("GET /libros → %d libros", len(lista))
		responderJSON(w, http.StatusOK, lista)

	case http.MethodPost:
		var nuevo Libro
		if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
			responderError(w, http.StatusBadRequest, "JSON inválido")
			return
		}
		if nuevo.Titulo == "" || nuevo.Autor == "" {
			responderError(w, http.StatusBadRequest, "titulo y autor son requeridos")
			return
		}

		db.mu.Lock()
		db.contador++
		nuevo.ID = db.contador
		db.libros[nuevo.ID] = nuevo
		db.mu.Unlock()

		log.Printf("POST /libros → creado ID=%d (%s)", nuevo.ID, nuevo.Titulo)
		responderJSON(w, http.StatusCreated, nuevo)

	default:
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// GET    /libros/{id} → obtiene un libro
// PUT    /libros/{id} → actualiza un libro
// DELETE /libros/{id} → elimina un libro
func handlerLibroPorID(w http.ResponseWriter, r *http.Request) {
	id, err := idDesdeURL(r.URL.Path)
	if err != nil {
		responderError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	switch r.Method {

	case http.MethodGet:
		db.mu.Lock()
		libro, existe := db.libros[id]
		db.mu.Unlock()

		if !existe {
			responderError(w, http.StatusNotFound, fmt.Sprintf("libro %d no encontrado", id))
			return
		}
		log.Printf("GET /libros/%d → encontrado", id)
		responderJSON(w, http.StatusOK, libro)

	case http.MethodPut:
		var actualizado Libro
		if err := json.NewDecoder(r.Body).Decode(&actualizado); err != nil {
			responderError(w, http.StatusBadRequest, "JSON inválido")
			return
		}

		db.mu.Lock()
		_, existe := db.libros[id]
		if !existe {
			db.mu.Unlock()
			responderError(w, http.StatusNotFound, fmt.Sprintf("libro %d no encontrado", id))
			return
		}
		actualizado.ID = id
		db.libros[id] = actualizado
		db.mu.Unlock()

		log.Printf("PUT /libros/%d → actualizado", id)
		responderJSON(w, http.StatusOK, actualizado)

	case http.MethodDelete:
		db.mu.Lock()
		_, existe := db.libros[id]
		if !existe {
			db.mu.Unlock()
			responderError(w, http.StatusNotFound, fmt.Sprintf("libro %d no encontrado", id))
			return
		}
		delete(db.libros, id)
		db.mu.Unlock()

		log.Printf("DELETE /libros/%d → eliminado", id)
		responderJSON(w, http.StatusOK, map[string]string{"mensaje": fmt.Sprintf("libro %d eliminado", id)})

	default:
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// -------------------------------------------------------------------
// Main
// -------------------------------------------------------------------

func main() {
	http.HandleFunc("/libros", handlerLibros)
	http.HandleFunc("/libros/", handlerLibroPorID)

	log.Println("Servidor REST escuchando en http://localhost:8080")
	log.Println("Endpoints disponibles:")
	log.Println("  GET    /libros")
	log.Println("  POST   /libros")
	log.Println("  GET    /libros/{id}")
	log.Println("  PUT    /libros/{id}")
	log.Println("  DELETE /libros/{id}")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}
