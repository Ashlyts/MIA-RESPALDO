// utils/utils_users.go
package utils

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/global"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func LeerSuperBloque(file *os.File, inicioParticion int32) (structures.SuperBloque, error) {
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

// LeerArchivoDesdeRuta navega por la estructura de directorios y lee el archivo.
func LeerArchivoDesdeRuta(file *os.File, sb *structures.SuperBloque, rutaCompleta string) (string, error) {
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
	inodoActual, err := LeerInodo(file, sb, 0)
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
		siguienteInodo, encontrado, err := BuscarEnCarpeta(file, sb, &inodoActual, parte)
		if err != nil {
			return "", err
		}

		if !encontrado {
			return "", fmt.Errorf("'%s' no encontrado", parte)
		}

		// Leer el siguiente inodo
		inodoActual, err = LeerInodoPorPosicion(file, siguienteInodo)
		if err != nil {
			return "", fmt.Errorf("error al leer inodo: %v", err)
		}

		// Si es el último elemento, verificar permisos y leer contenido
		if esUltimo {
			// Verificar que sea un archivo
			if inodoActual.I_type[0] != '1' {
				return "", fmt.Errorf("'%s' es una carpeta, no un archivo", parte)
			}

			// Verificar permisos de lectura - Pasa el nombre del archivo
			if !TienePermisoLectura(&inodoActual, global.SesionActiva, parte) { // Añade 'parte' como nombre del archivo
				return "", fmt.Errorf("sin permisos de lectura para '%s'", parte)
			}

			// Leer el contenido del archivo
			return LeerContenidoArchivo(file, sb, &inodoActual)
		}
	}

	return "", fmt.Errorf("ruta inválida")
}

// BuscarEnCarpeta busca un nombre en una carpeta y retorna el inodo
func BuscarEnCarpeta(file *os.File, sb *structures.SuperBloque, inodoCarpeta *structures.TablaInodo, nombre string) (int32, bool, error) {
	for i := 0; i < 12; i++ { // Bloques directos
		if inodoCarpeta.I_block[i] == -1 {
			break
		}

		var bloqueCarpeta structures.BloqueCarpeta
		posicionBloque := inodoCarpeta.I_block[i]

		if _, err := file.Seek(int64(posicionBloque), 0); err != nil {
			return -1, false, err
		}

		if err := binary.Read(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
			return -1, false, err
		}

		for _, entrada := range bloqueCarpeta.B_content { // Cambiado 'j, entrada' a '_ , entrada'
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

// LeerInodo lee un inodo por su índice
func LeerInodo(file *os.File, sb *structures.SuperBloque, indice int32) (structures.TablaInodo, error) {
	var inodo structures.TablaInodo
	posicion := sb.S_inode_start + (indice * sb.S_inode_s) // ¡Usar S_inode_s!
	if _, err := file.Seek(int64(posicion), 0); err != nil {
		return inodo, err
	}
	if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
		return inodo, err
	}
	return inodo, nil
}

// LeerInodoPorPosicion lee un inodo directamente por su posición en bytes
func LeerInodoPorPosicion(file *os.File, posicion int32) (structures.TablaInodo, error) {
	var inodo structures.TablaInodo

	if _, err := file.Seek(int64(posicion), 0); err != nil {
		return inodo, err
	}

	if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
		return inodo, err
	}

	return inodo, nil
}

// TienePermisoLectura verifica si el usuario tiene permiso de lectura
// Recibe el estado de sesión y el nombre del archivo para casos especiales.
func TienePermisoLectura(inodo *structures.TablaInodo, sesion *global.SesionUsuario, nombreArchivo string) bool {
	// Caso especial: Durante el login, se debe poder leer 'users.txt' sin sesión activa
	if sesion == nil && nombreArchivo == "users.txt" {
		// Permitir lectura si es el archivo users.txt y no hay sesión
		return true
	}

	if sesion == nil {
		return false // No hay sesión, y no es el caso especial, no permiso
	}

	// Si es root (UID=1), tiene todos los permisos
	if sesion.UID == 1 {
		return true
	}

	// Obtener permisos
	permisoUser := inodo.I_perm[0]  // Propietario
	permisoGroup := inodo.I_perm[1] // Grupo
	permisoOther := inodo.I_perm[2] // Otros

	// Verificar según la categoría del usuario
	if inodo.I_uid == sesion.UID {
		// Es el propietario - verificar permiso de lectura (r = 4 o mayor)
		return permisoUser >= '4'
	} else if inodo.I_gid == sesion.GID {
		// Es del mismo grupo
		return permisoGroup >= '4'
	} else {
		// Es otro usuario
		return permisoOther >= '4'
	}
}

// LeerContenidoArchivo lee el contenido completo de un archivo
func LeerContenidoArchivo(file *os.File, sb *structures.SuperBloque, inodo *structures.TablaInodo) (string, error) {
	var contenidoTotal strings.Builder

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		if inodo.I_block[i] == -1 {
			break
		}

		var bloqueArchivo structures.BloqueArchivo
		if _, err := file.Seek(int64(inodo.I_block[i]), 0); err != nil {
			return "", fmt.Errorf("error seek bloque %d: %v", i, err)
		}

		if err := binary.Read(file, binary.LittleEndian, &bloqueArchivo); err != nil {
			return "", fmt.Errorf("error lectura bloque %d: %v", i, err)
		}

		// Extraer contenido (hasta encontrar null byte o fin del bloque)
		contenido := string(bloqueArchivo.B_content[:])
		contenido = strings.TrimRight(contenido, "\x00")
		contenidoTotal.WriteString(contenido)
	}

	resultado := contenidoTotal.String()
	return resultado, nil
}

