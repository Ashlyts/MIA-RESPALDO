package Reportes

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generarReporteBlock(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP BLOCK]: Error al crear %s: %v", repDir, err), true
		}
	}

	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP BLOCK]: Partición con ID '%s' no montada", id), true
	}

	file, err := os.Open(particionMontada.DiskPath)
	if err != nil {
		return "[REP BLOCK]: Error al abrir disco", true
	}
	defer file.Close()

	// ✅ Usar la partición correcta
	inicioParticion := particionMontada.Partition.Part_start

	sb, err := utils.LeerSuperBloque(file, inicioParticion)
	if err != nil {
		return "[REP BLOCK]: Error al leer superbloque", true
	}

	htmlContent := generarHtmlBlock(file, sb)

	// Guardar archivo
	baseName := filepath.Base(path)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlFilePath := filepath.Join(repDir, htmlFileName)

	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP BLOCK]: Error al escribir: %v", err), true
	}

	return fmt.Sprintf("[REP BLOCK]: Reporte generado en %s", htmlFilePath), false
}

func generarHtmlBlock(file *os.File, sb structures.SuperBloque) string {
	var sbBuilder strings.Builder

	sbBuilder.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de Bloques</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f9f9f9; }
        h2 { color: #6a2c70; text-align: center; }
        .container { display: flex; flex-wrap: wrap; gap: 15px; justify-content: center; }
        .bloque {
            border: 1px solid #6a2c70; border-radius: 8px; padding: 12px;
            background: #fff; width: 280px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .titulo { 
            background: #6a2c70; color: white; padding: 6px; text-align: center; 
            border-radius: 4px; margin-bottom: 8px; font-weight: bold;
        }
        .campo { margin: 4px 0; font-size: 13px; word-break: break-all; }
    </style>
</head>
<body>
    <h2>REPORTE DE BLOQUES (solo usados)</h2>
    <div class="container">`)

	// Leer bitmap de bloques
	bitmapSize := int(sb.S_blocks_count)
	bitmapBloques := make([]byte, bitmapSize)
	if _, err := file.Seek(int64(sb.S_bm_block_start), 0); err != nil {
		sbBuilder.WriteString(`<p>Error al leer bitmap de bloques</p>`)
		sbBuilder.WriteString(`</div></body></html>`)
		return sbBuilder.String()
	}
	if _, err := file.Read(bitmapBloques); err != nil {
		sbBuilder.WriteString(`<p>Error al leer bitmap de bloques</p>`)
		sbBuilder.WriteString(`</div></body></html>`)
		return sbBuilder.String()
	}

	// Leer cada bloque usado
	for i := 0; i < int(sb.S_blocks_count); i++ {
		if bitmapBloques[i] != '1' && bitmapBloques[i] != 1 {
			continue
		}

		posBloque := sb.S_block_start + int32(i)*64 // tamaño fijo de 64 bytes
		var bloque [64]byte

		if _, err := file.Seek(int64(posBloque), 0); err != nil {
			continue
		}
		if _, err := file.Read(bloque[:]); err != nil {
			continue
		}

		// Mostrar contenido como texto (limpiar null bytes)
		contenido := strings.TrimRight(string(bloque[:]), "\x00")
		if contenido == "" {
			contenido = "[bloque vacío o binario]"
		}

		sbBuilder.WriteString(fmt.Sprintf(`
        <div class="bloque">
            <div class="titulo">Bloque #%d</div>
            <div class="campo">Posición: %d</div>
            <div class="campo">Contenido:<br/><pre style="margin: 5px 0; background: #f0f0f0; padding: 5px; border-radius: 4px;">%s</pre></div>
        </div>`, i+1, posBloque, contenido))
	}

	sbBuilder.WriteString(`</div>
    <p style="text-align: center; margin-top: 30px; color: #666;">
        Reporte de Bloques generado el ` + utils.IntFechaToStr(utils.ObFechaInt()) + `
    </p>
</body>
</html>`)

	return sbBuilder.String()
}
