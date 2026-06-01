# Instalación y configuración de Docker
## Instalación de Docker Linux
Para instalar Docker en Linux, se recomienda seguir la documentación oficial de Docker, ya que la instalación
puede variar dependiendo de la distribución de Linux que se esté utilizando.

### Instalación de Docker en Ubuntu (oficial)
1. Desinstala versiones antiguas (si existen):

```bash
for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done
```
2. Instala dependencias y agrega el repositorio oficial:

```bash
# Agregar llave GPG oficial de Docker:
sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Agregar el repositorio a los recursos Apt:
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu noble stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
```

3. Instalar Docker Engine y complementos:
```bash
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

4. Verificación de la instalación:
```bash
sudo docker run hello-world
```
Si ven un mensaje de bienvenida, Docker está correctamente instalado.

### Explicación del Docker Engine
#### Arquitectura del Docker Engine
El Docker Engine es una aplicación cliente-servidor con estos componentes:
- Un servidor que es un tipo de "demonio" que se ejecuta en la máquina host.
- Una API REST que especifica interfaces que los programas pueden usar para hablar con el "demonio" y darle instrucciones.
#### ¿Qué es un Deamon (demonio)?
Un demonio es un programa que se ejecuta en segundo plano, sin interacción directa con el usuario y se utilizan para realizar tareas de mantenimiento y administración del sistema, como la gestión de servicios, la programación de tareas y la monitorización del sistema.

En el caso de Docker, el demonio es el servidor de Docker que se ejecuta en la máquina host y se encarga de gestionar los contenedores y las imágenes de Docker.

## Docker Group
```bash
#Crear el grupo
sudo groupadd docker
# Agregar el usuario al grupo (Reemplazar $USER)
sudo usermod -aG docker $USER
# Activar el grupo
newgrp docker
```

## Ejemplos de Docker
### Primeros Comandos
```bash
# Verificar instalación
docker --version
# Descargar y ejecutar un contenedor 
docker run hello-world
# Listar contenedores en ejecución
docker ps
# Listar todos los contenedores (incluyendo detenidos)
docker ps -a
# Listar imagenes descargadas
docker images
                
```

### Primer Dockerfile
```dockerfile
# Dockerfile
FROM nginx:alpine
COPY index.html /usr/share/nginx/html/
EXPOSE 80
```

```bash
# Construir la imagen
docker build -t mi-web .

# Ejecutar el contenedor
# Ejecuta el contenedor "mi-web" en segundo plano (-d) y se mapea el puerto 80 del contenedor al puerto 8080 del host (-p 8080:80), con el nombre "web-container"
docker run -d -p 8080:80 --name web-container mi-web

# Ver logs
docker logs web-container

# Detener y eliminar
docker stop web-container
docker rm web-container
```

### Manejo de Imágenes y Contenedores
#### Trabajo con imágenes

```bash
# Descargar imagen específica 
docker pull ubuntu:20.04

#Ejecutar contenedor interactivo
docker run -it ubuntu:20.04 /bin/bash

# Desde otra terminal, ejecutar comandos en contenedor en ejecución
docker exec -it web-container /bin/sh

# Inspeccionar imagen/container
docker inspect web-container

#Ver historial de imagen
docker history mi-web   
```

#### Dockerfile Intermedio
```dockerfile
# Dockerfile para aplicación Python
From python:3.11-slim

# Establecer directorio de trabajo
WORKDIR /app

# Copiar requirements primero (para aprovechar el caché de Docker)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar el resto de la aplicación
COPY . .

# Variables de entorno
ENV PYTHONUMBUFFERED=1
ENV PORT=8080

# Exponer puerto
EXPOSE 8080

# Comandos para ejecutar
CMD ["python","app.py"]
```
Primero se selecciona la versión slim de Python 3.9,se define /app como la carpeta de trabajo e instala las dependencias del proyecto desde requirements.txt. Luego copia todo el código de la aplicación, configura dos variables de entorno (una para ver los logs en tiempo real y otra para definir el puerto), abre el puerto 8080 para recibir conexiones y finalmente indica que al iniciar el contenedor se debe ejecutar el archivo app.py.

Comandos para construir y ejecutar:
```bash
docker build -t mi-app-python .

# -p PUERTO_HOST:PUERTO_CONTENEDOR
# "todo lo que llegue al puerto 8080 de la máquina, redirigelo al puerto 8080 del contenedor".
docker run -p 8080:8080 mi-app-python

# Accede desde la máquina en localhost:3000
# pero la app dentro del contenedor escucha en el 8080
docker run -p 3000:8080 mi-app-python
```