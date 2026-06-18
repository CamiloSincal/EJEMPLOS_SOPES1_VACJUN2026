use std::io::Cursor;
use tiny_http::{Server, Response, Header, Method, StatusCode};
use serde::{Deserialize, Serialize};

// Estructura del JSON que recibe el servidor
#[derive(Deserialize)]
struct Mensaje {
    usuario: String,
    pais: String,
    mensaje: String,
}

// Estructura del JSON que devuelve el servidor
#[derive(Serialize)]
struct Respuesta {
    usuario: String,
    pais: String,
    mensaje: String,
}

fn main() {
    // Descomentar para pruebas locales
    // let server = Server::http("127.0.0.1:8080").unwrap();
    // println!("Servidor en http://127.0.0.1:8080");


    // Descomentar para pruebas en k8s
    let server = Server::http("0.0.0.0:8080").unwrap();
    println!("Servidor en http://0.0.0.0:8080");


    for request in server.incoming_requests() {
        if request.url() == "/" || request.url() == "/health" {
            let body = b"ok";
            let _ = request.respond(Response::new(
                StatusCode(200), vec![], Cursor::new(body.to_vec()), Some(body.len()), None,
            ));
            continue;
        }
        // Solo aceptamos POST /messages
        if request.method() != &Method::Post || request.url() != "/messages" {
            let body = b"404 - Ruta no encontrada";
            let _ = request.respond(Response::new(
                StatusCode(404), vec![], Cursor::new(body.to_vec()), Some(body.len()), None,
            ));
            continue;
        }

        handle(request);
    }
}

fn handle(mut req: tiny_http::Request) {
    // Leer el body de la petición
    let mut body = String::new();
    req.as_reader().read_to_string(&mut body).unwrap();

    // Deserializar el JSON entrante
    let datos: Mensaje = serde_json::from_str(&body).unwrap();

    // Construir la respuesta
    let respuesta = Respuesta {
        usuario: datos.usuario,
        pais: datos.pais,
        mensaje: datos.mensaje,
    };

    let json = serde_json::to_string_pretty(&respuesta).unwrap();
    let bytes = json.into_bytes();
    let header = Header::from_bytes(b"Content-Type", b"application/json").unwrap();

    // Enviar respuesta 200 OK con el JSON
    let _ = req.respond(Response::new(
        StatusCode(200),
        vec![header],
        Cursor::new(bytes.clone()),
        Some(bytes.len()),
        None,
    ));
}