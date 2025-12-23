package Reportes

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strings"
)

// GenerarReporteSB genera el reporte del SuperBloque en formato .html
func generarReporteSB(id string, path string) (string, bool) {
	// 1. Obtener la partición montada por ID
	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP SB]: Partición con ID '%s' no encontrada o no montada", id), true
	}

	// 2. Abrir el archivo del disco
	file, errOpen := os.OpenFile(particionMontada.DiskPath, os.O_RDONLY, 0666)
	if errOpen != nil {
		return "[REP SB]: Error al abrir el disco", true
	}
	defer file.Close()

	// 3. Leer el MBR y el SuperBloque
	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	sb, errSB := utils.LeerSuperBloque(file, mbr.Mbr_partitions[0].Part_start)
	if errSB != nil {
		return "[REP SB]: Error al leer SuperBloque", true
	}

	// 4. Generar el contenido HTML
	htmlContent := generarHtmlSB(sb)

	// 5. Escribir el archivo .html en la carpeta Rep
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		os.MkdirAll(repDir, 0777)
	}

	// Construir la ruta completa para el archivo .html
	finalFileName := strings.TrimSuffix(path, ".png") + ".html"
	if !strings.HasSuffix(finalFileName, ".html") {
		finalFileName += ".html"
	}
	htmlFilePath := repDir + "/" + finalFileName

	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP SB]: Error al escribir archivo HTML: %v", err), true
	}

	return fmt.Sprintf("[REP SB]: Reporte SB HTML generado exitosamente en %s", htmlFilePath), false
}

// generarHtmlSB crea el contenido HTML para el reporte del SuperBloque
func generarHtmlSB(sb structures.SuperBloque) string {
	var sbBuilder strings.Builder

	sbBuilder.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de SuperBloque</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #6a2c70; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .titulo { background-color: #6a2c70; color: white; font-weight: bold; }
    </style>
</head>
<body>
    <h2>REPORTE DE SUPERBLOQUE</h2>
    <table>
        <tr class="titulo">
            <th>Campo</th>
            <th>Valor</th>
        </tr>
`)

	// Agregar cada campo del SuperBloque REAL
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_filesistem_type</td><td>%d</td></tr>
`, sb.S_filesistem_type))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_inodes_count</td><td>%d</td></tr>
`, sb.S_inodes_count))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_blocks_count</td><td>%d</td></tr>
`, sb.S_blocks_count))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_free_blocks_count</td><td>%d</td></tr>
`, sb.S_free_blocks_count))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_free_inodes_count</td><td>%d</td></tr>
`, sb.S_free_inodes_count))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_mtime</td><td>%s</td></tr>
`, utils.IntFechaToStr(sb.S_mtime)))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_umtime</td><td>%s</td></tr>
`, utils.IntFechaToStr(sb.S_umtime)))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_mnt_count</td><td>%d</td></tr>
`, sb.S_mnt_count))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_magic</td><td>0x%X</td></tr>
`, sb.S_magic))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_inode_s</td><td>%d</td></tr>
`, sb.S_inode_s))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_block_s</td><td>%d</td></tr>
`, sb.S_block_s))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_first_ino</td><td>%d</td></tr>
`, sb.S_first_ino))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_first_blo</td><td>%d</td></tr>
`, sb.S_first_blo))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_bm_inode_start</td><td>%d</td></tr>
`, sb.S_bm_inode_start))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_bm_block_start</td><td>%d</td></tr>
`, sb.S_bm_block_start))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_inode_start</td><td>%d</td></tr>
`, sb.S_inode_start))
	sbBuilder.WriteString(fmt.Sprintf(`
        <tr><td>S_block_start</td><td>%d</td></tr>
`, sb.S_block_start))

	sbBuilder.WriteString(`
    </table>
    <p style="text-align: center;">Reporte de SuperBloque</p>
</body>
</html>`)

	return sbBuilder.String()
}
