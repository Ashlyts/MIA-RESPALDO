package general

import (
	"Proyecto/comandos"
	"strings"
)

var commandGroups = map[string][]string{
	"disk":    {"mkdisk", "fdisk", "rmdisk", "mount", "mounted", "mkfs"},
	"reports": {"rep"},
	"files":   {"mkfile", "mkdir"},
	"cat":     {"cat"},
	"users":   {"login", "logout"},
	"groups":  {"mkgrp", "mkusr"},
}

func DetectGroup(cmd string) (string, string, bool, string) {
	// Extraer solo el comando (primera palabra o hasta primer espacio/parámetro)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", "", true, "Comando vacío"
	}

	// Si tiene parámetros con -, extraer solo la palabra antes del primer -
	commandOnly := strings.Split(parts[0], "-")[0]
	cmdLower := strings.ToLower(strings.TrimSpace(commandOnly))

	for group, cmds := range commandGroups {
		for _, exactCmd := range cmds {
			if cmdLower == exactCmd {
				return group, cmdLower, false, ""
			}
		}
	}

	return "", "", true, "Comando no reconocido: " + cmdLower
}

func ParseParamList(raw []string) map[string]string {
	params := make(map[string]string)
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" || !strings.Contains(item, "=") {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		params[key] = value
	}
	return params
}

// error, mssgEror, comandos
// En general/general.go
func GlobalCom(lista []string) ([]string, []string, int) {
	var errores []string
	var salidas []string
	var contErrores = 0

	for _, comm := range lista {
		cmdParts := strings.Fields(comm)
		if len(cmdParts) == 0 {
			salidas = append(salidas, "Línea vacía ignorada")
			continue
		}

		group, command, blnError, strError := DetectGroup(comm)
		if blnError {
			msg := "Error: " + strError
			errores = append(errores, msg)
			salidas = append(salidas, msg)
			contErrores++
			continue
		}

		paramsList := ObtenerParametros(comm)
		paramsMap := ParseParamList(paramsList)
		var salida string
		var err bool

		switch group {
		case "disk":
			salida, err = comandos.DiskExecuteWithOutput(command, paramsMap)
		case "reports":
			salida, err = "Reportes aún no implementados", true
		case "files":
			salida, err = "Manejo de archivos aún no implementado", true
		case "cat":
			salida, err = "Comando CAT aún no implementado", true
		case "users":
			salida, err = "Gestión de usuarios aún no implementada", true
		case "groups":
			salida, err = "Gestión de grupos aún no implementada", true
		default:
			salida, err = "Grupo de comando desconocido", true
		}

		if err {
			errores = append(errores, salida)
			contErrores++
		}
		salidas = append(salidas, salida)
	}

	return errores, salidas, contErrores
}
