package admonUsers

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Variable global para almacenar la sesión activa
var SesionActiva *SesionUsuario

type SesionUsuario struct {
	UsuarioActual string
	UID           int32
	GID           int32
	IDParticion   string
	PathDisco     string
	Particion     *structures.Partition
}

// catExecute maneja el comando cat
func CatExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que haya sesión activa
	if SesionActiva == nil {
		return "[CAT]: No hay sesión activa. Use el comando LOGIN", true
	}

	// Obtener archivos a mostrar (file1, file2, file3, ...)
	var archivos []string
	for i := 1; i <= 10; i++ { // Soportar hasta 10 archivos
		key := fmt.Sprintf("file%d", i)
		if archivo, existe := parametros[key]; existe && archivo != "" {
			archivos = append(archivos, strings.TrimSpace(archivo))
		}
	}

	if len(archivos) == 0 {
		return "[CAT]: Debe especificar al menos un archivo con -file1=ruta", true
	}

	return mostrarContenidoArchivos(archivos)
}

// mostrarContenidoArchivos lee y muestra el contenido de los archivos especificados
func mostrarContenidoArchivos(rutas []string) (string, bool) {
	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("                    CONTENIDO DE ARCHIVO(S)")
	color.Green("═══════════════════════════════════════════════════════════\n")

	// Abrir el disco
	file, err := os.OpenFile(SesionActiva.PathDisco, os.O_RDONLY, 0666)
	if err != nil {
		return "[CAT]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer el SuperBloque
	sb, errSB := leerSuperBloque(file, SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[CAT]: Error al leer SuperBloque: " + errSB.Error(), true
	}

	// Mostrar contenido de cada archivo
	for idx, ruta := range rutas {
		if idx > 0 {
			fmt.Println() // Separador entre archivos
		}

		color.Cyan("─────────────────────────────────────────────────────────")
		color.Yellow("Archivo: %s", ruta)
		color.Cyan("─────────────────────────────────────────────────────────")

		contenido, errCat := leerArchivoDesdeRuta(file, &sb, ruta)
		if errCat != nil {
			color.Red("✗ Error: %s", errCat.Error())
			continue
		}

		// Mostrar el contenido
		if contenido == "" {
			color.White("(archivo vacío)")
		} else {
			color.White("%s", contenido)
		}
	}

	color.Green("\n═══════════════════════════════════════════════════════════")
	return "", false
}

// leerSuperBloque lee el SuperBloque de la partición
func leerSuperBloque(file *os.File, inicioParticion int32) (structures.SuperBloque, error) {
	var sb structures.SuperBloque

	if _, err := file.Seek(int64(inicioParticion), 0); err != nil {
		return sb, err
	}

	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		return sb, err
	}

	// Verificar que esté formateado
	if sb.S_magic != 0xEF53 {
		return sb, fmt.Errorf("partición no formateada")
	}

	return sb, nil
}

// leerArchivoDesdeRuta navega por la estructura de directorios y lee el archivo
func leerArchivoDesdeRuta(file *os.File, sb *structures.SuperBloque, rutaCompleta string) (string, error) {
	// Limpiar la ruta
	ruta := strings.TrimSpace(rutaCompleta)
	if ruta == "" {
		return "", fmt.Errorf("ruta vacía")
	}

	// Asegurar que empiece con /
	if !strings.HasPrefix(ruta, "/") {
		ruta = "/" + ruta
	}

	// Dividir la ruta en partes
	partes := strings.Split(strings.Trim(ruta, "/"), "/")
	if len(partes) == 0 || (len(partes) == 1 && partes[0] == "") {
		return "", fmt.Errorf("ruta inválida")
	}

	// Empezar desde el inodo raíz (inodo 0)
	inodoActual, err := leerInodo(file, sb, 0)
	if err != nil {
		return "", fmt.Errorf("error al leer inodo raíz: %v", err)
	}

	// Navegar por cada parte de la ruta
	for i, parte := range partes {
		if parte == "" {
			continue
		}

		esUltimo := (i == len(partes)-1)

		// Verificar que el inodo actual sea una carpeta
		if inodoActual.I_type[0] != '0' {
			return "", fmt.Errorf("'%s' no es una carpeta", strings.Join(partes[:i], "/"))
		}

		// Buscar la parte en la carpeta actual
		siguienteInodo, encontrado, err := buscarEnCarpeta(file, sb, &inodoActual, parte)
		if err != nil {
			return "", err
		}

		if !encontrado {
			return "", fmt.Errorf("'%s' no encontrado", parte)
		}

		// Leer el siguiente inodo
		inodoActual, err = leerInodoPorPosicion(file, siguienteInodo)
		if err != nil {
			return "", fmt.Errorf("error al leer inodo: %v", err)
		}

		// Si es el último elemento, verificar permisos y leer contenido
		if esUltimo {
			// Verificar que sea un archivo
			if inodoActual.I_type[0] != '1' {
				return "", fmt.Errorf("'%s' es una carpeta, no un archivo", parte)
			}

			// Verificar permisos de lectura
			if !tienePermisoLectura(&inodoActual) {
				return "", fmt.Errorf("sin permisos de lectura para '%s'", parte)
			}

			// Leer el contenido del archivo
			return leerContenidoArchivo(file, sb, &inodoActual)
		}
	}

	return "", fmt.Errorf("ruta inválida")
}

