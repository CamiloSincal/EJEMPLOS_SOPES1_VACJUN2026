# Dockerfile para aplicación Python
FROM python:3.11-slim

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