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

# /ping
test "Ping del servidor" "curl -s $URL/ping"

# /help
test "Ayuda del servidor" "curl -s $URL/help"

# /calculatepi?iterations=1000000000
test "Calcular Pi con 1,000,000,000 iteraciones" "curl -s $URL/calculatepi?iterations=1000000000"

# ------------------ ERROR ----------------
# /calculatepi?iterations=abc
test "ERROR: Calcular Pi con 'iterations' no numérico" "curl -s $URL/calculatepi?iterations=abc"

# /calculatepi?iterations=-100
test "ERROR: Calcular Pi con 'iterations' negativo" "curl -s $URL/calculatepi?iterations=-100"

# /calculatepi?iterations=0
test "ERROR: Calcular Pi con 'iterations' igual a 0" "curl -s $URL/calculatepi?iterations=0"

# ------------------ ERROR ----------------


# /countwords (archivo válido)
test "Calcular palabras de archivo válido (3500 líneas)" \
  "curl -X POST -H \"Content-Type: text/plain\" --data-binary \"@3500_lineas.txt\" $URL/countwords"

test "Estado de los workers" "curl -s $URL/workers"

echo "✅ PRUEBAS COMPLETADAS"
