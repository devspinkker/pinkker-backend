# Usar la imagen base de Go 1.20
FROM golang:1.20-alpine as build

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos de la aplicación
COPY . .

# Instalar las dependencias necesarias (FFmpeg y otras)
RUN apk add --no-cache ffmpeg

# Descargar las dependencias y construir la aplicación
RUN go mod download
RUN go build -o main .

# Crear una imagen mínima para ejecutar la aplicación
FROM alpine:latest

# Instalar FFmpeg en la imagen final
RUN apk add --no-cache ffmpeg

# Copiar el ejecutable desde la fase de construcción
COPY --from=build /app/main .

# Copiar el archivo .env al contenedor
COPY --from=build /app/.env .env

# Exponer el puerto en el que se ejecutará la aplicación
EXPOSE 8084

# Definir el comando de inicio
CMD ["./main"]
