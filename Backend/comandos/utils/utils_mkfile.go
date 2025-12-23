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

func TienePermisoEscritura(inodo *structures.TablaInodo, sesion *global.SesionUsuario, _ string) bool {
	if sesion == nil {
		return false
	}

	if sesion.UID == 1 {
		return true
	}

	// Obtener permisos
	permisoUser := inodo.I_perm[0]  // Propietario
	permisoGroup := inodo.I_perm[1] // Grupo
	permisoOther := inodo.I_perm[2] // Otros

	// Verificar según la categoría del usuario
	if inodo.I_uid == sesion.UID {
		// Es el propietario - verificar permiso de escritura (w = 2 o mayor)
		return permisoUser >= '2'
	} else if inodo.I_gid == sesion.GID {
		// Es del mismo grupo
		return permisoGroup >= '2'
	} else {
		// Es otro usuario
		return permisoOther >= '2'
	}
}

func LeerInodoDesdeRuta(file *os.File, sb *structures.SuperBloque, ruta string) (structures.TablaInodo, int32, error) {
	// Asumiendo que la ruta es absoluta (empieza con /)
	partesRuta := strings.Split(strings.Trim(ruta, "/"), "/")
	if len(partesRuta) == 0 || (len(partesRuta) == 1 && partesRuta[0] == "") {
		// Caso especial: ruta es "/"
		// Devolver el inodo raíz, que generalmente está en S_inode_start
		var inodoRaiz structures.TablaInodo
		if _, err := file.Seek(int64(sb.S_inode_start), 0); err != nil {
			return inodoRaiz, -1, fmt.Errorf("error al posicionar puntero para leer inodo raíz: %v", err)
		}
		if err := binary.Read(file, binary.LittleEndian, &inodoRaiz); err != nil {
			return inodoRaiz, -1, fmt.Errorf("error al leer inodo raíz: %v", err)
		}
		return inodoRaiz, sb.S_inode_start, nil // <-- Devuelve inodo, posicion, error
	}

	// Empezar desde el inodo raíz
	posInodoActual := sb.S_inode_start
	var inodoActual structures.TablaInodo

	for i, nombreParte := range partesRuta {
		// Leer el inodo actual
		if _, err := file.Seek(int64(posInodoActual), 0); err != nil {
			return inodoActual, -1, fmt.Errorf("error al posicionar puntero para leer inodo en nivel %d: %v", i, err)
		}
		if err := binary.Read(file, binary.LittleEndian, &inodoActual); err != nil {
			return inodoActual, -1, fmt.Errorf("error al leer inodo en nivel %d: %v", i, err)
		}

		// Verificar que sea un directorio
		if inodoActual.I_type[0] != '0' { // '0' = Carpeta
			return inodoActual, posInodoActual, fmt.Errorf("la ruta '%s' no es un directorio en '%s'", strings.Join(partesRuta[:i+1], "/"), nombreParte)
		}

		// Buscar la parte de la ruta en este inodo (directorio)
		posSiguienteInodo, encontrado, err := BuscarEnCarpeta(file, sb, &inodoActual, nombreParte)
		if err != nil {
			return inodoActual, posInodoActual, fmt.Errorf("error buscando '%s' en directorio '%s': %v", nombreParte, strings.Join(partesRuta[:i], "/"), err)
		}
		if !encontrado {
			return inodoActual, posInodoActual, fmt.Errorf("ruta '%s' no encontrada, '%s' no existe", strings.Join(partesRuta, "/"), nombreParte)
		}

		// Actualizar para la siguiente iteración
		posInodoActual = posSiguienteInodo
	}

	if _, err := file.Seek(int64(posInodoActual), 0); err != nil {
		return inodoActual, -1, fmt.Errorf("error al posicionar puntero para leer inodo final: %v", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &inodoActual); err != nil {
		return inodoActual, -1, fmt.Errorf("error al leer inodo final: %v", err)
	}

	return inodoActual, posInodoActual, nil
}

// BuscarInodoLIbre busca un inodo libre en el bitmap de inodos
func BuscarInodoLIbre(file *os.File, sb *structures.SuperBloque) int32 {
	for i := int32(0); i < sb.S_inodes_count; i++ {
		var bit byte
		if _, err := file.Seek(int64(sb.S_bm_inode_start+i), 0); err != nil {
			continue
		}
		if err := binary.Read(file, binary.LittleEndian, &bit); err != nil {
			continue
		}
		if bit == '0' {
			return sb.S_inode_start + (i * size.SizeTablaInodo())
		}
	}
	return -1
}

// MarcarInodoUsado marca un inodo como usado en el bitmap
func MarcarInodoUsado(file *os.File, sb *structures.SuperBloque, posicionInodo int32) {
	indice := (posicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
	var bit byte = '1'
	if _, err := file.Seek(int64(sb.S_bm_inode_start+indice), 0); err != nil {
		return
	}
	binary.Write(file, binary.LittleEndian, &bit)
}

// CrearArchivo crea un archivo en el directorio padre con el contenido especificado.
func CrearArchivo(file *os.File, sb *structures.SuperBloque, inodoPadre *structures.TablaInodo, nombreArchivo string, contenido string) error {
	// 1. Buscar un inodo libre para el nuevo archivo
	nuevaPosicionInodo := BuscarInodoLIbre(file, sb)
	if nuevaPosicionInodo == -1 {
		return fmt.Errorf("no hay inodos libres")
	}

	// 2. Marcar el inodo como usado
	MarcarInodoUsado(file, sb, nuevaPosicionInodo)
	sb.S_free_inodes_count-- // Actualizar contador en SuperBloque

	// 3. Crear la estructura del nuevo inodo
	var nuevoInodo structures.TablaInodo
	nuevoInodo.I_uid = global.SesionActiva.UID
	nuevoInodo.I_gid = global.SesionActiva.GID
	nuevoInodo.I_s = int32(len(contenido))
	nuevoInodo.I_atime = ObFechaInt()
	nuevoInodo.I_ctime = ObFechaInt()
	nuevoInodo.I_mtime = ObFechaInt()
	for i := range nuevoInodo.I_block {
		nuevoInodo.I_block[i] = -1 // Inicializar bloques en -1
	}
	nuevoInodo.I_type[0] = '1' // '1' = Archivo
	nuevoInodo.I_perm[0] = '6' // Permisos 644 (u+rw, g+r, o+r) por defecto
	nuevoInodo.I_perm[1] = '4'
	nuevoInodo.I_perm[2] = '4'

	// 4. Asignar bloques de datos y escribir contenido
	bloquesNecesarios := (len(contenido) + 63) / 64 // Tamaño de bloque es 64
	if len(contenido) > 0 && bloquesNecesarios == 0 {
		bloquesNecesarios = 1 // Si hay contenido pero < 64 bytes, se necesita al menos 1 bloque
	}

	if bloquesNecesarios > 12 {
		return fmt.Errorf("contenido demasiado grande (más de 12 bloques directos)")
	}

	offset := 0
	for i := 0; i < bloquesNecesarios; i++ {
		var bloqueArchivo structures.BloqueArchivo
		fin := offset + 64
		if fin > len(contenido) {
			fin = len(contenido)
		}
		copy(bloqueArchivo.B_content[:], contenido[offset:fin])

		// Buscar un bloque libre
		nuevoBloquePos := BuscarBloqueLIbre(file, sb)
		if nuevoBloquePos == -1 {
			// Si no hay bloques suficientes, liberar los ya asignados y el inodo
			for j := 0; j < i; j++ {
				marcarBloqueLibre(file, sb, nuevoInodo.I_block[j])
				sb.S_free_blocks_count++
			}
			var bit byte = '0'
			indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
			if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
				binary.Write(file, binary.LittleEndian, &bit)
			}
			sb.S_free_inodes_count++ // Restaurar contador
			return fmt.Errorf("no hay bloques libres suficientes")
		}

		// Marcar bloque como usado
		MarcarBloqueUsado(file, sb, nuevoBloquePos)
		sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

		// Asignar bloque al inodo
		nuevoInodo.I_block[i] = nuevoBloquePos

		// Escribir bloque en disco
		if _, err := file.Seek(int64(nuevoBloquePos), 0); err != nil {
			return fmt.Errorf("error al posicionar puntero para escribir bloque %d: %v", i, err)
		}
		if err := binary.Write(file, binary.LittleEndian, &bloqueArchivo); err != nil {
			return fmt.Errorf("error al escribir bloque %d: %v", i, err)
		}
		offset += 64
	}

	// 5. Escribir el nuevo inodo en disco
	if _, err := file.Seek(int64(nuevaPosicionInodo), 0); err != nil {
		return fmt.Errorf("error al posicionar puntero para escribir inodo: %v", err)
	}
	if err := binary.Write(file, binary.LittleEndian, &nuevoInodo); err != nil {
		return fmt.Errorf("error al escribir inodo: %v", err)
	}

	// 6. Agregar entrada del archivo al directorio padre
	// Buscar un bloque de carpeta disponible en el inodo padre

	var bloqueCarpeta structures.BloqueCarpeta
	for i := 0; i < 12; i++ {
		if inodoPadre.I_block[i] != -1 {
			// Leer bloque existente
			if _, err := file.Seek(int64(inodoPadre.I_block[i]), 0); err != nil {
				continue
			}
			if err := binary.Read(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
				continue
			}
			// Buscar entrada vacía
			for j := 0; j < 4; j++ {
				if bloqueCarpeta.B_content[j].B_inodo == -1 {
					// Encontramos una entrada vacía
					copy(bloqueCarpeta.B_content[j].B_name[:], nombreArchivo)
					bloqueCarpeta.B_content[j].B_inodo = nuevaPosicionInodo
					// Escribir bloque actualizado
					if _, err := file.Seek(int64(inodoPadre.I_block[i]), 0); err != nil {
						return fmt.Errorf("error al posicionar puntero para escribir bloque carpeta: %v", err)
					}
					if err := binary.Write(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
						return fmt.Errorf("error al escribir bloque carpeta: %v", err)
					}
					// Actualizar mtime del directorio padre
					inodoPadre.I_mtime = ObFechaInt()
					// Escribir inodo padre actualizado
					if _, err := file.Seek(int64(sb.S_inode_start), 0); err != nil { // Asumiendo que inodoPadre es el raíz, si no, necesitas la posición real
					}
					// Supongamos que inodoPadre es el raíz (índice 0)
					posInodoPadre := sb.S_inode_start // Esto es para raíz
					if _, err := file.Seek(int64(posInodoPadre), 0); err != nil {
						return fmt.Errorf("error al posicionar puntero para escribir inodo padre: %v", err)
					}
					if err := binary.Write(file, binary.LittleEndian, inodoPadre); err != nil {
						return fmt.Errorf("error al escribir inodo padre: %v", err)
					}
					return nil // Archivo creado exitosamente
				}
			}
		} else if inodoPadre.I_block[i] == -1 {
			// Bloque vacío, crear uno nuevo
			nuevoBloqueCarpetaPos := BuscarBloqueLIbre(file, sb)
			if nuevoBloqueCarpetaPos == -1 {
				// Si no hay bloques, liberar inodo y bloques ya asignados
				for j := 0; j < bloquesNecesarios; j++ {
					marcarBloqueLibre(file, sb, nuevoInodo.I_block[j])
					sb.S_free_blocks_count++
				}
				var bit byte = '0'
				indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
				if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
					binary.Write(file, binary.LittleEndian, &bit)
				}
				sb.S_free_inodes_count++ // Restaurar contador
				return fmt.Errorf("no hay bloques libres para crear bloque de carpeta en el directorio padre")
			}
			MarcarBloqueUsado(file, sb, nuevoBloqueCarpetaPos)
			sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

			// Crear bloque de carpeta con la nueva entrada
			var nuevoBloqueC structures.BloqueCarpeta
			copy(nuevoBloqueC.B_content[0].B_name[:], nombreArchivo)
			nuevoBloqueC.B_content[0].B_inodo = nuevaPosicionInodo
			for k := 1; k < 4; k++ {
				nuevoBloqueC.B_content[k].B_inodo = -1 // Resto vacío
			}

			// Escribir bloque de carpeta nuevo
			if _, err := file.Seek(int64(nuevoBloqueCarpetaPos), 0); err != nil {
				return fmt.Errorf("error al posicionar puntero para escribir nuevo bloque carpeta: %v", err)
			}
			if err := binary.Write(file, binary.LittleEndian, &nuevoBloqueC); err != nil {
				return fmt.Errorf("error al escribir nuevo bloque carpeta: %v", err)
			}

			// Asignar bloque al inodo padre
			inodoPadre.I_block[i] = nuevoBloqueCarpetaPos

			// Actualizar mtime del directorio padre
			inodoPadre.I_mtime = ObFechaInt()

			// Escribir inodo padre actualizado
			posInodoPadre := sb.S_inode_start // Suponiendo raíz
			if _, err := file.Seek(int64(posInodoPadre), 0); err != nil {
				return fmt.Errorf("error al posicionar puntero para escribir inodo padre: %v", err)
			}
			if err := binary.Write(file, binary.LittleEndian, inodoPadre); err != nil {
				return fmt.Errorf("error al escribir inodo padre: %v", err)
			}
			return nil // Archivo creado exitosamente
		}
	}

	return fmt.Errorf("el directorio padre está lleno (solo se manejan bloques directos)")
}
