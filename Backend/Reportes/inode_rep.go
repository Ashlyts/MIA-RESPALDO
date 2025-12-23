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

func generarReporteInode(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP INODE]: Error al crear %s: %v", repDir, err), true
		}
	}

	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP INODE]: Partición con ID '%s' no montada", id), true
	}

	file, err := os.Open(particionMontada.DiskPath)
	if err != nil {
		return "[REP INODE]: Error al abrir disco", true
	}
	defer file.Close()

	inicioParticion := particionMontada.Partition.Part_start

	sb, err := utils.LeerSuperBloque(file, inicioParticion)
	if err != nil {
		return "[REP INODE]: Error al leer superbloque", true
	}

	htmlContent := generarHtmlInode(file, sb)

	baseName := filepath.Base(path)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlFilePath := filepath.Join(repDir, htmlFileName)

	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP INODE]: Error al escribir: %v", err), true
	}

	return fmt.Sprintf("[REP INODE]: Reporte generado en %s", htmlFilePath), false
}

func generarHtmlInode(file *os.File, sb structures.SuperBloque) string {
	var sbBuilder strings.Builder

	sbBuilder.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de Inodos</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f9f9f9; }
        h2 { color: #6a2c70; text-align: center; }
        .container { display: flex; flex-wrap: wrap; gap: 15px; justify-content: center; }
        .inodo {
            border: 1px solid #6a2c70; border-radius: 8px; padding: 12px;
            background: #fff; width: 220px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .titulo { 
            background: #6a2c70; color: white; padding: 6px; text-align: center; 
            border-radius: 4px; margin-bottom: 8px; font-weight: bold;
        }
        .campo { margin: 4px 0; font-size: 13px; }
        .clave { font-weight: bold; color: #333; }
        .valor { margin-left: 6px; color: #555; }
    </style>
</head>
<body>
    <h2>REPORTE DE INODOS (solo usados)</h2>
    <div class="container">`)

	bitmapInodos := make([]byte, sb.S_inodes_count)
	if _, err := file.Seek(int64(sb.S_bm_inode_start), 0); err != nil {
		sbBuilder.WriteString(`<p>Error al leer bitmap</p>`)
		sbBuilder.WriteString(`</div></body></html>`)
		return sbBuilder.String()
	}
	if _, err := file.Read(bitmapInodos); err != nil {
		sbBuilder.WriteString(`<p>Error al leer bitmap</p>`)
		sbBuilder.WriteString(`</div></body></html>`)
		return sbBuilder.String()
	}

	// Leer inodos uno por uno
	for i := int32(0); i < sb.S_inodes_count; i++ {
		if bitmapInodos[i] != '1' && bitmapInodos[i] != 1 {
			continue // no usado
		}

		posInodo := sb.S_inode_start + i*sb.S_inode_s
		var inodo structures.TablaInodo
		if _, err := file.Seek(int64(posInodo), 0); err != nil {
			continue
		}
		if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
			continue
		}

		// Convertir tipo
		tipoStr := "Desconocido"
		if len(inodo.I_type) > 0 {
			switch inodo.I_type[0] {
			case 0:
				tipoStr = "Carpeta"
			case 1:
				tipoStr = "Archivo"
			}
		}

		perm := ""
		if len(inodo.I_perm) >= 3 {
			perm = fmt.Sprintf("%c%c%c", inodo.I_perm[0], inodo.I_perm[1], inodo.I_perm[2])
		} else {
			perm = "???"
		}

		// Mostrar los primeros 12 bloques directos
		blocks := ""
		for j := 0; j < 12 && j < len(inodo.I_block); j++ {
			blocks += fmt.Sprintf("<div class=\"campo\"><span class=\"clave\">Bloque %d:</span> <span class=\"valor\">%d</span></div>", j+1, inodo.I_block[j])
		}

		sbBuilder.WriteString(fmt.Sprintf(`
        <div class="inodo">
            <div class="titulo">Inodo #%d</div>
            <div class="campo"><span class="clave">UID:</span> <span class="valor">%d</span></div>
            <div class="campo"><span class="clave">GID:</span> <span class="valor">%d</span></div>
            <div class="campo"><span class="clave">Tamaño:</span> <span class="valor">%d</span></div>
            <div class="campo"><span class="clave">Tipo:</span> <span class="valor">%s</span></div>
            <div class="campo"><span class="clave">Permisos:</span> <span class="valor">%s</span></div>
            <div class="campo"><span class="clave">ATime:</span> <span class="valor">%s</span></div>
            <div class="campo"><span class="clave">CTime:</span> <span class="valor">%s</span></div>
            <div class="campo"><span class="clave">MTime:</span> <span class="valor">%s</span></div>
            %s
        </div>`,
			i+1,
			inodo.I_uid, inodo.I_gid, inodo.I_s,
			tipoStr, perm,
			utils.IntFechaToStr(inodo.I_atime),
			utils.IntFechaToStr(inodo.I_ctime),
			utils.IntFechaToStr(inodo.I_mtime),
			blocks,
		))
	}

	sbBuilder.WriteString(`</div>
    <p style="text-align: center; margin-top: 30px; color: #666;">
        Reporte de Inodos generado el ` + utils.IntFechaToStr(utils.ObFechaInt()) + `
    </p>
</body>
</html>`)

	return sbBuilder.String()
}
