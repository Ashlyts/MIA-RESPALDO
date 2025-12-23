package admonDisk

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func MountExecute(comando string, parametros map[string]string) (string, bool) {
	diskName, er, strError := utils.TieneDiskName(parametros["diskname"])
	if er {
		color.Red("[MOUNT ERROR]: %s", strError)
		return strError, er
	}

	nombreParticion, er, strError := utils.TieneName(parametros["name"])
	if er {
		color.Red("[MOUNT ERROR]: %s", strError)
		return strError, er
	}

	return mountPartition(diskName, nombreParticion)
}

func mountPartition(diskName string, nombreParticion string) (string, bool) {
	var nombreSinExtension, extensionArchivo string
	if strings.Contains(diskName, ".") {
		parts := strings.Split(diskName, ".")
		nombreSinExtension = parts[0]
		extensionArchivo = parts[1]
	} else {
		nombreSinExtension = diskName
		extensionArchivo = "mia"
	}

	if strings.ToLower(extensionArchivo) != "mia" {
		msg := "Extensión del archivo no válida. Debe ser .mia"
		color.Red("[MOUNT ERROR]: %s", msg)
		return msg, true
	}

	nombreCompleto := nombreSinExtension + ".mia"
	path := utils.DirectorioDisco + nombreCompleto

	if !utils.ExisteArchivo("MOUNT", path) {
		msg := fmt.Sprintf("Disco no encontrado: %s", path)
		color.Red("[MOUNT ERROR]: %s", msg)
		return msg, true
	}

	mbr, er, strError := utils.ObtenerEstructuraMBR(path)
	if er {
		color.Red("[MOUNT ERROR]: %s", strError)
		return strError, er
	}

	partIndex := -1
	for i := 0; i < 4; i++ {
		partName := utils.ConvertirByteAString(mbr.Mbr_partitions[i].Part_name[:])
		if partName == nombreParticion && mbr.Mbr_partitions[i].Part_s > 0 {
			partIndex = i
			break
		}
	}

	if partIndex == -1 {
		msg := fmt.Sprintf("Partición '%s' no encontrada en el disco", nombreParticion)
		color.Red("[MOUNT ERROR]: %s", msg)
		return msg, true
	}

	if mbr.Mbr_partitions[partIndex].Part_type != 'P' {
		msg := "Solo se pueden montar particiones primarias"
		color.Red("[MOUNT ERROR]: %s", msg)
		return msg, true
	}

	if mbr.Mbr_partitions[partIndex].Part_status == 1 {
		msg := fmt.Sprintf("La partición '%s' ya está montada", nombreParticion)
		color.Yellow("[MOUNT]: %s", msg)

		detalles := fmt.Sprintf(`  Disco:      %s
    Partición:  %s`, nombreCompleto, nombreParticion)
		salida := utils.SuccessBanner("PARTICIÓN YA MONTADA", detalles)
		return salida, false
	}

	letra := obtenerLetraDisco(nombreCompleto)
	correlativo := calcularCorrelativo(&mbr)
	carnetHex := obtenerCarnetHex()
	idParticion := fmt.Sprintf("%s%d%c", carnetHex, correlativo, letra)

	mbr.Mbr_partitions[partIndex].Part_status = 1
	mbr.Mbr_partitions[partIndex].Part_correlative = correlativo
	copy(mbr.Mbr_partitions[partIndex].Part_id[:], idParticion)

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		msg := "[MOUNT ERROR]: No se pudo abrir el disco para escritura"
		color.Red(msg)
		return msg, true
	}
	defer file.Close()

	if _, err := file.Seek(0, 0); err != nil || binary.Write(file, binary.LittleEndian, &mbr) != nil {
		msg := "[MOUNT ERROR]: No se pudo actualizar el MBR"
		color.Red(msg)
		return msg, true
	}

	detalles := fmt.Sprintf(`  Partición:  %s
    Disco:      %s
    ID:         %s
    Letra:      %c
    Correlativo: %d`,
		nombreParticion,
		nombreCompleto,
		idParticion,
		letra,
		correlativo)

	salida := utils.SuccessBanner("PARTICIÓN MONTADA EXITOSAMENTE", detalles)

	color.Green("===========================================================")
	color.Green("PARTICIÓN MONTADA EXITOSAMENTE")
	color.Green("===========================================================")
	color.Cyan("  Partición:  %s", nombreParticion)
	color.Cyan("  Disco:      %s", nombreCompleto)
	color.Cyan("  ID:         %s", idParticion)
	color.Cyan("  Letra:      %c", letra)
	color.Cyan("  Correlativo: %d", correlativo)
	color.Green("===========================================================")

	return salida, false
}

func calcularCorrelativo(mbr *structures.MBR) int32 {
	count := int32(0)
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partitions[i].Part_status == 1 {
			count++
		}
	}
	return count + 1
}

func obtenerLetraDisco(nombreDisco string) byte {
	if strings.HasPrefix(nombreDisco, "VDIC-") && len(nombreDisco) >= 7 {
		letra := nombreDisco[5]
		if letra >= 'A' && letra <= 'Z' {
			return letra
		}
		if letra >= 'a' && letra <= 'z' {
			return letra - 32
		}
	}
	return 'A'
}

func obtenerCarnetHex() string {
	carnet := "202308425"
	if len(carnet) < 2 {
		return "00"
	}
	ultimosDos := carnet[len(carnet)-2:]
	var num int
	fmt.Sscanf(ultimosDos, "%d", &num)
	return fmt.Sprintf("%02X", num)
}
