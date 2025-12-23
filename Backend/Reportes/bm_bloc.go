package Reportes

import (
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generarReporteBMBloc(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP BM_BLOC]: Error al crear %s: %v", repDir, err), true
		}
	}

	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP BM_BLOC]: Partición con ID '%s' no montada", id), true
	}

	file, err := os.Open(particionMontada.DiskPath)
	if err != nil {
		return "[REP BM_BLOC]: Error al abrir disco", true
	}
	defer file.Close()

	// ✅ Usar la partición correcta (no mbr.Mbr_partitions[0])
	inicioParticion := particionMontada.Partition.Part_start

	sb, err := utils.LeerSuperBloque(file, inicioParticion)
	if err != nil {
		return "[REP BM_BLOC]: Error al leer superbloque", true
	}

	// ✅ Leer bitmap como bytes (1 byte por bloque)
	bitmapBloques := make([]byte, sb.S_blocks_count)
	if _, err := file.Seek(int64(sb.S_bm_block_start), 0); err != nil {
		return "[REP BM_BLOC]: Error al posicionar puntero", true
	}
	if _, err := file.Read(bitmapBloques); err != nil {
		return fmt.Sprintf("[REP BM_BLOC]: Error al leer bitmap: %v", err), true
	}

	txtContent := generarTxtBMBloc(bitmapBloques)

	baseName := filepath.Base(path)
	txtFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".txt"
	txtFilePath := filepath.Join(repDir, txtFileName)

	if err := os.WriteFile(txtFilePath, []byte(txtContent), 0644); err != nil {
		return fmt.Sprintf("[REP BM_BLOC]: Error al escribir: %v", err), true
	}

	return fmt.Sprintf("[REP BM_BLOC]: Reporte generado en %s", txtFilePath), false
}

// generarTxtBMBloc convierte el bitmap (1 byte por bloque) a texto
func generarTxtBMBloc(bitmapBloques []byte) string {
	var sb strings.Builder

	// Convertir cada byte a '0' o '1'
	var line string
	for i, b := range bitmapBloques {
		if b == '1' || b == 1 {
			line += "1"
		} else {
			line += "0"
		}

		// Cada 60 caracteres, hacer salto de línea (ajustable)
		if (i+1)%60 == 0 {
			sb.WriteString(line)
			sb.WriteString("\n")
			line = ""
		}
	}

	// Añadir última línea si queda algo
	if line != "" {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}
