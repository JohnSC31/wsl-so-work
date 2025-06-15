#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <ctype.h> // Para toupper

// --- Definiciones de Constantes ---
#define BLOCK_SIZE 512              // Tamaño de cada bloque en bytes
#define DISK_SIZE_MB 1              // Tamaño total del disco simulado en MB
#define MAX_FILES 100               // Número máximo de archivos que el sistema puede manejar
#define TOTAL_BLOCKS (DISK_SIZE_MB * 1024 * 1024 / BLOCK_SIZE) // Número total de bloques en el disco
#define MAX_FILENAME_LEN 32         // Longitud máxima para el nombre del archivo (incluyendo el nulo terminador)
                                    // Esto significa que el nombre real del archivo puede tener hasta 31 caracteres.
#define MAX_WRITE_DATA_LEN 512      // Longitud máxima de datos para una operación de escritura única

// --- Estructuras de Datos ---

// Estructura que representa una entrada de archivo en la tabla de archivos.
// Cada entrada almacena metadatos sobre un archivo.
typedef struct {
    char name[MAX_FILENAME_LEN]; // Nombre del archivo.
    int size;                    // Tamaño del archivo en bytes.
    int start_block;             // Índice del primer bloque asignado al archivo.
                                 // -1 indica que la entrada está libre (no asignada a un archivo).
    int num_blocks;              // Número total de bloques que ocupa el archivo.
} FileEntry;

// Tipo que simula un bloque de disco. Es un arreglo de caracteres del tamaño BLOCK_SIZE.
typedef char DiskBlock[BLOCK_SIZE];

// --- Variables Globales del Sistema de Archivos ---

FileEntry file_table[MAX_FILES];    // Tabla de archivos: Almacena las entradas para cada archivo.
DiskBlock disk[TOTAL_BLOCKS];       // Disco simulado: Arreglo de bloques de disco.
bool free_blocks[TOTAL_BLOCKS];     // Mapa de bits de bloques libres: 'true' si el bloque está libre, 'false' si está ocupado.
int num_files;                      // Contador del número actual de archivos en el sistema.

// --- Prototipos de Funciones ---

void init_file_system();
int find_file_entry_index(const char* name);
int allocate_blocks(int required_blocks);
void free_blocks_for_file(int start_block, int num_blocks);
void create_file(const char* name, int size);
void write_file(const char* name, int offset, const char* data);
void read_file(const char* name, int offset, int size_to_read);
void delete_file(const char* name);
void list_files();
void display_help();
void parse_command(char* command);

// --- Implementación de Funciones ---

/**
 * @brief Inicializa el sistema de archivos.
 * Reinicia la tabla de archivos, marca todos los bloques del disco como libres
 * y establece el contador de archivos a cero.
 */
void init_file_system() {
    // Inicializar la tabla de archivos: marcar todas las entradas como libres
    for (int i = 0; i < MAX_FILES; i++) {
        file_table[i].start_block = -1; // -1 indica que la entrada está libre
        file_table[i].size = 0;
        file_table[i].num_blocks = 0;
        memset(file_table[i].name, 0, MAX_FILENAME_LEN); // Limpiar el nombre
    }

    // Inicializar el mapa de bits de bloques libres: marcar todos los bloques como libres
    for (int i = 0; i < TOTAL_BLOCKS; i++) {
        free_blocks[i] = true; // 'true' significa que el bloque está libre
    }

    num_files = 0; // No hay archivos en el sistema al inicio
    printf("Sistema de archivos inicializado. Espacio total: %d MB, Bloques: %d, Tamano de Bloque: %d bytes.\n",
           DISK_SIZE_MB, TOTAL_BLOCKS, BLOCK_SIZE);
}

/**
 * @brief Busca una entrada de archivo en la tabla de archivos por nombre.
 * @param name El nombre del archivo a buscar.
 * @return El índice de la entrada del archivo en la tabla de archivos, o -1 si no se encuentra.
 */
