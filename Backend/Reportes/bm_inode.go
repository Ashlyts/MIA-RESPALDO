package Reportes

import (
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// GenerarReporteBMInode genera el reporte del bitmap de inodos en formato .txt
func generarReporteBMInode(id string, path string) (string, bool) {
	// 1. Obtener la partición montada por ID
	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP BM_INODE]: Partición con ID '%s' no encontrada o no montada", id), true
	}

	// 2. Abrir el archivo del disco
	file, errOpen := os.OpenFile(particionMontada.DiskPath, os.O_RDONLY, 0666)
	if errOpen != nil {
		return "[REP BM_INODE]: Error al abrir el disco", true
	}
	defer file.Close()

	// 3. Leer el MBR y el SuperBloque
	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	sb, errSB := utils.LeerSuperBloque(file, mbr.Mbr_partitions[0].Part_start)
	if errSB != nil {
		return "[REP BM_INODE]: Error al leer SuperBloque", true
	}

	// 4. Leer el bitmap de inodos
	bitmapInodos := make([]byte, sb.S_inodes_count)
	if _, err := file.Seek(int64(sb.S_bm_inode_start), 0); err != nil {
		return "[REP BM_INODE]: Error al posicionar puntero para leer bitmap de inodos", true
	}
	if err := binary.Read(file, binary.LittleEndian, &bitmapInodos); err != nil {
		return "[REP BM_INODE]: Error al leer bitmap de inodos", true
	}

	// 5. Generar el contenido del archivo .txt
	txtContent := generarTxtBMInode(bitmapInodos, sb.S_inodes_count)

	// 6. Escribir el archivo .txt en la carpeta Rep
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		os.MkdirAll(repDir, 0777)
	}

	// Construir la ruta completa para el archivo .txt
	finalFileName := strings.TrimSuffix(path, ".png") + ".txt"
	if !strings.HasSuffix(finalFileName, ".txt") {
		finalFileName += ".txt"
	}
	txtFilePath := repDir + "/" + finalFileName

	if err := os.WriteFile(txtFilePath, []byte(txtContent), 0644); err != nil {
		return fmt.Sprintf("[REP BM_INODE]: Error al escribir archivo TXT: %v", err), true
	}

	return fmt.Sprintf("[REP BM_INODE]: Reporte BM_INODE generado exitosamente en %s", txtFilePath), false
}

// generarTxtBMInode crea el contenido del archivo .txt para el reporte del bitmap de inodos
func generarTxtBMInode(bitmapInodos []byte, totalInodos int32) string {
	var sb strings.Builder

	// Convertir el bitmap de bytes a una cadena de '0's y '1's
	var bits string
	for i := int32(0); i < totalInodos; i++ {
		bit := (bitmapInodos[i/8] >> (i % 8)) & 1
		if bit == 1 {
			bits += "1"
		} else {
			bits += "0"
		}
	}

	// Formatear en líneas de 20 caracteres
	for i := 0; i < len(bits); i += 20 {
		end := i + 20
		if end > len(bits) {
			end = len(bits)
		}
		linea := bits[i:end]
		sb.WriteString(linea)
		sb.WriteString("\n")
	}

	return sb.String()
}
