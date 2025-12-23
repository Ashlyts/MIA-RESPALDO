package Reportes

import (
	"fmt"
	"strings"
)

// RepExecute maneja el comando rep
func RepExecute(comando string, parametros map[string]string) (string, bool) {
	// Validar parámetros obligatorios
	name := strings.TrimSpace(parametros["name"])
	if name == "" {
		return "[REP]: Parámetro -name es obligatorio", true
	}

	id := strings.TrimSpace(parametros["id"])
	if id == "" {
		return "[REP]: Parámetro -id es obligatorio", true
	}

	namereport := strings.TrimSpace(parametros["namereport"])
	if namereport == "" {
		return "[REP]: Parámetro -namereport es obligatorio", true
	}

	switch strings.ToLower(name) {
	case "mbr":
		return generarReporteMBR(id, namereport)
	case "disk": // Añadir cuando lo implementes
		return generarReporteDisk(id, namereport)
	case "inode": // <-- Añadir este caso
		return generarReporteInode(id, namereport)
	case "block": // Añadir cuando lo implementes
		return generarReporteBlock(id, namereport)
	case "tree": // Añadir cuando lo implementes
		return generarReporteTree(id, namereport)
	case "sb": // Añadir cuando lo implementes
		return generarReporteSB(id, namereport)
	case "bm_inode": // Añadir cuando lo implementes
		return generarReporteBMInode(id, namereport)
	case "bm_bloc": // Añadir cuando lo implementes
		return generarReporteBMBloc(id, namereport)
	default:
		return fmt.Sprintf("[REP]: Tipo de reporte '%s' no soportado", name), true
	}
}