// buscarEnCarpeta busca un nombre en una carpeta y retorna el inodo
func buscarEnCarpeta(file *os.File, sb *structures.SuperBloque, inodoCarpeta *structures.TablaInodo, nombre string) (int32, bool, error) {
	// Recorrer los bloques directos de la carpeta
	for i := 0; i < 12; i++ {
		if inodoCarpeta.I_block[i] == -1 {
			break
		}

		// Leer el bloque carpeta
		var bloqueCarpeta structures.BloqueCarpeta
		posicionBloque := inodoCarpeta.I_block[i]

		if _, err := file.Seek(int64(posicionBloque), 0); err != nil {
			return -1, false, err
		}

		if err := binary.Read(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
			return -1, false, err
		}

		// Buscar en las entradas del bloque
		for _, entrada := range bloqueCarpeta.B_content {
			if entrada.B_inodo == -1 {
				continue
			}

			nombreEntrada := strings.TrimRight(string(entrada.B_name[:]), "\x00")
			if nombreEntrada == nombre {
				return entrada.B_inodo, true, nil
			}
		}
	}

	return -1, false, nil
}

// leerInodo lee un inodo por su índice
func leerInodo(file *os.File, sb *structures.SuperBloque, indice int32) (structures.TablaInodo, error) {
	var inodo structures.TablaInodo

	posicion := sb.S_inode_start + (indice * size.SizeTablaInodo())

	if _, err := file.Seek(int64(posicion), 0); err != nil {
		return inodo, err
	}

	if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
		return inodo, err
	}

	return inodo, nil
}

// leerInodoPorPosicion lee un inodo directamente por su posición en bytes
func leerInodoPorPosicion(file *os.File, posicion int32) (structures.TablaInodo, error) {
	var inodo structures.TablaInodo

	if _, err := file.Seek(int64(posicion), 0); err != nil {
		return inodo, err
	}

	if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
		return inodo, err
	}

	return inodo, nil
}

// tienePermisoLectura verifica si el usuario tiene permiso de lectura
func tienePermisoLectura(inodo *structures.TablaInodo) bool {
	// Si es root (UID=1), tiene todos los permisos
	if SesionActiva.UID == 1 {
		return true
	}

	// Obtener permisos
	permisoUser := inodo.I_perm[0]  // Propietario
	permisoGroup := inodo.I_perm[1] // Grupo
	permisoOther := inodo.I_perm[2] // Otros

	// Verificar según la categoría del usuario
	if inodo.I_uid == SesionActiva.UID {
		// Es el propietario - verificar permiso de lectura (r = 4 o mayor)
		return permisoUser >= '4'
	} else if inodo.I_gid == SesionActiva.GID {
		// Es del mismo grupo
		return permisoGroup >= '4'
	} else {
		// Es otro usuario
		return permisoOther >= '4'
	}
}

// leerContenidoArchivo lee el contenido completo de un archivo
func leerContenidoArchivo(file *os.File, sb *structures.SuperBloque, inodo *structures.TablaInodo) (string, error) {
	var contenidoTotal strings.Builder

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		if inodo.I_block[i] == -1 {
			break
		}

		var bloqueArchivo structures.BloqueArchivo
		if _, err := file.Seek(int64(inodo.I_block[i]), 0); err != nil {
			return "", err
		}

		if err := binary.Read(file, binary.LittleEndian, &bloqueArchivo); err != nil {
			return "", err
		}

		// Extraer contenido (hasta encontrar null byte o fin del bloque)
		contenido := string(bloqueArchivo.B_content[:])
		contenido = strings.TrimRight(contenido, "\x00")
		contenidoTotal.WriteString(contenido)
	}

	return contenidoTotal.String(), nil
}
