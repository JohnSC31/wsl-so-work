

##  `README.md` — Servidor HTTP desde cero (Go + Sockets)

---

###  Instalación y Ejecución

#### Requisitos

- Sistema operativo: **Unix-like** (Ubuntu recomendado, o WSL en Windows).
- **Go instalado** (`go version` debe mostrar v1.18 o superior).
- Terminal compatible con `curl` o Postman.

#### Clonación del repositorio

```bash
git clone https://github.com/JohnSC31/wsl-so-work.git
cd wsl-so-work
```

####  Ejecutar el servidor

```bash
go run .
```

El servidor se ejecutará en `localhost:8080`.


---

### Rutas Disponibles y Parámetros (Actualizado)

| Ruta                    | Descripción                                                                 | Parámetros                                     |
|-------------------------|-----------------------------------------------------------------------------|------------------------------------------------|
| `/help`                 | Lista los comandos disponibles.                                             | Ninguno                                        |
| `/timestamp`            | Devuelve la hora actual en formato ISO 8601.                                | Ninguno                                        |
| `/fibonacci`            | Calcula el Fibonacci recursivamente.                                        | `num=N` (entero positivo)                      |
| `/createfile`           | Crea un archivo con contenido repetido en `./files`.                        | `name`, `content`, `repeat`                    |
| `/deletefile`           | Elimina un archivo dentro de `./files`.                                     | `name`                                         |
| `/reverse`              | Devuelve el texto invertido.                                                | `text=abc`                                     |
| `/toupper`              | Convierte el texto a mayúsculas.                                            | `text=abc`                                     |
| `/random`               | Genera un arreglo de `n` números aleatorios entre `min` y `max`.            | `count=n`, `min=a`, `max=b`                    |
| `/hash`                 | Devuelve el hash SHA-256 del texto de entrada.                              | `text=abc`                                     |
| `/simulate`             | Simula una tarea ficticia con retardo de `s` segundos.                      | `seconds=s`, `task=name`                       |
| `/sleep`                | Simula latencia (sin procesamiento real).                                   | `seconds=s`                                    |
| `/loadtest`             | Ejecuta `n` tareas simuladas en paralelo, cada una con `x` segundos de retardo. | `tasks=n`, `sleep=x`                        |
| `/status`               | Retorna el estado actual del servidor en JSON.                              | Ninguno                                        |

---

### Ejemplos de uso

```bash
curl "http://localhost:8080/help"

curl "http://localhost:8080/timestamp"
curl "http://localhost:8080/fibonacci?num=10"
curl "http://localhost:8080/createfile?name=test&content=hello&repeat=10"
curl "http://localhost:8080/deletefile?name=test"
curl "http://localhost:8080/reverse?text=abcdefg"
curl "http://localhost:8080/toupper?text=holamundo"
curl "http://localhost:8080/random?count=5&min=10&max=100"
curl "http://localhost:8080/hash?text=hola"
curl "http://localhost:8080/simulate?seconds=2&task=fibonacci"
curl "http://localhost:8080/sleep?seconds=5"
curl "http://localhost:8080/loadtest?tasks=3&sleep=2"

curl "http://localhost:8080/status"
```

### Arquitectura del Sistema

El servidor está organizado por diferentes estructuras:

La estructura **Server** contiene las metricas (TiempoInicio, TotalRequests, ActWorkers), las pools de los workers, una es para tareas rapidas y la otra para tareas cortas.

Cada **WorkerPool** contiene una lista de workers y un canal de request que recibe y procesa las solicitudes. Cada **Worker** ejecuta la tarea correspondiente y luego se libera para que pueda recibir otra tarea.

#### Estructura del proyecto

```
project1/
├── main.go                  # Punto de entrada
├── go.mod                   # Definición del módulo
├── worker.go                # Definicion del worker
├── workerPool.go            # Definicion del la piscina de trabajadores
├── handlers/                # Lógica de cada ruta
│   ├── createfile.go
│   ├── deletefile.go
│   ├── fibonacci.go
│   ├── hash.go
│   ├── help.go
│   ├── loadtest.go
│   ├── random.go
│   ├── reverse.go
│   ├── simulate.go
│   ├── sleep.go
│   ├── timestamp.go
│   ├── toupper.go
├── utils/                   # Funciones auxiliares
│   ├── parser.go
└── files/                   # Carpeta donde se crean archivos
```

#### Detalles técnicos

- **Concurrencia**: cada conexión entrante se maneja con una `goroutine`.
- **Sin dependencias externas**: todo está hecho desde cero con `net` y estructuras estándar de Go.
- **Mutex global**: se usa un `sync.Mutex` para proteger operaciones en la carpeta `files/`.
- **Estado del servidor**: se mantienen métricas como PID principal, uptime, total de conexiones y lista de workers (simulados).
- **Formato de respuesta**: rutas como `/status` retornan salida en JSON para fácil integración con otras herramientas.

---

### Notas adicionales

- Todas las rutas usan el método `GET`.
- Las respuestas siguen el protocolo HTTP/1.0.
- No se usa el paquete `net/http` de Go: la implementación está construida manualmente usando sockets TCP.