// BuscarBloqueLIbre busca un bloque libre en el bitmap de bloques
func BuscarBloqueLIbre(file *os.File, sb *structures.SuperBloque) int32 {
	for i := int32(0); i < sb.S_blocks_count; i++ {
		var bit byte
		if _, err := file.Seek(int64(sb.S_bm_block_start+i), 0); err != nil {
			continue
		}
		if err := binary.Read(file, binary.LittleEndian, &bit); err != nil {
			continue
		}
		if bit == '0' {
			return sb.S_block_start + (i * size.SizeBloqueArchivo())
		}
	}
	return -1
}

// MarcarBloqueUsado marca un bloque como usado en el bitmap
func MarcarBloqueUsado(file *os.File, sb *structures.SuperBloque, posicionBloque int32) {
	indice := (posicionBloque - sb.S_block_start) / size.SizeBloqueArchivo()
	var bit byte = '1'
	if _, err := file.Seek(int64(sb.S_bm_block_start+indice), 0); err != nil {
		return
	}
	binary.Write(file, binary.LittleEndian, &bit)
}

// marcarBloqueLibre marca un bloque como libre en el bitmap
func marcarBloqueLibre(file *os.File, sb *structures.SuperBloque, posicionBloque int32) {
	indice := (posicionBloque - sb.S_block_start) / size.SizeBloqueArchivo()
	var bit byte = '0'
	if _, err := file.Seek(int64(sb.S_bm_block_start+indice), 0); err != nil {
		return
	}
	binary.Write(file, binary.LittleEndian, &bit)
}

// EscribirArchivoUsersText actualiza el contenido de users.txt
func EscribirArchivoUsersText(file *os.File, sb *structures.SuperBloque, nuevoContenido string) error {
	inodoUsers, err := LeerInodo(file, sb, 1) // Inodo de users.txt es el 1
	if err != nil {
		return err
	}

	bloquesAntiguos := (inodoUsers.I_s + 63) / 64
	if inodoUsers.I_s == 0 {
		bloquesAntiguos = 0
	}

	inodoUsers.I_s = int32(len(nuevoContenido))
	inodoUsers.I_mtime = ObFechaInt() // Asegúrate de que ObFechaInt esté definida en este paquete o importada

	bloquesNecesarios := (len(nuevoContenido) + 63) / 64

	if bloquesNecesarios > 12 {
		return fmt.Errorf("contenido demasiado grande")
	}

	offset := 0
	for i := 0; i < bloquesNecesarios; i++ {
		var bloqueArchivo structures.BloqueArchivo
		fin := offset + 64
		if fin > len(nuevoContenido) {
			fin = len(nuevoContenido)
		}
		copy(bloqueArchivo.B_content[:], nuevoContenido[offset:fin])

		if inodoUsers.I_block[i] == -1 {
			nuevoBloque := BuscarBloqueLIbre(file, sb)
			if nuevoBloque == -1 {
				return fmt.Errorf("no hay bloques libres")
			}
			inodoUsers.I_block[i] = nuevoBloque
			MarcarBloqueUsado(file, sb, nuevoBloque)
		}

		if _, err := file.Seek(int64(inodoUsers.I_block[i]), 0); err != nil {
			return err
		}
		if err := binary.Write(file, binary.LittleEndian, &bloqueArchivo); err != nil {
			return err
		}
		offset += 64
	}

	for i := bloquesNecesarios; i < 12; i++ {
		if inodoUsers.I_block[i] != -1 {
			marcarBloqueLibre(file, sb, inodoUsers.I_block[i])
			inodoUsers.I_block[i] = -1
		}
	}

	diferenciaBloques := int32(bloquesNecesarios) - bloquesAntiguos
	sb.S_free_blocks_count -= diferenciaBloques

	// Escribir el SuperBloque actualizado
	if _, err := file.Seek(int64(sb.S_bm_inode_start-size.SizeSuperBloque()), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, sb); err != nil {
		return err
	}

	// Escribir el inodo de users.txt actualizado
	posInodo := sb.S_inode_start + size.SizeTablaInodo()
	if _, err := file.Seek(int64(posInodo), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &inodoUsers); err != nil {
		return err
	}

	return nil
}

// LimpiarParticion limpia una partición escribiendo ceros
func LimpiarParticion(file *os.File, inicio int32, tamanio int32) error {
	buffer := make([]byte, 1024)
	restante := tamanio

	if _, err := file.Seek(int64(inicio), 0); err != nil {
		return err
	}

	for restante > 0 {
		escribir := int32(1024)
		if restante < escribir {
			escribir = restante
		}

		if _, err := file.Write(buffer[:escribir]); err != nil {
			return err
		}

		restante -= escribir
	}

	return nil
}
