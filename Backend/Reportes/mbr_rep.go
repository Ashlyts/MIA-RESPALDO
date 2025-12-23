package Reportes

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generarReporteMBR genera un reporte HTML del MBR y todos los EBRs asociados
func generarReporteMBR(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"

	// 0. Asegurar que el directorio de reportes exista
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP MBR]: Error al crear carpeta %s: %v", repDir, err), true
		}
	} else if err != nil {
		return fmt.Sprintf("[REP MBR]: Error al verificar carpeta %s: %v", repDir, err), true
	}

	// 1. Obtener la partición montada por ID
	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP MBR]: Partición con ID '%s' no encontrada o no montada", id), true
	}

	// 2. Leer el MBR del disco
	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	// 3. Generar el contenido HTML incluyendo EBRs
	htmlContent := generarHtmlMBRConEBRs(mbr, particionMontada.DiskPath, particionMontada.DiskName)

	// 4. Construir la ruta final del reporte
	baseName := filepath.Base(path)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlFilePath := filepath.Join(repDir, htmlFileName)

	// 5. Escribir el archivo HTML
	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP MBR]: Error al escribir archivo HTML en %s: %v", htmlFilePath, err), true
	}

	return fmt.Sprintf("[REP MBR]: Reporte MBR+EBR HTML generado exitosamente en %s", htmlFilePath), false
}

// generarHtmlMBRConEBRs genera HTML con MBR y todos los EBRs
func generarHtmlMBRConEBRs(mbr structures.MBR, diskPath, diskName string) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de MBR + EBRs</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f9f9f9; }
        h2 { color: #6a2c70; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { border: 1px solid #ccc; padding: 10px; text-align: left; }
        th { background-color: #6a2c70; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .particion-primaria { background-color: #e9d8fd; }
        .particion-extendida { background-color: #c8e6c9; }
        .particion-logica { background-color: #ffecb3; }
        .titulo { background-color: #6a2c70; color: white; font-weight: bold; }
        .seccion { margin-top: 30px; }
    </style>
</head>
<body>
    <h2>REPORTE DE MBR + EBRs — Disco: ` + diskName + `</h2>

    <table>
        <tr class="titulo">
            <th colspan="2">Información General del MBR</th>
        </tr>
        <tr><td>mbr_tamano</td><td>` + fmt.Sprintf("%d", mbr.Mbr_tamano) + `</td></tr>
        <tr><td>mbr_fecha_creacion</td><td>` + utils.IntFechaToStr(mbr.Mbr_fecha_creacion) + `</td></tr>
        <tr><td>mbr_disk_signature</td><td>` + fmt.Sprintf("%d", mbr.Mbr_disk_signature) + `</td></tr>
        <tr><td>dsk_fit</td><td>` + string(mbr.Dsk_fit) + `</td></tr>
    </table>
`)

	// Mostrar particiones del MBR
	for i := 0; i < 4; i++ {
		part := mbr.Mbr_partitions[i]
		if part.Part_s == 0 {
			continue // partición no usada
		}

		statusStr := "Inactive"
		if part.Part_status == 1 {
			statusStr = "Active"
		}

		typeStr := string(part.Part_type)
		fitStr := string(part.Part_fit)
		partName := utils.ConvertirByteAString(part.Part_name[:])
		if partName == "" {
			partName = "N/A"
		}

		rowClass := "particion-primaria"
		if typeStr == "E" {
			rowClass = "particion-extendida"
		}

		sb.WriteString(fmt.Sprintf(`
    <table>
        <tr class="%s">
            <th colspan="2">Partición %d (%s)</th>
        </tr>
        <tr><td>part_status</td><td>%s</td></tr>
        <tr><td>part_type</td><td>%s</td></tr>
        <tr><td>part_fit</td><td>%s</td></tr>
        <tr><td>part_start</td><td>%d</td></tr>
        <tr><td>part_size</td><td>%d</td></tr>
        <tr><td>part_name</td><td>%s</td></tr>
        <tr><td>part_correlative</td><td>%d</td></tr>
    </table>
`, rowClass, i+1, typeStr, statusStr, typeStr, fitStr, part.Part_start, part.Part_s, partName, part.Part_correlative))

		// Si es partición extendida, leer y mostrar EBRs
		if typeStr == "E" {
			ebrs := leerEBRs(diskPath, part.Part_start)
			if len(ebrs) == 0 {
				sb.WriteString(`<p style="color: #d32f2f;">⚠️ No se encontraron EBRs en esta partición extendida.</p>`)
			} else {
				for j, ebr := range ebrs {
					ebrName := utils.ConvertirByteAString(ebr.Name[:])
					if ebrName == "" {
						ebrName = "N/A"
					}
					mountStr := "Unmounted"
					if ebr.Part_mount == 1 {
						mountStr = "Mounted"
					}
					fitEBR := string(ebr.Part_fit)

					sb.WriteString(fmt.Sprintf(`
    <table>
        <tr class="particion-logica">
            <th colspan="2">EBR #%d (Unidad Lógica)</th>
        </tr>
        <tr><td>part_mount</td><td>%s</td></tr>
        <tr><td>part_fit</td><td>%s</td></tr>
        <tr><td>part_start</td><td>%d</td></tr>
        <tr><td>part_size</td><td>%d</td></tr>
        <tr><td>part_next</td><td>%d</td></tr>
        <tr><td>part_name</td><td>%s</td></tr>
    </table>
`, j+1, mountStr, fitEBR, ebr.Part_start, ebr.Part_s, ebr.Part_next, ebrName))
				}
			}
		}
	}

	sb.WriteString(`
    <p style="text-align: center; margin-top: 40px; color: #666;">
        Reporte de MBR y EBRs generado el ` + utils.IntFechaToStr(utils.ObFechaInt()) + `
    </p>
</body>
</html>`)

	return sb.String()
}

// leerEBRs lee la cadena enlazada de EBRs desde una partición extendida
func leerEBRs(diskPath string, extendidaStart int32) []structures.EBR {
	var ebrs []structures.EBR
	current := extendidaStart

	file, err := os.Open(diskPath)
	if err != nil {
		return ebrs
	}
	defer file.Close()

	for current != -1 {
		ebr := structures.EBR{}
		// Leer 30 bytes desde la posición 'current'
		buf := make([]byte, 30)
		_, err := file.ReadAt(buf, int64(current))
		if err != nil {
			break
		}

		// Deserializar manualmente (Go no garantiza padding fijo)
		ebr.Part_mount = int8(buf[0])
		ebr.Part_fit = buf[1]
		ebr.Part_start = int32(binary.LittleEndian.Uint32(buf[2:6]))
		ebr.Part_s = int32(binary.LittleEndian.Uint32(buf[6:10]))
		ebr.Part_next = int32(binary.LittleEndian.Uint32(buf[10:14]))
		copy(ebr.Name[:], buf[14:30])

		// Validación mínima: tamaño > 0 y nombre no vacío
		if ebr.Part_s <= 0 {
			break
		}

		ebrs = append(ebrs, ebr)
		if ebr.Part_next == -1 {
			break
		}
		current = ebr.Part_next
	}

	return ebrs
}
