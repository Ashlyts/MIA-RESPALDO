# Manual de Usuario: GoDisk/ExtreamFS
+ Nombre: Ashley Mishell Tubac Sitavi
+ Curso: manejo e implementación de archivos
+ Auxiliar: Juan Francisco Silva Uriba

# MANUAL TÉCNICO



## 1. Introducción
El sistema **GoDisk** es una aplicación desarrollada en el lenguaje de programación **Go**, diseñada para simular la gestión de un sistema de archivos tipo Linux. El sistema permite la creación de discos virtuales, particionamiento y una administración completa de usuarios y grupos a través de una consola de comandos.


## 2. Arquitectura del Sistema

El software sigue una **arquitectura modular**, donde cada paquete tiene una responsabilidad única, facilitando el mantenimiento y la escalabilidad.

### 2.1 Módulos Principales

* **Package `general`**
  Contiene el motor de ejecución (*dispatcher*). Se encarga de recibir las instrucciones, limpiar comentarios, parsear parámetros mediante expresiones regulares y dirigir el flujo hacia el comando correspondiente.

* **Package `filecomands`**
  Implementa la lógica de bajo nivel para la creación de usuarios (mkusr), grupos (mkgrp), carpetas (mkdir) y archivos (mkfile).

* **Package `global`**
  Gestiona el estado de la aplicación, específicamente la sesión activa del usuario, permitiendo un control de acceso basado en roles.

* **Package `utils`**
  Proveedor de funciones auxiliares para lectura y escritura binaria en el archivo `.dsk`.



## 3. Administración de Usuarios y Grupos

La seguridad y organización del sistema se basan en el archivo central `/users.txt`.

### 3.1 Estructura del Archivo `users.txt`

A diferencia de un sistema real, este sistema utiliza un formato **CSV** almacenado dentro de los bloques de datos de la partición.

**Grupos:**


GID, G, Nombre_Grupo

**Usuarios:**

```
UID, U, Grupo, Usuario, Password
```

### 3.2 Lógica de Comandos

#### MKGRP (Make Group)

* Verifica que el usuario logueado sea `root`.
* Lee el archivo `users.txt`.
* Busca el **GID** más alto para autoincrementarlo.
* Valida que el nombre no exceda los **10 caracteres**.

#### MKUSR (Make User)

* Valida la existencia del grupo asignado.
* Verifica que el nombre de usuario sea único.
* Genera un **UID** correlativo.



## 4. Motor de Comandos (Parsing)

El sistema utiliza un procesador de cadenas robusto para interpretar las entradas del usuario.

### 4.1 Expresiones Regulares

Para garantizar que parámetros como rutas con espacios o nombres complejos sean capturados correctamente, se utiliza la siguiente expresión regular:

```go
re := regexp.MustCompile("-([a-zA-Z]\\w*)=(\"[^\"]*\"|\\S+)")
```

Esta expresión separa las banderas de sus valores.

### 4.2 Flujo de Ejecución del Dispatcher

1. **Recepción:** El comando entra como una lista de strings.
2. **Identificación:** Se determina a qué grupo pertenece (Disk, Users, File).
3. **Ejecución:** Se invoca la función Execute correspondiente pasando un map[string]string con los parámetros procesados.



## 5. Estructuras de Datos de Sesión

Para mantener la persistencia del login entre comandos, se utiliza una estructura global de sesión:

```go
type SesionUsuario struct {
    UsuarioActual string                // Nombre del usuario 
    UID           int32                 // ID único del usuario
    GID           int32                 // ID del grupo al que 
    IDParticion   string                // ID de la partición 
    PathDisco     string                // Ruta física al 
    Particion     *structures.Partition // Datos de la partición 
}
```


## 6. Manejo de Almacenamiento (I/O)

El sistema opera directamente sobre archivos binarios. Cada operación de escritura sigue el siguiente protocolo:

1. **Lectura del Superbloque:** Para conocer el estado de los bitmaps y las tablas de inodos.
2. **Cálculo de Offsets:** Se utiliza la posición Part_start para no corromper otras particiones del disco.
3. **Serialización:** Los datos se escriben utilizando binary. Write para asegurar que el tamaño de las estructuras en el disco sea constante.



## 7. Limitaciones Técnicas

* **Longitud de Cadenas:** Nombres de usuario, contraseñas y grupos están limitados a **10 caracteres** por compatibilidad con el sistema de archivos.
* **Acceso:** Solo se permite una sesión activa por ejecución del programa.
* **Formato:** El comando mkfs es requisito indispensable antes de cualquier operación de usuarios en una partición nueva.
