#!/bin/bash

# Configuración
IMAGE_NAME="project1-worker"
NETWORK_NAME="project1_app_network" 
DISPATCHER_URL="http://dispatcher:8080"

# Verificar si la imagen existe
if ! docker image inspect $IMAGE_NAME &> /dev/null; then
  echo "Construyendo imagen $IMAGE_NAME..."
  docker build -t $IMAGE_NAME -f ./server/Dockerfile ./server
fi

# Obtener el último número de worker, excluyendo contenedores detenidos
LAST_WORKER=$(docker ps -a --filter "name=worker" --format "{{.Names}}" | grep -oE '[0-9]+$' | sort -n | tail -1)
LAST_WORKER=${LAST_WORKER:-0}  # Si no hay workers, asigna 0
NEW_WORKER=$((LAST_WORKER + 1))

# Comprobar si el nombre del worker ya existe
while docker ps -a --filter "name=worker${NEW_WORKER}" --format "{{.Names}}" | grep -q "worker${NEW_WORKER}"; do
  NEW_WORKER=$((NEW_WORKER + 1))
done

# Levantar un nuevo worker con docker run
docker run -d \
  --hostname worker${NEW_WORKER} \
  --name worker${NEW_WORKER} \
  -e PORT=8080 \
  -e DISPATCHER_URL=$DISPATCHER_URL \
  -e WORKER_NAME=worker${NEW_WORKER} \
  --network $NETWORK_NAME \
  $IMAGE_NAME

echo "--------------------------------------------------"
echo " Worker ${NEW_WORKER} creado exitosamente"
echo " Hostname: worker${NEW_WORKER}"
echo " Nombre contenedor: worker${NEW_WORKER}"
echo "--------------------------------------------------"
