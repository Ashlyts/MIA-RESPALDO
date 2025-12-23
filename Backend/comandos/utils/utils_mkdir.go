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

func CrearDirectorio(file *os.File, sb *structures.SuperBloque, inodoPadre *structures.TablaInodo, posInodoPadre int32, nombreDirectorio string) error {
	// 1. Buscar un inodo libre para el nuevo directorio
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
	nuevoInodo.I_s = 0 // Tamaño de carpeta es 0
	nuevoInodo.I_atime = ObFechaInt()
	nuevoInodo.I_ctime = ObFechaInt()
	nuevoInodo.I_mtime = ObFechaInt()
	for i := range nuevoInodo.I_block {
		nuevoInodo.I_block[i] = -1 // Inicializar bloques en -1
	}
	// Cambio de permisos: usar 644 para consistencia con archivos
	nuevoInodo.I_type[0] = '0' // '0' = Carpeta
	nuevoInodo.I_perm[0] = '6' // Permisos 644 (u+rw, g+r, o+r) para carpeta
	nuevoInodo.I_perm[1] = '4'
	nuevoInodo.I_perm[2] = '4'

	bloqueCarpetaInicial := CrearBloqueCarpetaInicial(nuevaPosicionInodo, posInodoPadre)

	// 5. Buscar un bloque libre para el bloque de carpeta
	nuevoBloquePos := BuscarBloqueLIbre(file, sb)
	if nuevoBloquePos == -1 {
		// Si no hay bloques, liberar el inodo
		var bit byte = '0'
		indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
		if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
			binary.Write(file, binary.LittleEndian, &bit)
		}
		sb.S_free_inodes_count++ // Restaurar contador
		return fmt.Errorf("no hay bloques libres para crear el bloque de carpeta del directorio")
	}

	// 6. Marcar bloque como usado
	MarcarBloqueUsado(file, sb, nuevoBloquePos)
	sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

	// 7. Asignar bloque al inodo
	nuevoInodo.I_block[0] = nuevoBloquePos // Usamos el primer bloque directo

	// 8. Escribir bloque de carpeta en disco
	if _, err := file.Seek(int64(nuevoBloquePos), 0); err != nil {
		return fmt.Errorf("error al posicionar puntero para escribir bloque carpeta: %v", err)
	}
	if err := binary.Write(file, binary.LittleEndian, &bloqueCarpetaInicial); err != nil {
		return fmt.Errorf("error al escribir bloque carpeta: %v", err)
	}

	// 9. Escribir el nuevo inodo en disco
	if _, err := file.Seek(int64(nuevaPosicionInodo), 0); err != nil {
		return fmt.Errorf("error al posicionar puntero para escribir inodo: %v", err)
	}
	if err := binary.Write(file, binary.LittleEndian, &nuevoInodo); err != nil {
		return fmt.Errorf("error al escribir inodo: %v", err)
	}

	// 10. Agregar entrada del directorio al directorio padre
	// Buscar un bloque de carpeta disponible en el inodo padre
	bloqueCarpeta := structures.BloqueCarpeta{} // Variable temporal
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
					copy(bloqueCarpeta.B_content[j].B_name[:], nombreDirectorio)
					bloqueCarpeta.B_content[j].B_inodo = nuevaPosicionInodo
					// Escribir bloque actualizado
					if _, err := file.Seek(int64(inodoPadre.I_block[i]), 0); err != nil {
						return fmt.Errorf("error al posicionar puntero para escribir bloque carpeta padre: %v", err)
					}
					if err := binary.Write(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
						return fmt.Errorf("error al escribir bloque carpeta padre: %v", err)
					}
					// Actualizar mtime del directorio padre
					inodoPadre.I_mtime = ObFechaInt()
					// Escribir inodo padre actualizado usando la posicion conocida
					if _, err := file.Seek(int64(posInodoPadre), 0); err != nil { // Usar posInodoPadre
						return fmt.Errorf("error al posicionar puntero para escribir inodo padre: %v", err)
					}
					if err := binary.Write(file, binary.LittleEndian, inodoPadre); err != nil {
						return fmt.Errorf("error al escribir inodo padre: %v", err)
					}
					return nil // Directorio creado exitosamente
				}
			}
		} else if inodoPadre.I_block[i] == -1 {
			// Bloque vacío en el inodo padre, crear uno nuevo
			nuevoBloqueCarpetaPos := BuscarBloqueLIbre(file, sb)
			if nuevoBloqueCarpetaPos == -1 {
				// Si no hay bloques, liberar inodo y el bloque ya asignado al nuevo directorio
				marcarBloqueLibre(file, sb, nuevoBloquePos)
				var bit byte = '0'
				indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
				if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
					binary.Write(file, binary.LittleEndian, &bit)
				}
				sb.S_free_inodes_count++ // Restaurar contador
				sb.S_free_blocks_count++ // Restaurar contador del bloque del inodo
				return fmt.Errorf("no hay bloques libres para crear bloque de carpeta en el directorio padre")
			}
			MarcarBloqueUsado(file, sb, nuevoBloqueCarpetaPos)
			sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

			// Crear bloque de carpeta con la nueva entrada
			var nuevoBloqueC structures.BloqueCarpeta
			copy(nuevoBloqueC.B_content[0].B_name[:], nombreDirectorio)
			nuevoBloqueC.B_content[0].B_inodo = nuevaPosicionInodo
			for k := 1; k < 4; k++ {
				nuevoBloqueC.B_content[k].B_inodo = -1 // Resto vacía
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

			// Escribir inodo padre actualizado usando la posicion conocida
			if _, err := file.Seek(int64(posInodoPadre), 0); err != nil { // Usar posInodoPadre
				return fmt.Errorf("error al posicionar puntero para escribir inodo padre: %v", err)
			}
			if err := binary.Write(file, binary.LittleEndian, inodoPadre); err != nil {
				return fmt.Errorf("error al escribir inodo padre: %v", err)
			}
			return nil // Directorio creado exitosamente
		}
	}

	return fmt.Errorf("el directorio padre está lleno (solo se manejan bloques directos)")
}

