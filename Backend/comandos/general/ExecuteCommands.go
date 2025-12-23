package general

import (
	"Proyecto/comandos"
	"Proyecto/comandos/admonUsers"
	"Proyecto/comandos/filecomands"
	"fmt"
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
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", "", true, "Comando vacío"
	}

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

func GlobalCom(lista []string) ([]string, []string, int) {
	var errores []string
	var salidas []string
	var contErrores = 0

	for _, comm := range lista {
		comm = strings.TrimSpace(comm)
		if comm == "" || strings.HasPrefix(comm, "#") {
			salidas = append(salidas, "Línea vacía o comentario ignorado")
			continue
		}
		fmt.Printf("Procesando comando: [%s]\n", comm)
		parts := strings.Fields(comm)
		fmt.Printf("Partes: %+v\n", parts)
		if len(parts) == 0 {
			salidas = append(salidas, "Comando vacío")
			continue
		}

		_, command, blnError, strError := DetectGroup(comm)
		if blnError {
			msg := "Error: " + strError
			errores = append(errores, msg)
			salidas = append(salidas, msg)
			contErrores++
			continue
		}

		resto := strings.Join(parts[1:], " ")
		fmt.Printf("Resto (para parámetros): [%s]\n", resto)
		paramsMap := ObtenerParametros(resto)
		fmt.Printf("Parámetros parseados: %+v\n", paramsMap)

		var salida string
		var err bool

		switch command { // ← usa 'command', no 'group'
		case "login":
			salida, err = admonUsers.LoginExecute(comm, paramsMap)
		case "logout":
			salida, err = admonUsers.LogoutExecute(comm, paramsMap)
		case "mkdisk", "fdisk", "rmdisk", "mount", "mounted", "mkfs":
			salida, err = comandos.DiskExecuteWithOutput(command, paramsMap)
		case "mkgrp":
			salida, err = filecomands.MkgrpExecute(comm, paramsMap)
		case "mkusr":
			salida, err = filecomands.MkusrExecute(comm, paramsMap)
		case "cat":
			salida, err = filecomands.CatExecute(comm, paramsMap)
		// ... otros
		default:
			fmt.Printf("🔍 [%s] → salida: %q, err: %v\n", command, salida, err)
			salida, err = "Comando no implementado: "+command, true
		}
		if err {
			errores = append(errores, salida)
			contErrores++
		}
		salidas = append(salidas, salida)
	}

	return errores, salidas, contErrores
}