int find_file_entry_index(const char* name) {
    for (int i = 0; i < MAX_FILES; i++) {
        // Solo compara si la entrada no está libre y los nombres coinciden
        if (file_table[i].start_block != -1 && strcmp(file_table[i].name, name) == 0) {
            return i;
        }
    }
    return -1; // Archivo no encontrado
}

/**
 * @brief Asigna un número específico de bloques contiguos en el disco.
 * Implementa una estrategia de "primer ajuste" para encontrar bloques contiguos.
 * @param required_blocks El número de bloques que se necesitan asignar.
 * @return El índice del primer bloque asignado, o -1 si no se pueden encontrar bloques contiguos suficientes.
 */
int allocate_blocks(int required_blocks) {
    if (required_blocks <= 0) {
        return -1; // No se pueden asignar 0 o menos bloques
    }

    int free_count = 0; // Contador de bloques libres contiguos
    for (int i = 0; i < TOTAL_BLOCKS; i++) {
        if (free_blocks[i]) {
            free_count++;
            if (free_count == required_blocks) {
                // Se encontraron suficientes bloques contiguos
                int start_index = i - required_blocks + 1;
                // Marcar los bloques como ocupados
                for (int j = 0; j < required_blocks; j++) {
                    free_blocks[start_index + j] = false;
                }
                return start_index; // Retornar el índice del primer bloque asignado
            }
        } else {
            free_count = 0; // Reiniciar el contador si se encuentra un bloque ocupado
        }
    }
    return -1; // No se encontraron suficientes bloques contiguos
}

/**
 * @brief Libera los bloques de disco previamente asignados a un archivo.
 * Marca los bloques como libres en el mapa de bits de bloques libres.
 * @param start_block El índice del primer bloque a liberar.
 * @param num_blocks El número de bloques a liberar.
 */
void free_blocks_for_file(int start_block, int num_blocks) {
    if (start_block == -1 || num_blocks <= 0) {
        return; // No hay bloques que liberar o el número de bloques es inválido
    }

    for (int i = 0; i < num_blocks; i++) {
        int block_index = start_block + i;
        if (block_index >= 0 && block_index < TOTAL_BLOCKS) {
            free_blocks[block_index] = true; // Marcar el bloque como libre
        }
    }
}

/**
 * @brief Crea un nuevo archivo en el sistema de archivos.
 * Asigna una entrada en la tabla de archivos y bloques en el disco.
 * @param name El nombre del archivo a crear.
 * @param size El tamaño del archivo en bytes.
 */
void create_file(const char* name, int size) {
    // Validar el nombre del archivo
    if (strlen(name) >= MAX_FILENAME_LEN) { // MAX_FILENAME_LEN incluye el nulo, así que >= MAX_FILENAME_LEN significa desborde
        printf("Error: Nombre de archivo demasiado largo (max %d caracteres).\n", MAX_FILENAME_LEN - 1);
        return;
    }
    // Validar el tamaño del archivo
    if (size <= 0) {
        printf("Error: El tamano del archivo debe ser mayor a 0.\n");
        return;
    }
    // Verificar si el archivo ya existe
    if (find_file_entry_index(name) != -1) {
        printf("Error: El archivo '%s' ya existe.\n", name);
        return;
    }
    // Verificar si hay espacio para nuevos archivos en la tabla de archivos
    if (num_files >= MAX_FILES) {
        printf("Error: No se puede crear el archivo. Limite maximo de archivos (%d) alcanzado.\n", MAX_FILES);
        return;
    }

    // Calcular el número de bloques necesarios
    int required_blocks = (size + BLOCK_SIZE - 1) / BLOCK_SIZE; // Redondeo hacia arriba

    // Verificar el tamaño máximo permitido para un archivo (ej. 1MB por archivo como parte de la restricción general)
    // Aunque el límite general es 1MB, podemos interpretar que cada archivo no debe sobrepasar ese límite.
    if (size > (1024 * 1024)) { // 1 MB
        printf("Error: Tamano de archivo excede el limite maximo permitido (1 MB).\n");
        return;
    }

    // Intentar asignar los bloques en el disco
    int start_block = allocate_blocks(required_blocks);
    if (start_block == -1) {
        printf("Error: No hay suficiente espacio contiguo en disco para crear el archivo '%s' de %d bytes.\n", name, size);
        return;
    }

    // Encontrar una entrada libre en la tabla de archivos
    int file_entry_index = -1;
    for (int i = 0; i < MAX_FILES; i++) {
        if (file_table[i].start_block == -1) { // Entrada libre
            file_entry_index = i;
            break;
        }
    }

    // Si se encontró una entrada libre y se asignaron bloques, crear el archivo
    if (file_entry_index != -1) {
        strcpy(file_table[file_entry_index].name, name);
        file_table[file_entry_index].size = size;
        file_table[file_entry_index].start_block = start_block;
        file_table[file_entry_index].num_blocks = required_blocks;
        num_files++;
        printf("Archivo '%s' creado con %d bytes (%d bloques, inicio en bloque %d).\n",
               name, size, required_blocks, start_block);
    } else {
        // Esto no debería suceder si num_files < MAX_FILES, pero es un resguardo
        printf("Error interno: No se pudo encontrar una entrada de archivo libre.\n");
        free_blocks_for_file(start_block, required_blocks); // Liberar bloques si no se pudo crear la entrada
    }
}

