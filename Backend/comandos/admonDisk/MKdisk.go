package admonDisk

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func MkdiskExecute(comando string, parametros map[string]string) (string, bool) {
	tamanio, er, msg := utils.TieneSize(comando, parametros["size"])
	if er {
		color.Red("[MKDISK ERROR]: %s", msg)
		return msg, er
	}

	unidad, er, msg := utils.TieneUnit(comando, parametros["unit"])
	if er {
		color.Red("[MKDISK ERROR]: %s", msg)
		return msg, er
	}

	fit, er, msg := utils.TieneFit(comando, parametros["fit"])
	if er {
		color.Red("[MKDISK ERROR]: %s", msg)
		return msg, er
	}

	return mkdisk_Create(tamanio, unidad, fit)
}

func mkdisk_Create(_size int32, _unit byte, _fit byte) (string, bool) {
	for i := 0; i < 26; i++ {
		nombreDisco := fmt.Sprintf("VDIC-%c.mia", 'A'+i)
		archivo := utils.DirectorioDisco + nombreDisco
		if _, err := os.Stat(archivo); os.IsNotExist(err) {
			er, strmsg := createDiskFile(archivo, _size, _fit, _unit)
			if er {
				color.Red("[MKDISK ERROR]: %s", strmsg)
				return strmsg, er
			}

			// Construir mensaje para frontend
			unidadStr := string(_unit)
			if _unit == 'K' {
				unidadStr = "KB"
			} else if _unit == 'M' {
				unidadStr = "MB"
			}

			msg := fmt.Sprintf("[MKDISK]: Disco '%s' creado exitosamente con tamaño %d %s", nombreDisco, _size, unidadStr)
			color.Green("===========================================================")
			color.Green(" %s", msg)
			color.Green("===========================================================")

			// Retornar el mismo mensaje para el frontend
			return msg, false
		}
	}

	msg := "No hay más nombres disponibles para discos (límite: VDIC-A.mia a VDIC-Z.mia)"
	color.Red("[MKDISK ERROR]: %s", msg)
	return msg, true
}
func createDiskFile(archivo string, tamanio int32, fit byte, unidad byte) (bool, string) {
	file, err := os.Create(archivo)
	if err != nil {
		color.Red("Error al crear el archivo")
		return true, "Error al crear el archivo"
	}
	defer file.Close()

	var estructura structures.MBR

	tamanioDiscco := utils.ObtenerTamanioDisco(tamanio, unidad)
	estructura.Mbr_tamano = tamanioDiscco
	estructura.Mbr_fecha_creacion = utils.ObFechaInt()
	estructura.Mbr_disk_signature = utils.ObtenerDiskSignature()
	estructura.Dsk_fit = fit
	for i := 0; i < len(estructura.Mbr_partitions); i++ {
		estructura.Mbr_partitions[i] = utils.NuevaPartitionVacia()
	}

	bytes_llenar := make([]byte, int(tamanioDiscco))
	if _, err := file.Write(bytes_llenar); err != nil {
		color.Red("Error al escribir bytes en el disco")
		return true, "Error al escribir bytes en el disco"
	}

	// Cambio de posición del puntero
	if _, err := file.Seek(0, 0); err != nil {
		color.Red("Error al mover puntero del archivo")
		return true, "Error al mover puntero del archivo"
	}

	if err := binary.Write(file, binary.LittleEndian, &estructura); err != nil {
		color.Red("Error al escribir datos del MBR")
		return true, "Error al escribir datos del MBR"
	}

	return false, ""
}
