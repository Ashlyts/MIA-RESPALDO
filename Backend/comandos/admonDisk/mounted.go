package admonDisk

import (
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// ParticionMontada representa una partición montada en el sistema
type ParticionMontada struct {
	ID          string
	DiskName    string
	PartName    string
	Correlative int32
	DiskPath    string
}

// MountedExecute muestra TODAS las particiones montadas y retorna la salida para el frontend
func MountedExecute(comando string, parametros map[string]string) (string, bool) {
	particionesMontadas, err := leerParticionesMontadasDelSistema()
	if err != nil {
		msg := fmt.Sprintf("Error al leer particiones: %v", err)
		color.Red("[MOUNTED ERROR]: %s", msg)
		return msg, true
	}

	if len(particionesMontadas) == 0 {
		// Imprimir en backend
		color.Yellow("═══════════════════════════════════════════════════════════")
		color.Yellow("  No hay particiones montadas en el sistema")
		color.Yellow("═══════════════════════════════════════════════════════════")

		// Retornar para frontend
		return "No hay particiones montadas en el sistema", false
	}

	// === Imprimir en terminal del backend (formato bonito) ===
	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("       PARTICIONES MONTADAS EN EL SISTEMA")
	color.Green("═══════════════════════════════════════════════════════════")
	fmt.Println()
	color.Cyan("%-12s %-20s %-20s %-12s", "ID", "DISCO", "PARTICIÓN", "CORRELATIVO")
	color.White("-----------------------------------------------------------")
	for _, part := range particionesMontadas {
		color.White("%-12s %-20s %-20s %-12d",
			part.ID,
			part.DiskName,
			part.PartName,
			part.Correlative)
	}
	color.White("-----------------------------------------------------------")
	color.Cyan("Total de particiones montadas: %d", len(particionesMontadas))
	fmt.Println()
	color.Green("═══════════════════════════════════════════════════════════")

	// === Construir salida para el frontend (texto limpio) ===
	var salida strings.Builder
	salida.WriteString("═══════════════════════════════════════════════════════════\n")
	salida.WriteString("       PARTICIONES MONTADAS EN EL SISTEMA\n")
	salida.WriteString("═══════════════════════════════════════════════════════════\n\n")
	salida.WriteString(fmt.Sprintf("%-12s %-20s %-20s %-12s\n", "ID", "DISCO", "PARTICIÓN", "CORRELATIVO"))
	salida.WriteString("-----------------------------------------------------------\n")

	for _, part := range particionesMontadas {
		salida.WriteString(fmt.Sprintf("%-12s %-20s %-20s %-12d\n",
			part.ID,
			part.DiskName,
			part.PartName,
			part.Correlative))
	}

	salida.WriteString("-----------------------------------------------------------\n")
	salida.WriteString(fmt.Sprintf("Total de particiones montadas: %d\n", len(particionesMontadas)))
	salida.WriteString("\n═══════════════════════════════════════════════════════════")

	return salida.String(), false
}

// leerParticionesMontadasDelSistema lee todos los discos y encuentra particiones montadas
func leerParticionesMontadasDelSistema() ([]ParticionMontada, error) {
	var particiones []ParticionMontada

	if _, err := os.Stat(utils.DirectorioDisco); os.IsNotExist(err) {
		return particiones, nil
	}

	files, err := filepath.Glob(filepath.Join(utils.DirectorioDisco, "*.mia"))
	if err != nil {
		return nil, err
	}

	for _, diskPath := range files {
		diskName := filepath.Base(diskPath)
		mbr, er, _ := utils.ObtenerEstructuraMBR(diskPath)
		if er {
			continue
		}

		for i := 0; i < 4; i++ {
			part := mbr.Mbr_partitions[i]
			if part.Part_status == 1 && part.Part_s > 0 {
				partName := utils.ConvertirByteAString(part.Part_name[:])
				partID := utils.ConvertirByteAString(part.Part_id[:])

				particiones = append(particiones, ParticionMontada{
					ID:          strings.TrimSpace(partID),
					DiskName:    diskName,
					PartName:    partName,
					Correlative: part.Part_correlative,
					DiskPath:    diskPath,
				})
			}
		}
	}

	return particiones, nil
}

// GetMountedPartitionByID busca una partición montada por su ID en todo el sistema
func GetMountedPartitionByID(id string) (*ParticionMontada, error) {
	particiones, err := leerParticionesMontadasDelSistema()
	if err != nil {
		return nil, err
	}

	for _, part := range particiones {
		if part.ID == id {
			return &part, nil
		}
	}

	return nil, fmt.Errorf("partición con ID '%s' no encontrada o no está montada", id)
}