/**
 * @brief Escribe datos en un archivo existente.
 * @param name El nombre del archivo.
 * @param offset La posición (offset) en bytes donde comenzar a escribir.
 * @param data Los datos a escribir.
 */
void write_file(const char* name, int offset, const char* data) {
    int file_entry_index = find_file_entry_index(name);
    if (file_entry_index == -1) {
        printf("Error: Archivo '%s' no encontrado.\n", name);
        return;
    }

    FileEntry* file = &file_table[file_entry_index];
    int data_len = strlen(data);

    // Validar offset y tamaño de los datos
    if (offset < 0 || offset > file->size) {
        printf("Error: Offset de escritura invalido (%d) para el archivo '%s' (tamano %d).\n", offset, name, file->size);
        return;
    }

    // Calcular la cantidad de bytes que se escribirán (limitado por el tamaño del archivo y el offset)
    int bytes_to_write = data_len;
    if (offset + bytes_to_write > file->size) {
        // Si la escritura excede el límite del archivo, solo escribir hasta el final
        bytes_to_write = file->size - offset;
        printf("Advertencia: Se intentó escribir mas allá del final del archivo. Solo se escribiran %d bytes.\n", bytes_to_write);
    }

    if (bytes_to_write <= 0) {
        printf("No hay datos validos para escribir o el offset ya esta al final del archivo.\n");
        return;
    }

    // Recorrer los datos y escribir en los bloques correspondientes
    int current_data_offset = 0; // Offset dentro de los datos a escribir
    while (current_data_offset < bytes_to_write) {
        // Calcular el bloque de disco y el offset dentro de ese bloque
        int absolute_offset = offset + current_data_offset;
        int block_num_in_file = absolute_offset / BLOCK_SIZE; // Bloque relativo al inicio del archivo
        int block_offset = absolute_offset % BLOCK_SIZE;       // Offset dentro de ese bloque

        // Calcular el índice real del bloque en el disco
        int actual_disk_block_index = file->start_block + block_num_in_file;

        // Calcular cuántos bytes se pueden escribir en el bloque actual
        int bytes_in_current_block = BLOCK_SIZE - block_offset;
        int bytes_to_copy = (bytes_to_write - current_data_offset < bytes_in_current_block) ?
                             (bytes_to_write - current_data_offset) : bytes_in_current_block;

        // Copiar los datos al bloque de disco simulado
        memcpy(disk[actual_disk_block_index] + block_offset,
               data + current_data_offset,
               bytes_to_copy);

        current_data_offset += bytes_to_copy;
    }
    printf("Escritos %d bytes en el archivo '%s' desde el offset %d.\n", bytes_to_write, name, offset);
}


/**
 * @brief Lee una cantidad específica de bytes desde un archivo.
 * @param name El nombre del archivo.
 * @param offset La posición (offset) en bytes desde donde comenzar a leer.
 * @param size_to_read La cantidad de bytes a leer.
 */
