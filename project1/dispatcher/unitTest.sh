#!/bin/bash

URL="http://localhost:8080"

echo "📡 INICIANDO PRUEBAS DEL SERVIDOR HTTP"
echo "---------------------------------------"

# Función utilitaria para validar respuestas
function test() {
    local desc="$1"
    local command="$2"
    echo -e "\n🔸 $desc"
    echo "→ $command"
    response=$(eval "$command")
    echo "$response"
    echo "---------------------------------------"
}

test "Ping del servidor" "curl -s $URL/ping"
# /help
test "Ayuda del servidor" "curl -s $URL/help"

# /timestamp
test "Timestamp actual" "curl -s $URL/timestamp"

# /fibonacci?num=10
test "Fibonacci de 10" "curl -s \"$URL/fibonacci?num=10\""
# /fibonacci?num=40
test "Fibonacci de 40" "curl -s \"$URL/fibonacci?num=40\""
# ------------------ ERROR ----------------
# /fibonacci?num=string
test "ERROR: Fibonacci de string" "curl -s \"$URL/fibonacci?num=string\""


# /createfile
test "Crear archivo prueba.txt" \
  "curl -s \"$URL/createfile?name=prueba.txt&content=hola&repeat=3\""
# /createfile
test "Crear archivo prueba2.txt" \
  "curl -s \"$URL/createfile?name=prueba2.txt&content=adios&repeat=20\""

# ------------------ ERROR ----------------
# /createfile
test "ERROR: Crear archivo sin nombre" \
  "curl -s \"$URL/createfile?name=&content=un100profe&repeat=100\""
# /createfile
test "ERROR: Crear archivo con repeat 0" \
  "curl -s \"$URL/createfile?name=pruebaERR2.txt&content=un100profe&repeat=0\""

# /deletefile
test "Eliminar archivo prueba.txt" \
  "curl -s \"$URL/deletefile?name=prueba.txt\""
# ------------------- ERROR ----------------------
# /deletefile
test "ERROR: Eliminar archivo no existe" \
  "curl -s \"$URL/deletefile?name=prueba.txt\""

# /reverse
test "Invertir texto 'hola mundo'" \
  "curl -s \"$URL/reverse?text=hola%20mundo\""

# ---------------- ERROR ----------------------
# /reverse
test "ERROR: Invertir texto sin texto" \
  "curl -s \"$URL/reverse?text=\""

# /toupper
test "Convertir 'hello world' a mayúsculas" \
  "curl -s \"$URL/toupper?text=hello%20world\""
# -------------- ERR ----------------------------
test "ERROR: Convertir '' a mayúsculas" \
  "curl -s \"$URL/toupper?text=\""
# ERR ----------------------------
test "ERROR: Convertir '12345' a mayúsculas" \
  "curl -s \"$URL/toupper?text=12345\""

# /random
test "Generar 5 números entre 1 y 100" \
  "curl -s \"$URL/random?count=5&min=1&max=100\""
# ----------------- Errores --------------------
# /random
test "ERROR: Generar 0 números entre 1 y 100" \
  "curl -s \"$URL/random?count=0&min=1&max=100\""
# /random
test "ERROR: Generar 3 números entre -10 y 10" \
  "curl -s \"$URL/random?count=3&min=-10&max=10\""
# /random
test "ERROR: Generar 3 números entre 10 y -10" \
  "curl -s \"$URL/random?count=3&min=10&max=-10\""
# /random
test "ERROR: Generar 3 números entre 5 y 2 (min 5 y max 2)" \
  "curl -s \"$URL/random?count=0&min=5&max=2\""
# /random
test "ERROR: Generar 3 números entre vacio y 100" \
  "curl -s \"$URL/random?count=3&min=&max=100\""


# /hash
test "Hash SHA-256 de 'test123'" \
  "curl -s \"$URL/hash?text=test123\""
# ----------------- Errores --------------------
# /hash
test "ERROR: Hash SHA-256 de ''" \
  "curl -s \"$URL/hash?text=\""

# /simulate
test "Simular tarea lenta de 3 segundos" \
  "curl -s \"$URL/simulate?seconds=3&task=heavy\""
# ----------------- Errores --------------------
# /simulate
test "ERROR: Simular tarea lenta de 0 segundos" \
  "curl -s \"$URL/simulate?seconds=0&task=lightweightbaby\""

# /sleep
test "Simular retardo simple de 2 segundos" \
  "curl -s \"$URL/sleep?seconds=2\""
# ----------------- Errores --------------------
# /sleep
test "ERROR: Simular retardo simple de 0 segundos" \
  "curl -s \"$URL/sleep?seconds=0\""
# /sleep
test "ERROR: Simular retardo simple de -1 segundos" \
  "curl -s \"$URL/sleep?seconds=-1\""
  # /sleep
test "ERROR: Simular retardo simple de '' segundos" \
  "curl -s \"$URL/sleep?seconds=\""

# /loadtest
test "Cargar 10 tareas de 5 segundo en paralelo" \
  "curl -s \"$URL/loadtest?tasks=10&sleep=5\""
# ----------------- Errores -------------------- 
# /loadtest
test "ERROR: Cargar '' tareas de '' segundo en paralelo" \
  "curl -s \"$URL/loadtest?tasks=&sleep=\""

# /status
test "Estado del servidor (JSON) requiere jq (sudo apt install jq)" "curl -s $URL/status | jq"

echo "✅ PRUEBAS COMPLETADAS"
