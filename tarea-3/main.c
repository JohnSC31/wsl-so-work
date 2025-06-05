#include <stdio.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>

// #define MEMORY_SIZE 1024 * 1024  // 1MB de bloque de memoria
#define MEMORY_SIZE 1000
#define MAX_BLOCKS 100
#define SEARCH_ALGORITHM 3 // 1 First fit, 2 Best fit, 3 Worst fit

typedef struct
{
    char name[20]; // Nombre de la variable
    size_t size;   // Tamaño asignado
    size_t start;  // Inicio del bloque (offset en memoria)
    int isFree;    // 1 si está libre, 0 si está ocupado
} Block;

Block blockList[MAX_BLOCKS]; // Lista de bloques
int num_blocks = 0;
char *memory;

void init_memory()
{
    memory = malloc(MEMORY_SIZE);
    if (!memory)
    {
        perror("No se pudo asignar memoria principal");
        exit(1);
    }

    // Crear el primer bloque libre con toda la memoria
    strcpy(blockList[0].name, "");
    blockList[0].size = MEMORY_SIZE;
    blockList[0].start = 0;
    blockList[0].isFree = 1;
    num_blocks = 1;
}

// Buscar un bloque libre
int search_free_block(size_t size)
{

    if (SEARCH_ALGORITHM == 1)
    {
        // FIRST FIT
        for (int i = 0; i < num_blocks; i++)
        {
            if (blockList[i].isFree && blockList[i].size >= size)
            {
                return i;
            }
        }
    }

    if (SEARCH_ALGORITHM == 2)
    {
        int best_index = -1;
        size_t best_size = MEMORY_SIZE + 1; // Inicializar con un valor mayor que el máximo posible

        for (int i = 0; i < num_blocks; i++)
        {
            if (blockList[i].isFree && blockList[i].size >= size)
            {
                // Si se encuentra un bloque más pequeño que el mejor actual
                if (blockList[i].size < best_size)
                {
                    best_index = i;
                    best_size = blockList[i].size;
                }
            }
        }
        return best_index;
    }

    if (SEARCH_ALGORITHM == 3)
    {
        int worst_index = -1;
        size_t worst_size = 0;
        for (int i = 0; i < num_blocks; i++) {
            if (blockList[i].isFree && blockList[i].size >= size) {
                if (blockList[i].size > worst_size) {
                    worst_index = i;
                    worst_size = blockList[i].size;
                }
            }
        }
        return worst_index;
    }

    return -1;
}

// ALLOC recibe el nombre y el tamano del bloque por asignar
void alloc_block(const char *name, size_t size)
{
    int idx = search_free_block(size);
    if (idx == -1)
    {
        printf("No hay espacio para %s (%zu bytes)\n", name, size);
        return;
    }

    Block *b = &blockList[idx];

    // Dividir el bloque si sobra espacio
    if (b->size > size)
    {
        // mueve la lista un bloque hacia adelante
        for (int i = num_blocks; i > idx + 1; i--)
        {
            blockList[i] = blockList[i - 1];
        }

        blockList[idx + 1].start = b->start + size;
        blockList[idx + 1].size = b->size - size;
        blockList[idx + 1].isFree = 1;
        strcpy(blockList[idx + 1].name, "");

        num_blocks++;
    }

    b->size = size;
    b->isFree = 0;
    strncpy(b->name, name, sizeof(b->name) - 1);
    b->name[sizeof(b->name) - 1] = '\0';

    // Rellenar la memoria con el nombre
    for (size_t i = 0; i < size; i++)
    {
        memory[b->start + i] = name[0];
    }
}

// FREE recibe el nombre de una variable y lo libera
void free_block(const char *name)
{

    for (int i = 0; i < num_blocks; i++)
    {
        if (!blockList[i].isFree && strcmp(blockList[i].name, name) == 0)
        {
            blockList[i].isFree = 1;
            strcpy(blockList[i].name, "");

            // Rellenar con '-'
            for (size_t j = 0; j < blockList[i].size; j++)
            {
                memory[blockList[i].start + j] = '-';
            }

            // Intentar fusionar con siguiente
            if (i + 1 < num_blocks && blockList[i + 1].isFree)
            {
                blockList[i].size += blockList[i + 1].size;
                for (int j = i + 1; j < num_blocks - 1; j++)
                {
                    blockList[j] = blockList[j + 1];
                }
                num_blocks--;
                i--;
            }

            // Intentar fusionar con anterior
            if (i > 0 && blockList[i - 1].isFree)
            {
                blockList[i - 1].size += blockList[i].size;
                for (int j = i; j < num_blocks - 1; j++)
                {
                    blockList[j] = blockList[j + 1];
                }
                num_blocks--;
                i--;
            }

            // printf("Liberado %s\n", name);
            return;
        }
    }

    // printf("Variable %s no encontrada\n", name);
}