void read_file(const char* name, int offset, int size_to_read) {
    int file_entry_index = find_file_entry_index(name);
    if (file_entry_index == -1) {
        printf("Error: Archivo '%s' no encontrado.\n", name);
        return;
    }

    FileEntry* file = &file_table[file_entry_index];

    // Validar offset y tamaño de la lectura
    if (offset < 0 || offset >= file->size) {
        printf("Error: Offset de lectura invalido (%d) para el archivo '%s' (tamano %d).\n", offset, name, file->size);
        return;
    }
    if (size_to_read <= 0) {
        printf("Error: Tamano de lectura debe ser mayor a 0.\n");
        return;
    }

    // Ajustar size_to_read si excede los límites del archivo
    if (offset + size_to_read > file->size) {
        size_to_read = file->size - offset;
        printf("Advertencia: Se intento leer mas alla del final del archivo. Se leyeron %d bytes.\n", size_to_read);
    }

    if (size_to_read == 0) {
        printf("No hay bytes para leer desde el offset %d.\n", offset);
        return;
    }

    // Usar un buffer dinámico para almacenar los datos leídos
    char* read_buffer = (char*)malloc(size_to_read + 1); // +1 para el terminador nulo
    if (read_buffer == NULL) {
        printf("Error: No se pudo asignar memoria para el buffer de lectura.\n");
        return;
    }

    int current_read_offset = 0; // Offset dentro de los datos que ya se han leído
    while (current_read_offset < size_to_read) {
        // Calcular el bloque de disco y el offset dentro de ese bloque
        int absolute_offset = offset + current_read_offset;
        int block_num_in_file = absolute_offset / BLOCK_SIZE;
        int block_offset = absolute_offset % BLOCK_SIZE;

        // Calcular el índice real del bloque en el disco
        int actual_disk_block_index = file->start_block + block_num_in_file;

        // Calcular cuántos bytes se pueden leer del bloque actual
        int bytes_in_current_block = BLOCK_SIZE - block_offset;
        int bytes_to_copy = (size_to_read - current_read_offset < bytes_in_current_block) ?
                             (size_to_read - current_read_offset) : bytes_in_current_block;

        // Copiar los datos del bloque de disco simulado al buffer de lectura
        memcpy(read_buffer + current_read_offset,
               disk[actual_disk_block_index] + block_offset,
               bytes_to_copy);

        current_read_offset += bytes_to_copy;
    }

    read_buffer[size_to_read] = '\0'; // Asegurar que la cadena esté terminada en nulo
    printf("Salida: \"%s\"\n", read_buffer);

    free(read_buffer); // Liberar el buffer dinámico
}

/**
 * @brief Elimina un archivo del sistema de archivos.
 * Libera la entrada en la tabla de archivos y los bloques de disco asociados.
 * @param name El nombre del archivo a eliminar.
 */
void delete_file(const char* name) {
    int file_entry_index = find_file_entry_index(name);
    if (file_entry_index == -1) {
        printf("Error: Archivo '%s' no encontrado.\n", name);
        return;
    }

    FileEntry* file_to_delete = &file_table[file_entry_index];

    // Liberar los bloques de disco ocupados por el archivo
    free_blocks_for_file(file_to_delete->start_block, file_to_delete->num_blocks);

    // Marcar la entrada de la tabla de archivos como libre
    file_to_delete->start_block = -1; // Esto indica que la entrada está libre
    file_to_delete->size = 0;
    file_to_delete->num_blocks = 0;
    memset(file_to_delete->name, 0, MAX_FILENAME_LEN); // Limpiar el nombre

    num_files--;
    printf("Archivo '%s' eliminado. Bloques liberados.\n", name);
}

/**
 * @brief Lista todos los archivos actualmente almacenados en el sistema de archivos.
 * Muestra el nombre y el tamaño de cada archivo.
 */
