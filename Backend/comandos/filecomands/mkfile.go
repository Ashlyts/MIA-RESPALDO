package filecomands

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// MkfileExecute maneja el comando mkfile
func MkfileExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar sesión activa
	if global.SesionActiva == nil {
		return "[MKFILE]: No hay sesión activa. Use LOGIN primero", true
	}

	// Validar parámetros obligatorios
	path := strings.TrimSpace(parametros["path"])
	if path == "" {
		return "[MKFILE]: Parámetro -path es obligatorio", true
	}

	// Parámetros opcionales
	content := strings.TrimSpace(parametros["cont"])
	sizeParam := strings.TrimSpace(parametros["size"])

	var sizeValue int32 = 0
	var sizeProvided bool = false
	if sizeParam != "" {
		// Intentar convertir el tamaño a int32
		sizeInt, err := strconv.Atoi(sizeParam)
		if err != nil {
			return "[MKFILE]: Parámetro -size debe ser un número entero", true
		}
		if sizeInt < 0 {
			return "[MKFILE]: Parámetro -size no puede ser negativo", true
		}
		sizeValue = int32(sizeInt)
		sizeProvided = true
	}

	// Verificar que no se proporcionen cont y size simultáneamente
	if content != "" && sizeProvided {
		return "[MKFILE]: No se puede especificar -cont y -size al mismo tiempo", true
	}

	return crearArchivo(path, content, sizeValue, sizeProvided)
}

func crearArchivo(path string, content string, sizeValue int32, sizeProvided bool) (string, bool) {
	// Abrir el disco
	file, err := os.OpenFile(global.SesionActiva.PathDisco, os.O_RDWR, 0666)
	if err != nil {
		return "[MKFILE]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer SuperBloque
	sb, errSB := utils.LeerSuperBloque(file, global.SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[MKFILE]: Error al leer SuperBloque", true
	}

	// Dividir la ruta para obtener el nombre del archivo y la ruta del directorio padre
	ruta := strings.TrimSpace(path)
	if !strings.HasPrefix(ruta, "/") {
		ruta = "/" + ruta
	}
	partes := strings.Split(strings.Trim(ruta, "/"), "/")
	if len(partes) == 0 || (len(partes) == 1 && partes[0] == "") {
		return "[MKFILE]: Ruta inválida para archivo", true
	}

	nombreArchivo := partes[len(partes)-1]
	rutaDirectorioPadre := "/" + strings.Join(partes[:len(partes)-1], "/")

	if rutaDirectorioPadre == "/" {
		rutaDirectorioPadre = "/" // Asegurar que sea la raíz
	}

	// Obtener el inodo del directorio padre
	inodoPadre, _, errDir := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre)
	if errDir != nil {
		return fmt.Sprintf("[MKFILE]: Error al acceder al directorio padre '%s': %v", rutaDirectorioPadre, errDir), true
	}

	// Verificar permisos de escritura en el directorio padre
	if !utils.TienePermisoEscritura(&inodoPadre, global.SesionActiva, "") { // Nombre del archivo no es relevante para permiso de escritura en el padre
		return fmt.Sprintf("[MKFILE]: No tiene permisos de escritura en el directorio '%s'", rutaDirectorioPadre), true
	}

	// Verificar si el archivo ya existe en el directorio padre
	_, existe, errBusqueda := utils.BuscarEnCarpeta(file, &sb, &inodoPadre, nombreArchivo)
	if errBusqueda != nil {
		return fmt.Sprintf("[MKFILE]: Error buscando archivo en directorio '%s': %v", rutaDirectorioPadre, errBusqueda), true
	}
	if existe {
		return fmt.Sprintf("[MKFILE]: El archivo '%s' ya existe en '%s'", nombreArchivo, rutaDirectorioPadre), true
	}

	// Determinar el contenido del archivo
	var contenidoFinal string
	if content != "" {
		contenidoFinal = content
	} else if sizeProvided {
		// Generar contenido basado en size
		contenidoFinal = generarContenido(sizeValue)
	} else {
		// Si no se proporciona ni cont ni size, crear archivo vacío
		contenidoFinal = ""
	}

	// Crear el nuevo archivo
	errCrear := utils.CrearArchivo(file, &sb, &inodoPadre, nombreArchivo, contenidoFinal)
	if errCrear != nil {
		return fmt.Sprintf("[MKFILE]: Error al crear archivo '%s': %v", nombreArchivo, errCrear), true
	}

	// Actualizar SuperBloque (ya se actualiza dentro de CrearArchivo)
	// Escribir SuperBloque actualizado
	if _, err := file.Seek(int64(global.SesionActiva.Particion.Part_start), 0); err != nil {
		return "[MKFILE]: Error al posicionar puntero para escribir SuperBloque", true
	}
	if err := binary.Write(file, binary.LittleEndian, &sb); err != nil {
		return "[MKFILE]: Error al escribir SuperBloque actualizado", true
	}

	color.Green("===========================================================")
	color.Green(" ARCHIVO CREADO EXITOSAMENTE")
	color.Green("===========================================================")
	color.Cyan("  Ruta:           %s", ruta)
	color.Cyan("  Tamaño:         %d bytes", len(contenidoFinal))
	color.Green("===========================================================")

	return "", false
}

// generarContenido genera contenido basado en el tamaño especificado
func generarContenido(tamano int32) string {
	if tamano == 0 {
		return ""
	}
	var contenido strings.Builder
	contador := 0
	for i := int32(0); i < tamano; i++ {
		contenido.WriteString(fmt.Sprintf("%d", contador))
		contador++
		if contador == 10 {
			contador = 0
		}
	}
	return contenido.String()
}
