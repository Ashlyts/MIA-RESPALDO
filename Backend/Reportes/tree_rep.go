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

func generarReporteTree(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP TREE]: Error al crear %s: %v", repDir, err), true
		}
	}

	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP TREE]: Partición con ID '%s' no montada", id), true
	}

	file, err := os.Open(particionMontada.DiskPath)
	if err != nil {
		return "[REP TREE]: Error al abrir disco", true
	}
	defer file.Close()

	inicioParticion := particionMontada.Partition.Part_start
	sb, err := utils.LeerSuperBloque(file, inicioParticion)
	if err != nil {
		return "[REP TREE]: Error al leer superbloque", true
	}

	htmlContent := generarHtmlTreeVisual(file, sb)

	baseName := filepath.Base(path)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlFilePath := filepath.Join(repDir, htmlFileName)

	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP TREE]: Error al escribir: %v", err), true
	}

	return fmt.Sprintf("[REP TREE]: Reporte generado en %s", htmlFilePath), false
}

func generarHtmlTreeVisual(file *os.File, sb structures.SuperBloque) string {
	var sbBuilder strings.Builder

	sbBuilder.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de Árbol del Sistema de Archivos</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f9f9f9; }
        h2 { color: #6a2c70; text-align: center; }
        .tree-container { position: relative; padding: 20px; }
        .node {
            display: inline-block;
            padding: 8px;
            margin: 10px;
            border: 2px solid #666;
            border-radius: 6px;
            font-size: 12px;
            line-height: 1.4;
            min-width: 120px;
            text-align: center;
        }
        .inode { background: #cce5ff; } /* azul claro */
        .block-carpeta { background: #ffcccc; } /* rojo claro */
        .block-archivo { background: #ffffcc; } /* amarillo claro */
        .block-apuntador { background: #e6f3ff; } /* azul muy claro */
        .connector {
            position: absolute;
            border-left: 2px solid #999;
            height: 20px;
            left: 50%;
            transform: translateX(-50%);
        }
        .children {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-top: 20px;
        }
        .node-label {
            font-weight: bold;
            margin-bottom: 5px;
        }
        .node-info {
            font-size: 10px;
            color: #333;
        }
    </style>
</head>
<body>
    <h2>ÁRBOL DEL SISTEMA DE ARCHIVOS</h2>
    <div class="tree-container">`)

	var procesarInodo func(int32, int)
	procesarInodo = func(numInodo int32, profundidad int) {
		posInodo := sb.S_inode_start + numInodo*92
		var inodo structures.TablaInodo
		if _, err := file.Seek(int64(posInodo), 0); err != nil {
			return
		}
		if err := binary.Read(file, binary.LittleEndian, &inodo); err != nil {
			return
		}

		esCarpeta := (inodo.I_type[0] == 0)

		// Mostrar inodo
		sbBuilder.WriteString(fmt.Sprintf(`
        <div class="node inode" style="margin-left: %dpx;">
            <div class="node-label">Inodo %d</div>
            <div class="node-info">i_type: %d<br/>i_uid: %d<br/>i_gid: %d<br/>i_s: %d<br/>i_perm: %c%c%c</div>
        </div>`, profundidad*40, numInodo, inodo.I_type[0], inodo.I_uid, inodo.I_gid, inodo.I_s, inodo.I_perm[0], inodo.I_perm[1], inodo.I_perm[2]))

		if esCarpeta {
			sbBuilder.WriteString(`<div class="children">`)
			for i := 0; i < 12; i++ {
				if inodo.I_block[i] != -1 {
					posBloque := sb.S_block_start + (inodo.I_block[i] * 64)
					var bloque structures.BloqueCarpeta
					if _, err := file.Seek(int64(posBloque), 0); err != nil {
						continue
					}
					if err := binary.Read(file, binary.LittleEndian, &bloque); err != nil {
						continue
					}

					// Mostrar bloque de carpeta
					sbBuilder.WriteString(fmt.Sprintf(`
                        <div class="node block-carpeta">
                            <div class="node-label">Bloque Carpeta %d</div>
                            <div class="node-info">`, inodo.I_block[i]))
					for j := 0; j < 4; j++ {
						if bloque.B_content[j].B_inodo != -1 {
							nombre := strings.TrimRight(string(bloque.B_content[j].B_name[:]), "\x00")
							if nombre != "" && nombre != "." && nombre != ".." {
								sbBuilder.WriteString(fmt.Sprintf("→ %s (%d)<br/>", nombre, bloque.B_content[j].B_inodo))
							}
						}
					}
					sbBuilder.WriteString(`</div></div>`)

					// Procesar hijos recursivamente
					for j := 0; j < 4; j++ {
						if bloque.B_content[j].B_inodo != -1 {
							nombreHijo := strings.TrimRight(string(bloque.B_content[j].B_name[:]), "\x00")
							if nombreHijo != "" && nombreHijo != "." && nombreHijo != ".." {
								sbBuilder.WriteString(`<div class="connector"></div>`)
								procesarInodo(bloque.B_content[j].B_inodo, profundidad+1)
							}
						}
					}
				}
			}
			sbBuilder.WriteString(`</div>`)
		}
	}

	procesarInodo(0, 0)

	sbBuilder.WriteString(`
    </div>
    <p style="text-align: center; margin-top: 30px; color: #666;">
        Reporte de Árbol generado el ` + utils.IntFechaToStr(utils.ObFechaInt()) + `
    </p>
</body>
</html>`)

	return sbBuilder.String()
}