void list_files() {
    printf("Archivos en el sistema:\n");
    bool found_files = false;
    for (int i = 0; i < MAX_FILES; i++) {
        if (file_table[i].start_block != -1) { // Si la entrada no está libre
            printf("%s %d bytes\n", file_table[i].name, file_table[i].size);
            found_files = true;
        }
    }
    if (!found_files) {
        printf("(no hay archivos)\n");
    }
}

/**
 * @brief Muestra la ayuda de comandos disponibles.
 */
void display_help() {
    printf("Comandos disponibles:\n");
    printf("  CREATE <nombre_archivo> <tamano_bytes> - Crea un archivo.\n");
    printf("  WRITE <nombre_archivo> <offset> \"<datos>\" - Escribe datos en un archivo.\n");
    printf("  READ <nombre_archivo> <offset> <tamano> - Lee datos de un archivo.\n");
    printf("  DELETE <nombre_archivo> - Elimina un archivo.\n");
    printf("  LIST - Lista todos los archivos.\n");
    printf("  HELP - Muestra esta ayuda.\n");
    printf("  EXIT - Sale del programa.\n");
}

/**
 * @brief Analiza y ejecuta un comando ingresado por el usuario.
 * @param command_line La línea de comando completa ingresada por el usuario.
 */
void parse_command(char* command_line) {
    char *token;
    char *rest = command_line;
    char cmd[MAX_FILENAME_LEN]; // Buffer para el nombre del comando
    // Leer el primer token (el comando)
    token = strtok_r(rest, " ", &rest);
    if (token == NULL) {
        return; // Línea vacía
    }

    // Copiar el comando y convertir a mayúsculas
    // strncpy con MAX_FILENAME_LEN - 1 asegura espacio para el nulo terminador
    strncpy(cmd, token, MAX_FILENAME_LEN - 1);

    cmd[MAX_FILENAME_LEN - 1] = '\0'; // Asegurar terminación nula
    for (int i = 0; cmd[i]; i++) {
        cmd[i] = toupper(cmd[i]);
    }
    

    // Buffer para nombres de archivo que se pasan a las funciones de archivo
    char filename_buffer[MAX_FILENAME_LEN];

    if (strcmp(cmd, "CREATE") == 0) {
        int size;
        token = strtok_r(rest, " ", &rest); // nombre_archivo
        if (token == NULL) { printf("Uso: CREATE <nombre_archivo> <tamano_bytes>\n"); return; }
        
        // ** VALIDACIÓN CRÍTICA: Prevenir desbordamiento del búfer antes de strcpy **
        if (strlen(token) >= MAX_FILENAME_LEN) {
            printf("Error: Nombre de archivo '%s' demasiado largo (max %d caracteres).\n", token, MAX_FILENAME_LEN - 1);
            return;
        }
        strcpy(filename_buffer, token); // Copiar al búfer seguro

        token = strtok_r(rest, " ", &rest); // tamano_bytes
        if (token == NULL) { printf("Uso: CREATE <nombre_archivo> <tamano_bytes>\n"); return; }
        size = atoi(token);
        printf(" Parse 6 (Create file)\n");
        create_file(filename_buffer, size); // Usar el búfer seguro

    } else if (strcmp(cmd, "WRITE") == 0) {
        int offset;
        char data[MAX_WRITE_DATA_LEN + 1]; // +1 para el terminador nulo

        token = strtok_r(rest, " ", &rest); // nombre_archivo
        if (token == NULL) { printf("Uso: WRITE <nombre_archivo> <offset> \"<datos>\"\n"); return; }
        
        // ** VALIDACIÓN CRÍTICA: Prevenir desbordamiento del búfer antes de strcpy **
        if (strlen(token) >= MAX_FILENAME_LEN) {
            printf("Error: Nombre de archivo '%s' demasiado largo (max %d caracteres).\n", token, MAX_FILENAME_LEN - 1);
            return;
        }
        strcpy(filename_buffer, token); // Copiar al búfer seguro

        token = strtok_r(rest, " ", &rest); // offset
        if (token == NULL) { printf("Uso: WRITE <nombre_archivo> <offset> \"<datos>\"\n"); return; }
        offset = atoi(token);

        // Los datos están entre comillas dobles, necesitamos encontrar el inicio y el fin
        char* data_start = strchr(rest, '\"');
        if (data_start == NULL) {
            printf("Uso: WRITE <nombre_archivo> <offset> \"<datos>\" (los datos deben estar entre comillas dobles).\n");
            return;
        }
        data_start++; // Moverse más allá de la primera comilla

        char* data_end = strchr(data_start, '\"');
        if (data_end == NULL) {
            printf("Error: Las comillas dobles de cierre no fueron encontradas.\n");
            return;
        }

        int data_len = data_end - data_start;
        // Si data_len es igual o mayor a MAX_WRITE_DATA_LEN, significa que no cabe el nulo terminador
        if (data_len >= MAX_WRITE_DATA_LEN) { 
            printf("Error: Los datos a escribir son demasiado largos (max %d caracteres).\n", MAX_WRITE_DATA_LEN - 1);
            return;
        }
        strncpy(data, data_start, data_len);
        data[data_len] = '\0'; // Asegurar terminación nula

        write_file(filename_buffer, offset, data); // Usar el búfer seguro
    } else if (strcmp(cmd, "READ") == 0) {
        int offset, size_to_read;

        token = strtok_r(rest, " ", &rest); // nombre_archivo
        if (token == NULL) { printf("Uso: READ <nombre_archivo> <offset> <tamano>\n"); return; }
        
        // ** VALIDACIÓN CRÍTICA: Prevenir desbordamiento del búfer antes de strcpy **
        if (strlen(token) >= MAX_FILENAME_LEN) {
            printf("Error: Nombre de archivo '%s' demasiado largo (max %d caracteres).\n", token, MAX_FILENAME_LEN - 1);
            return;
        }
        strcpy(filename_buffer, token); // Copiar al búfer seguro

        token = strtok_r(rest, " ", &rest); // offset
        if (token == NULL) { printf("Uso: READ <nombre_archivo> <offset> <tamano>\n"); return; }
        offset = atoi(token);

        token = strtok_r(rest, " ", &rest); // tamano
        if (token == NULL) { printf("Uso: READ <nombre_archivo> <offset> <tamano>\n"); return; }
        size_to_read = atoi(token);

        read_file(filename_buffer, offset, size_to_read); // Usar el búfer seguro
    } else if (strcmp(cmd, "DELETE") == 0) {
        token = strtok_r(rest, " ", &rest); // nombre_archivo
        if (token == NULL) { printf("Uso: DELETE <nombre_archivo>\n"); return; }
        
        // ** VALIDACIÓN CRÍTICA: Prevenir desbordamiento del búfer antes de strcpy **
        if (strlen(token) >= MAX_FILENAME_LEN) {
            printf("Error: Nombre de archivo '%s' demasiado largo (max %d caracteres).\n", token, MAX_FILENAME_LEN - 1);
            return;
        }
        strcpy(filename_buffer, token); // Copiar al búfer seguro
        delete_file(filename_buffer); // Usar el búfer seguro
    } else if (strcmp(cmd, "LIST") == 0) {
        list_files();
    } else if (strcmp(cmd, "HELP") == 0) {
        display_help();
    } else if (strcmp(cmd, "EXIT") == 0) {
        printf("Saliendo del sistema de archivos.\n");
        exit(0);
    } else {
        printf("Comando desconocido. Escriba 'HELP' para ver los comandos disponibles.\n");
    }
}

/**
 * @brief Función principal del programa.
 * Inicializa el sistema de archivos y entra en un bucle de comandos interactivo.
 */
int main() {
    init_file_system();

    char command_line[256]; // Buffer para la línea de comando

    while (1) {
        printf("\n> "); // Prompt del sistema de archivos
        if (fgets(command_line, sizeof(command_line), stdin) == NULL) {
            // En caso de error o EOF (Ctrl+D), salir del bucle
            break;
        }
        // Eliminar el salto de línea al final si existe
        command_line[strcspn(command_line, "\n")] = 0;

        // Si la línea está vacía, continuar
        if (strlen(command_line) == 0) {
            continue;
        }

        parse_command(command_line);
    }

    return 0;
}
