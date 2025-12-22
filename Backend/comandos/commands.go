package comandos

import (
	"Proyecto/comandos/admonDisk"
	/*royecto/comandos/admonUsers"*/
	"fmt"
	"strings"
)

type Handler func(comando string, props map[string]string) (string, bool)

type CommandDef struct {
	Allowed  map[string]bool
	Required []string
	Defaults map[string]string
	Run      Handler
}

var commands = map[string]CommandDef{
	"mkdisk": {
		Allowed: map[string]bool{
			"size": true, "fit": true, "unit": true,
		},
		Required: []string{"size"},
		Defaults: map[string]string{"fit": "FF", "unit": "M"},
		Run:      admonDisk.MkdiskExecute,
	},
	"rmdisk": {
		Allowed: map[string]bool{
			"diskname": true,
		},
		Required: []string{"diskname"},
		Defaults: map[string]string{},
		Run:      admonDisk.RmdiskExecute,
	},
	"fdisk": {
		Allowed: map[string]bool{
			"size": true, "unit": true, "diskName": true, "type": true, "fit": true, "name": true,
		},
		Required: []string{"size", "diskname", "name"},
		Defaults: map[string]string{"unit": "K", "type": "P", "fit": "WF"},
		Run:      admonDisk.FdiskExecute,
	},
	"mount": {
		Allowed: map[string]bool{
			"diskname": true, "name": true,
		},
		Required: []string{"diskname", "name"},
		Defaults: map[string]string{},
		Run:      admonDisk.MountExecute,
	},
	"mounted": {
		Allowed:  map[string]bool{},
		Required: []string{},
		Defaults: map[string]string{},
		Run:      admonDisk.MountedExecute,
	},
	"mkfs": {
		Allowed: map[string]bool{
			"id": true, "type": true,
		},
		Required: []string{"id"},
		Defaults: map[string]string{"type": "FULL"},
		Run:      admonDisk.MkfsExecute,
	},
	"cat": {
		Allowed: map[string]bool{
			"file1": true, "file2": true, "file3": true, "file4": true, "file5": true,
			"file6": true, "file7": true, "file8": true, "file9": true, "file10": true,
		},
		Required: []string{"file1"},
		Defaults: map[string]string{},
		Run:      admonDisk.CatExecute,
	},
	// En el mapa commands, agregar después de "cat":

	/*ogin": {
		Allowed: map[string]bool{
			"user": true, "pass": true, "id": true,
		},
		Required: []string{"user", "pass", "id"},
		Defaults: map[string]string{},
		Run:      admonUsers.LoginExecute,
	},
	"logout": {
		Allowed:  map[string]bool{},
		Required: []string{},
		Defaults: map[string]string{},
		Run:      admonUsers.LogoutExecute,
	},
	"mkgrp": {
		Allowed: map[string]bool{
			"name": true,
		},
		Required: []string{"name"},
		Defaults: map[string]string{},
		Run:      admonUsers.MkgrpExecute,
	},
	"mkusr": {
		Allowed: map[string]bool{
			"user": true, "pass": true, "grp": true,
		},
		Required: []string{"user", "pass", "grp"},
		Defaults: map[string]string{},
		Run:      admonUsers.MkusrExecute,
	},*/
}

func DiskCommandProps(comando string, instrucciones []string) (string, bool) {
	//fmt.Println(comando, instrucciones)
	cmd := strings.ToLower(comando)
	def, ok := commands[cmd]

	if !ok {
		// return nil, fmt.Sprintf("Comando no reconocido: %s", comando), true
		return fmt.Sprintf("Comando no reconocido: %s", comando), true
	}

	// fmt.Println(def)
	allowedLower := make(map[string]bool, len(def.Allowed))
	for k := range def.Allowed {
		allowedLower[strings.ToLower(k)] = true
	}

	props := make(map[string]string)
	for k, v := range def.Defaults {
		props[strings.ToLower(k)] = v
	}

	seen := make(map[string]bool)

	// parseamos los parámetros
	for _, token := range instrucciones {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if !strings.Contains(token, "=") {
			// return nil, fmt.Sprintf("Parámetro inválido: '%v'", token), true
			return fmt.Sprintf("Parámetro inválido: '%v'", token), true
		}

		parts := strings.SplitN(token, "=", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "" {
			// return nil, fmt.Sprintf("Parámetro inválido: '%s'", token), true
			return fmt.Sprintf("Parámetro inválido: '%s'", token), true
		}

		if !allowedLower[key] {
			// return nil, fmt.Sprintf("Parámetro no permitido para '%s': '%s'", cmd, key), true
			return fmt.Sprintf("Parámetro no permitido para '%s': '%s'", cmd, key), true
		}

		if seen[key] {
			// return nil, fmt.Sprintf("Parámetro no permitido: %s", key), true
			return fmt.Sprintf("Parámetro no permitido: %s", key), true
		}

		seen[key] = true
		props[key] = val
	}

	// verificar valores mínimos
	for _, req := range def.Required {
		reqLower := strings.ToLower(req)
		if strings.TrimSpace(props[reqLower]) == "" {
			// return nil, fmt.Sprintf("Parámetro obligatorio faltante: %s", req), true
			return fmt.Sprintf("Parámetro obligatorio faltante: %s", req), true
		}
	}

	// spec, ok := commands[cmd]
	if def.Run == nil {
		return fmt.Sprintf("Comando que no tiene handler: %s", cmd), true
	}

	return def.Run(comando, props)

}

func DiskExecuteCommanWithProps(command string, parameters []string) {
	temp, ok := DiskCommandProps(command, parameters)
	if !ok {
		return
	}

	fmt.Println(temp)
}

func DiskExecuteWithOutput(command string, rawParams map[string]string) (string, bool) {
	// Reutiliza la lógica de validación y ejecución
	// pero sin imprimir, solo retornando
	cmd := strings.ToLower(command)
	def, ok := commands[cmd]
	if !ok {
		return fmt.Sprintf("Comando no reconocido: %s", command), true
	}

	props := make(map[string]string)
	for k, v := range def.Defaults {
		props[strings.ToLower(k)] = v
	}

	for k, v := range rawParams {
		props[strings.ToLower(k)] = v
	}

	// Validar required
	for _, req := range def.Required {
		if strings.TrimSpace(props[strings.ToLower(req)]) == "" {
			return fmt.Sprintf("Parámetro obligatorio faltante: %s", req), true
		}
	}

	return def.Run(command, props)
}
