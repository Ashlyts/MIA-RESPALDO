package Reportes

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Segmento struct {
	Nombre     string
	Tipo       string // "mbr", "primaria", "extendida", "libre"
	Inicio     int32
	Tamaño     int32
	Porcentaje float64
}

// generarReporteDisk genera el reporte DISK en HTML según el enunciado
func generarReporteDisk(id string, path string) (string, bool) {
	repDir := "VDIC-MIA/Rep"

	// 0. Crear directorio si no existe
	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		if err := os.MkdirAll(repDir, 0755); err != nil {
			return fmt.Sprintf("[REP DISK]: Error al crear carpeta %s: %v", repDir, err), true
		}
	}

	// 1. Obtener partición montada por ID
	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("[REP DISK]: Partición con ID '%s' no encontrada o no montada", id), true
	}

	// 2. Leer MBR
	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	// 3. Generar segmentos del disco
	segmentos := generarSegmentosDisco(mbr, mbr.Mbr_tamano)

	// 4. Generar HTML
	htmlContent := generarHtmlDisk(segmentos, particionMontada.DiskName)

	// 5. Guardar archivo
	baseName := filepath.Base(path)
	htmlFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".html"
	htmlFilePath := filepath.Join(repDir, htmlFileName)

	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		return fmt.Sprintf("[REP DISK]: Error al escribir archivo HTML: %v", err), true
	}

	return fmt.Sprintf("[REP DISK]: Reporte DISK HTML generado exitosamente en %s", htmlFilePath), false
}

// generarSegmentosDisco construye la lista de segmentos físicos del disco
func generarSegmentosDisco(mbr structures.MBR, tamanoTotal int32) []Segmento {
	var segmentos []Segmento

	// 1. Añadir MBR
	segmentos = append(segmentos, Segmento{
		Nombre: "MBR",
		Tipo:   "mbr",
		Inicio: 0,
		Tamaño: size.SizeMBR(),
	})

	// 2. Añadir particiones activas (Part_s > 0)
	for _, p := range mbr.Mbr_partitions {
		if p.Part_s > 0 && p.Part_start >= 0 {
			tipo := "primaria"
			if p.Part_type == 'E' {
				tipo = "extendida"
			}
			nombre := utils.ConvertirByteAString(p.Part_name[:])
			if nombre == "" {
				nombre = "SinNombre"
			}
			segmentos = append(segmentos, Segmento{
				Nombre: nombre,
				Tipo:   tipo,
				Inicio: p.Part_start,
				Tamaño: p.Part_s,
			})
		}
	}

	// 3. Ordenar por posición física
	sort.Slice(segmentos, func(i, j int) bool {
		return segmentos[i].Inicio < segmentos[j].Inicio
	})

	// 4. Recalcular con espacios libres
	var resultado []Segmento
	posActual := int32(0)

	for _, seg := range segmentos {
		// Espacio libre antes del segmento
		if seg.Inicio > posActual {
			libre := Segmento{
				Nombre: "Libre",
				Tipo:   "libre",
				Inicio: posActual,
				Tamaño: seg.Inicio - posActual,
			}
			resultado = append(resultado, libre)
		}
		resultado = append(resultado, seg)
		posActual = seg.Inicio + seg.Tamaño
	}

	// 5. Espacio libre al final
	if posActual < tamanoTotal {
		resultado = append(resultado, Segmento{
			Nombre: "Libre",
			Tipo:   "libre",
			Inicio: posActual,
			Tamaño: tamanoTotal - posActual,
		})
	}

	// 6. Calcular porcentajes
	for i := range resultado {
		resultado[i].Porcentaje = (float64(resultado[i].Tamaño) * 100.0) / float64(tamanoTotal)
	}

	return resultado
}

// generarHtmlDisk crea el HTML con la visualización del disco
func generarHtmlDisk(segmentos []Segmento, diskName string) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>Reporte de Disco</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f9f9f9; }
        h2 { color: #6a2c70; text-align: center; }
        .disco-container { display: flex; justify-content: center; margin: 30px 0; }
        .disco-barra { 
            width: 90%; 
            height: 60px; 
            border: 2px solid #6a2c70; 
            display: flex; 
            position: relative;
        }
        .segmento {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            color: #000;
            text-align: center;
            overflow: hidden;
            white-space: nowrap;
        }
        .mbr { background-color: #e9d8fd; }
        .primaria { background-color: #d9f2d9; }
        .extendida { background-color: #fff2cc; }
        .libre { background-color: #cce5ff; }
        .leyenda {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            gap: 15px;
            margin-top: 20px;
        }
        .leyenda-item {
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .color-indicador {
            width: 20px;
            height: 20px;
            border: 1px solid #666;
        }
        .tabla-resumen {
            margin: 30px auto;
            border-collapse: collapse;
            width: 90%;
        }
        .tabla-resumen th,
        .tabla-resumen td {
            border: 1px solid #ccc;
            padding: 8px;
            text-align: left;
        }
        .tabla-resumen th {
            background-color: #6a2c70;
            color: white;
        }
        .tabla-resumen tr:nth-child(even) {
            background-color: #f2f2f2;
        }
    </style>
</head>
<body>
    <h2>REPORTE DE DISCO — ` + diskName + `</h2>

    <div class="disco-container">
        <div class="disco-barra">`)

	// Generar cada segmento visual
	for _, seg := range segmentos {
		if seg.Tamaño <= 0 {
			continue
		}
		// Ancho relativo (mínimo 1% para visibilidad)
		ancho := seg.Porcentaje
		if ancho < 1 {
			ancho = 1
		}
		nombreMostrar := seg.Nombre
		if len(nombreMostrar) > 12 {
			nombreMostrar = nombreMostrar[:12] + "…"
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="segmento %s" style="flex: 0 0 %.2f%%;" title="%s (%.2f%%)">
                %s
            </div>`,
			seg.Tipo, ancho, seg.Nombre, seg.Porcentaje, nombreMostrar))
	}

	sb.WriteString(`</div></div>

    <table class="tabla-resumen">
        <thead>
            <tr>
                <th>Nombre</th>
                <th>Tipo</th>
                <th>Inicio (bytes)</th>
                <th>Tamaño (bytes)</th>
                <th>Porcentaje</th>
            </tr>
        </thead>
        <tbody>`)

	for _, seg := range segmentos {
		tipoMostrar := map[string]string{
			"mbr":       "MBR",
			"primaria":  "Primaria",
			"extendida": "Extendida",
			"libre":     "Libre",
		}[seg.Tipo]

		sb.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%d</td>
                <td>%d</td>
                <td>%.2f%%</td>
            </tr>`,
			seg.Nombre, tipoMostrar, seg.Inicio, seg.Tamaño, seg.Porcentaje))
	}

	sb.WriteString(`</tbody></table>

    <div class="leyenda">
        <div class="leyenda-item">
            <div class="color-indicador mbr"></div>
            <span>MBR</span>
        </div>
        <div class="leyenda-item">
            <div class="color-indicador primaria"></div>
            <span>Primaria</span>
        </div>
        <div class="leyenda-item">
            <div class="color-indicador extendida"></div>
            <span>Extendida</span>
        </div>
        <div class="leyenda-item">
            <div class="color-indicador libre"></div>
            <span>Libre</span>
        </div>
    </div>

    <p style="text-align: center; margin-top: 40px; color: #666;">
        Reporte de Disco generado el ` + utils.IntFechaToStr(utils.ObFechaInt()) + `
    </p>
</body>
</html>`)

	return sb.String()
}