func CrearBloqueCarpetaInicial(posNuevoDir int32, posInodoPadre int32) structures.BloqueCarpeta {
	var bloque structures.BloqueCarpeta

	// Entrada 0: . (punto - referencia a sí mismo)
	copy(bloque.B_content[0].B_name[:], ".")
	bloque.B_content[0].B_inodo = posNuevoDir

	// Entrada 1: .. (punto punto - referencia al directorio padre)
	copy(bloque.B_content[1].B_name[:], "..")
	bloque.B_content[1].B_inodo = posInodoPadre // Usar la posición pasada como parámetro

	// Entradas 2 y 3: vacías
	bloque.B_content[2].B_inodo = -1
	bloque.B_content[3].B_inodo = -1

	return bloque
}

// CrearDirectoriosRecursivos intenta crear directorios recursivamente.
func CrearDirectoriosRecursivos(file *os.File, sb *structures.SuperBloque, partes []string, indice int, posInodoActual int32) error {
	if indice >= len(partes) {
		return nil // Todos los directorios en la ruta han sido procesados
	}

	nombreDir := partes[indice]

	// Leer el inodo actual (el directorio donde se intenta crear el siguiente)
	var inodoActual structures.TablaInodo
	if _, err := file.Seek(int64(posInodoActual), 0); err != nil {
		return fmt.Errorf("error al posicionar puntero para leer inodo actual: %v", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &inodoActual); err != nil {
		return fmt.Errorf("error al leer inodo actual: %v", err)
	}

	// Verificar permisos de escritura en el directorio actual
	if !TienePermisoEscritura(&inodoActual, global.SesionActiva, "") {
		return fmt.Errorf("sin permisos de escritura en '%s'", strings.Join(partes[:indice+1], "/"))
	}

	// Buscar si el directorio ya existe en el inodo actual
	posSiguienteInodo, existe, errBusqueda := BuscarEnCarpeta(file, sb, &inodoActual, nombreDir)
	if errBusqueda != nil {
		return fmt.Errorf("error buscando directorio '%s': %v", nombreDir, errBusqueda)
	}

	if existe {
		// El directorio ya existe, continuar con el siguiente nivel
		return CrearDirectoriosRecursivos(file, sb, partes, indice+1, posSiguienteInodo)
	} else {
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
		nuevoInodo.I_s = 0 // Tamaño de carpeta es 0
		nuevoInodo.I_atime = ObFechaInt()
		nuevoInodo.I_ctime = ObFechaInt()
		nuevoInodo.I_mtime = ObFechaInt()
		for i := range nuevoInodo.I_block {
			nuevoInodo.I_block[i] = -1 // Inicializar bloques en -1
		}
		// Cambio de permisos: usar 644 para consistencia con archivos y otros directorios
		nuevoInodo.I_type[0] = '0' // '0' = Carpeta
		nuevoInodo.I_perm[0] = '6' // Permisos 644
		nuevoInodo.I_perm[1] = '4'
		nuevoInodo.I_perm[2] = '4'

		// 4. Crear el bloque de carpeta inicial con . y ..
		// Aquí necesitamos la posición del inodo actual (el padre del nuevo directorio)
		bloqueCarpetaInicial := CrearBloqueCarpetaInicial(nuevaPosicionInodo, posInodoActual) // Pasa la pos del nuevo dir y el inodo del padre

		// 5. Buscar un bloque libre para el bloque de carpeta
		nuevoBloquePos := BuscarBloqueLIbre(file, sb)
		if nuevoBloquePos == -1 {
			// Si no hay bloques, liberar el inodo
			var bit byte = '0'
			indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
			if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
				binary.Write(file, binary.LittleEndian, &bit)
			}
			sb.S_free_inodes_count++ // Restaurar contador
			return fmt.Errorf("no hay bloques libres para crear el bloque de carpeta del directorio")
		}

		// 6. Marcar bloque como usado
		MarcarBloqueUsado(file, sb, nuevoBloquePos)
		sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

		// 7. Asignar bloque al inodo
		nuevoInodo.I_block[0] = nuevoBloquePos // Usamos el primer bloque directo

		// 8. Escribir bloque de carpeta en disco
		if _, err := file.Seek(int64(nuevoBloquePos), 0); err != nil {
			return fmt.Errorf("error al posicionar puntero para escribir bloque carpeta: %v", err)
		}
		if err := binary.Write(file, binary.LittleEndian, &bloqueCarpetaInicial); err != nil {
			return fmt.Errorf("error al escribir bloque carpeta: %v", err)
		}

		// 9. Escribir el nuevo inodo en disco
		if _, err := file.Seek(int64(nuevaPosicionInodo), 0); err != nil {
			return fmt.Errorf("error al posicionar puntero para escribir inodo: %v", err)
		}
		if err := binary.Write(file, binary.LittleEndian, &nuevoInodo); err != nil {
			return fmt.Errorf("error al escribir inodo: %v", err)
		}

		// 10. Agregar entrada del directorio al directorio padre (inodoActual)
		// Buscar un bloque de carpeta disponible en el inodoActual (padre)
		bloqueCarpeta := structures.BloqueCarpeta{}
		bloqueEncontrado := false
		for i := 0; i < 12; i++ {
			if inodoActual.I_block[i] != -1 {
				// Leer bloque existente
				if _, err := file.Seek(int64(inodoActual.I_block[i]), 0); err != nil {
					continue
				}
				if err := binary.Read(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
					continue
				}
				// Buscar entrada vacía
				for j := 0; j < 4; j++ {
					if bloqueCarpeta.B_content[j].B_inodo == -1 {
						// Encontramos una entrada vacía
						copy(bloqueCarpeta.B_content[j].B_name[:], nombreDir)
						bloqueCarpeta.B_content[j].B_inodo = nuevaPosicionInodo
						// Escribir bloque actualizado
						if _, err := file.Seek(int64(inodoActual.I_block[i]), 0); err != nil {
							return fmt.Errorf("error al posicionar puntero para escribir bloque carpeta padre: %v", err)
						}
						if err := binary.Write(file, binary.LittleEndian, &bloqueCarpeta); err != nil {
							return fmt.Errorf("error al escribir bloque carpeta padre: %v", err)
						}
						// Actualizar mtime del directorio padre
						inodoActual.I_mtime = ObFechaInt()
						bloqueEncontrado = true
						break // Salir del bucle de entradas
					}
				}
				if bloqueEncontrado {
					break // Salir del bucle de bloques
				}
			} else if inodoActual.I_block[i] == -1 {
				// Bloque vacío en el inodo padre, crear uno nuevo
				nuevoBloqueCarpetaPos := BuscarBloqueLIbre(file, sb)
				if nuevoBloqueCarpetaPos == -1 {
					// Si no hay bloques, liberar inodo y el bloque ya asignado al nuevo directorio
					marcarBloqueLibre(file, sb, nuevoBloquePos)
					var bit byte = '0'
					indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
					if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
						binary.Write(file, binary.LittleEndian, &bit)
					}
					sb.S_free_inodes_count++ // Restaurar contador
					sb.S_free_blocks_count++ // Restaurar contador del bloque del inodo
					return fmt.Errorf("no hay bloques libres para crear bloque de carpeta en el directorio padre")
				}
				MarcarBloqueUsado(file, sb, nuevoBloqueCarpetaPos)
				sb.S_free_blocks_count-- // Actualizar contador en SuperBloque

				// Crear bloque de carpeta con la nueva entrada
				var nuevoBloqueC structures.BloqueCarpeta
				copy(nuevoBloqueC.B_content[0].B_name[:], nombreDir)
				nuevoBloqueC.B_content[0].B_inodo = nuevaPosicionInodo
				for k := 1; k < 4; k++ {
					nuevoBloqueC.B_content[k].B_inodo = -1 // Resto vacía
				}

				// Escribir bloque de carpeta nuevo
				if _, err := file.Seek(int64(nuevoBloqueCarpetaPos), 0); err != nil {
					return fmt.Errorf("error al posicionar puntero para escribir nuevo bloque carpeta: %v", err)
				}
				if err := binary.Write(file, binary.LittleEndian, &nuevoBloqueC); err != nil {
					return fmt.Errorf("error al escribir nuevo bloque carpeta: %v", err)
				}

				// Asignar bloque al inodo padre
				inodoActual.I_block[i] = nuevoBloqueCarpetaPos

				// Actualizar mtime del directorio padre
				inodoActual.I_mtime = ObFechaInt()

				bloqueEncontrado = true
				break // Salir del bucle de bloques
			}
		}

		if !bloqueEncontrado {
			// No se pudo encontrar un bloque de carpeta disponible en los 12 directos del padre
			// Liberar recursos del nuevo directorio
			marcarBloqueLibre(file, sb, nuevoBloquePos)
			var bit byte = '0'
			indiceInodo := (nuevaPosicionInodo - sb.S_inode_start) / size.SizeTablaInodo()
			if _, err := file.Seek(int64(sb.S_bm_inode_start+indiceInodo), 0); err == nil {
				binary.Write(file, binary.LittleEndian, &bit)
			}
			sb.S_free_inodes_count++ // Restaurar contador
			sb.S_free_blocks_count++ // Restaurar contador del bloque
			return fmt.Errorf("no se pudo añadir entrada al directorio padre (sin bloques directos disponibles)")
		}

		if _, err := file.Seek(int64(posInodoActual), 0); err != nil {
			return fmt.Errorf("error al posicionar puntero para escribir inodo padre actualizado: %v", err)
		}
		if err := binary.Write(file, binary.LittleEndian, &inodoActual); err != nil {
			return fmt.Errorf("error al escribir inodo padre actualizado: %v", err)
		}

		// Continuar recursivamente con el siguiente nivel, usando la nueva posición
		return CrearDirectoriosRecursivos(file, sb, partes, indice+1, nuevaPosicionInodo)
	}

}
