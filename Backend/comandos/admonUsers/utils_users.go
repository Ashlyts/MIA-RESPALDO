package admonUsers

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
)

// buscarBloqueLIbre busca un bloque libre en el bitmap de bloques
func buscarBloqueLIbre(file *os.File, sb *structures.SuperBloque) int32 {
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

// marcarBloqueUsado marca un bloque como usado en el bitmap
func marcarBloqueUsado(file *os.File, sb *structures.SuperBloque, posicionBloque int32) {
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
	inodoUsers, err := leerInodo(file, sb, 1)
	if err != nil {
		return err
	}

	bloquesAntiguos := (inodoUsers.I_s + 63) / 64
	if inodoUsers.I_s == 0 {
		bloquesAntiguos = 0
	}

	inodoUsers.I_s = int32(len(nuevoContenido))
	inodoUsers.I_mtime = utils.ObFechaInt()

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
			nuevoBloque := buscarBloqueLIbre(file, sb)
			if nuevoBloque == -1 {
				return fmt.Errorf("no hay bloques libres")
			}
			inodoUsers.I_block[i] = nuevoBloque
			marcarBloqueUsado(file, sb, nuevoBloque)
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
	if _, err := file.Seek(int64(SesionActiva.Particion.Part_start), 0); err != nil {
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
