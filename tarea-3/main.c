#include <stdio.h>
#include <stddef.h>

#define MEMORY_SIZE 1024 * 1024  // 1MB de bloque de memoria
#define MAX_BLOQUES 100

typedef struct {
    char nombre[20];    // Nombre de la variable
    size_t size;        // Tamaño asignado
    size_t start;       // Inicio del bloque (offset en memoria)
    int libre;          // 1 si está libre, 0 si está ocupado
} Bloque;

Bloque bloques[MAX_BLOQUES];    // Lista de bloques
int num_bloques = 0;
char *memoria;

void inicializar_memoria() {
    memoria = malloc(MEMORY_SIZE);
    if (!memoria) {
        perror("No se pudo asignar memoria principal");
        exit(1);
    }

    // Crear el primer bloque libre con toda la memoria
    strcpy(bloques[0].nombre, ""); 
    bloques[0].size = MEMORY_SIZE;
    bloques[0].start = 0;
    bloques[0].libre = 1;
    num_bloques = 1;
}

int buscar_bloque_libre(size_t size) {
    for (int i = 0; i < num_bloques; i++) {
        if (bloques[i].libre && bloques[i].size >= size) {
            return i;
        }
    }
    return -1;
}

// ALLOC 
void asignar_bloque(const char *nombre, size_t size) {
    int idx = buscar_bloque_libre(size);
    if (idx == -1) {
        printf("No hay espacio para %s (%zu bytes)\n", nombre, size);
        return;
    }

    Bloque *b = &bloques[idx];

    // Dividir el bloque si sobra espacio
    if (b->size > size) {
        for (int i = num_bloques; i > idx + 1; i--) {
            bloques[i] = bloques[i - 1];
        }

        bloques[idx + 1].start = b->start + size;
        bloques[idx + 1].size = b->size - size;
        bloques[idx + 1].libre = 1;
        strcpy(bloques[idx + 1].nombre, "");

        num_bloques++;
    }

    b->size = size;
    b->libre = 0;
    strncpy(b->nombre, nombre, sizeof(b->nombre));

    // Rellenar la memoria con el nombre
    for (size_t i = 0; i < size; i++) {
        memoria[b->start + i] = nombre[0];
    }
}

// REALLOC
void realloc_bloque(const char *nombre, size_t nuevo_tamano) {
    for (int i = 0; i < num_bloques; i++) {
        if (!bloques[i].libre && strcmp(bloques[i].nombre, nombre) == 0) {
            // Guardar datos si es necesario (opcional)
            size_t viejo_tamano = bloques[i].size;

            // Liberar el actual
            liberar_bloque(nombre);

            // Asignar nuevo
            asignar_bloque(nombre, nuevo_tamano);

            printf("Reasignado %s de %zu a %zu bytes\n", nombre, viejo_tamano, nuevo_tamano);
            return;
        }
    }

    printf("Variable %s no encontrada\n", nombre);
}


// FREE
void liberar_bloque(const char *nombre) {

    for (int i = 0; i < num_bloques; i++) {
        if (!bloques[i].libre && strcmp(bloques[i].nombre, nombre) == 0) {
            bloques[i].libre = 1;
            strcpy(bloques[i].nombre, "");

            // Rellenar con '-'
            for (size_t j = 0; j < bloques[i].size; j++) {
                memoria[bloques[i].start + j] = '-';
            }

            // Intentar fusionar con siguiente
            if (i + 1 < num_bloques && bloques[i + 1].libre) {
                bloques[i].size += bloques[i + 1].size;
                for (int j = i + 1; j < num_bloques - 1; j++) {
                    bloques[j] = bloques[j + 1];
                }
                num_bloques--;
            }

            // Intentar fusionar con anterior
            if (i > 0 && bloques[i - 1].libre) {
                bloques[i - 1].size += bloques[i].size;
                for (int j = i; j < num_bloques - 1; j++) {
                    bloques[j] = bloques[j + 1];
                }
                num_bloques--;
            }

            printf("Liberado %s\n", nombre);
            return;
        }
    }

    printf("Variable %s no encontrada\n", nombre);
}

int first_fit(size_t size) {
    for (int i = 0; i < num_bloques; i++) {
        if (bloques[i].libre && bloques[i].size >= size) {
            return i;
        }
    }
    return -1;
}

void procesar_archivo(char *nombre_archivo) {
    FILE *archivo = fopen(nombre_archivo, "r");
    if (!archivo) {
        perror("Error al abrir el archivo");
        exit(1);
    }

    inicializar_memoria();

    char linea[100];
    while (fgets(linea, sizeof(linea), archivo)) {
        if (linea[0] == '#' || strlen(linea) < 2) continue;

        if (strncmp(linea, "ALLOC", 5) == 0) {
            char nombre[20]; size_t size;
            sscanf(linea, "ALLOC %s %zu", nombre, &size);
            asignar_bloque(nombre, size);

        } else if (strncmp(linea, "REALLOC", 7) == 0) {
            char nombre[20]; size_t new_size;
            sscanf(linea, "ALLOC %s %zu", nombre, &new_size);
            realloc_bloque(nombre, new_size);

        } else if (strncmp(linea, "FREE", 4) == 0) {
             char nombre[20];
            sscanf(linea, "ALLOC %s %zu", nombre);
            liberar_bloque(nombre);

        } else if (strncmp(linea, "PRINT", 5) == 0) {
            imprimir_estado();
        }
    }

    fclose(archivo);
}

void imprimir_estado() {
    printf("Estado de la memoria:\n");
    for (int i = 0; i < num_bloques; i++) {
        printf("[%s] Inicio: %zu, Tamaño: %zu, %s\n",
            bloques[i].nombre,
            bloques[i].start,
            bloques[i].size,
            bloques[i].libre ? "Libre" : "Ocupado");
    }
}

int main(int argc, char *argv[]) {

    if (argc < 2) {
        printf("Por favor, proporciona el nombre del archivo como argumento.\n");
        return 1;
    }

    // Procesar el archivo
    procesar_archivo(argv[1]);

    return 0;
}