// REALLOC recibe el nombre del bloque y el nuevo tamanno
void realloc_block(const char *name, size_t new_size)
{

    for (int i = 0; i < num_blocks; i++)
    {
        if (!blockList[i].isFree && strcmp(blockList[i].name, name) == 0)
        {
            // Guardar datos si es necesario (opcional)
            size_t old_size = blockList[i].size;

            // Liberar el actual
            free_block(name);

            // Asignar nuevo
            alloc_block(name, new_size);

            // printf("Reasignado %s de %zu a %zu bytes\n", name, old_size, new_size);
            return;
        }
    }

    // printf("Variable %s no encontrada\n", name);
}

void print_status()
{
    printf("\nEstado de la memoria (Total: %d bytes):\n", MEMORY_SIZE);
    printf("+-----------------------+---------------+\n");
    printf("| Bloque                | Tamanno (bytes)|\n");
    printf("+-----------------------+---------------+\n");

    size_t current_pos = 0;

    for (int i = 0; i < num_blocks; i++)
    {
        // Mostrar espacio libre antes del bloque si hay gap
        if (blockList[i].start > current_pos)
        {
            size_t free_size = blockList[i].start - current_pos;
            printf("| %-21s | %-13zu |\n", "[LIBRE]", free_size);
            printf("+-----------------------+---------------+\n");
        }

        // Mostrar el bloque actual
        printf("| %-21s | %-13zu |\n",
               blockList[i].name[0] ? blockList[i].name : "[LIBRE]",
               blockList[i].size);

        current_pos = blockList[i].start + blockList[i].size;

        // Mostrar línea divisoria si no es el último bloque
        if (i < num_blocks - 1)
        {
            printf("+-----------------------+---------------+\n");
        }
    }

    // Mostrar espacio libre al final si queda
    if (current_pos < MEMORY_SIZE)
    {
        printf("+-----------------------+---------------+\n");
        printf("| %-21s | %-13zu |\n", "[LIBRE]", MEMORY_SIZE - current_pos);
    }

    printf("+-----------------------+---------------+\n");

    // Estadísticas
    size_t used = 0;
    for (int i = 0; i < num_blocks; i++)
    {
        if (!blockList[i].isFree)
        {
            used += blockList[i].size;
        }
    }
    printf("\nResumen:\n");
    printf("- Memoria usada: %zu bytes (%.1f%%)\n", used, (float)used / MEMORY_SIZE * 100);
    printf("- Memoria libre: %zu bytes (%.1f%%)\n", MEMORY_SIZE - used, (float)(MEMORY_SIZE - used) / MEMORY_SIZE * 100);
}
void process_file(char *file_name)
{
    FILE *file = fopen(file_name, "r");
    if (!file)
    {
        perror("Error al abrir el archivo");
        exit(1);
    }
    init_memory();

    char fileLine[100];
    while (fgets(fileLine, sizeof(fileLine), file))
    {

        if (fileLine[0] == '#' || strlen(fileLine) < 2)
            continue;

        if (strncmp(fileLine, "ALLOC", 5) == 0)
        {
            // ALLOC block
            char name[20];
            size_t size;
            sscanf(fileLine, "ALLOC %s %zu", name, &size);
            alloc_block(name, size);
        }
        else if (strncmp(fileLine, "REALLOC", 7) == 0)
        {
            // REALLOC block
            char name[20];
            size_t new_size;
            sscanf(fileLine, "REALLOC %s %zu", name, &new_size);
            realloc_block(name, new_size);
        }
        else if (strncmp(fileLine, "FREE", 4) == 0)
        {
            // FREE block
            char name[20];
            sscanf(fileLine, "FREE %s %zu", name);
            free_block(name);
        }
        else if (strncmp(fileLine, "PRINT", 5) == 0)
        {
            // print
            // print_status();
        }
        printf(fileLine, '\n');
        print_status();
    }

    fclose(file);
}

int main(int argc, char *argv[])
{

    if (argc < 2)
    {
        printf("Por favor, proporciona el nombre del archivo como argumento.\n");
        return 1;
    }

    // Procesar el archivo
    process_file(argv[1]);

    return 0;
}