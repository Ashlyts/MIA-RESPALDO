// filecomands/mkgrp.go
package filecomands

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// MkgrpExecute maneja el comando mkgrp
func MkgrpExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar sesión activa
	if global.SesionActiva == nil {
		return "[MKGRP]: No hay sesión activa. Use LOGIN primero", true
	}

	// Solo root puede crear grupos
	if global.SesionActiva.UsuarioActual != "root" {
		return "[MKGRP]: Solo el usuario root puede crear grupos", true
	}

	// Validar parámetro name
	nombreGrupo := strings.TrimSpace(parametros["name"])
	if nombreGrupo == "" {
		return "[MKGRP]: Parámetro -name es obligatorio", true
	}

	if len(nombreGrupo) > 10 {
		return "[MKGRP]: El nombre del grupo no puede exceder 10 caracteres", true
	}

	return crearGrupo(nombreGrupo)
}

func crearGrupo(nombreGrupo string) (string, bool) {
	// Abrir el disco
	file, err := os.OpenFile(global.SesionActiva.PathDisco, os.O_RDWR, 0666)
	if err != nil {
		return "[MKGRP]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer SuperBloque
	sb, errSB := utils.LeerSuperBloque(file, global.SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[MKGRP]: Error al leer SuperBloque", true
	}

	// Leer contenido actual de users.txt
	contenidoActual, errRead := utils.LeerArchivoDesdeRuta(file, &sb, "/users.txt")
	if errRead != nil {
		return "[MKGRP]: Error al leer users.txt", true
	}

	// Verificar que el grupo no exista
	if ExisteGrupo(contenidoActual, nombreGrupo) {
		return fmt.Sprintf("[MKGRP]: El grupo '%s' ya existe", nombreGrupo), true
	}

	// Calcular nuevo GID
	nuevoGID := ObtenerSiguienteGID(contenidoActual)

	// Agregar nueva línea: GID,G,nombre_grupo
	nuevaLinea := fmt.Sprintf("%d,G,%s\n", nuevoGID, nombreGrupo)
	nuevoContenido := contenidoActual + nuevaLinea

	// Escribir el nuevo contenido usando la función de utils
	if err := utils.EscribirArchivoUsersText(file, &sb, nuevoContenido); err != nil {
		return "[MKGRP]: Error al escribir en users.txt: " + err.Error(), true
	}

	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("✓ GRUPO CREADO EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  Nombre:         %s", nombreGrupo)
	color.Cyan("  GID:            %d", nuevoGID)
	color.Green("═══════════════════════════════════════════════════════════")

	return "", false
}

// ExisteGrupo verifica si un grupo ya existe
func ExisteGrupo(contenido string, nombreGrupo string) bool {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		if len(partes) >= 3 && partes[1] == "G" && strings.TrimSpace(partes[2]) == nombreGrupo {
			return true
		}
	}
	return false
}

// ObtenerSiguienteGID calcula el siguiente GID disponible
func ObtenerSiguienteGID(contenido string) int32 {
	maxGID := int32(0)
	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		if len(partes) >= 2 && partes[1] == "G" {
			gid, _ := strconv.Atoi(strings.TrimSpace(partes[0]))
			if int32(gid) > maxGID {
				maxGID = int32(gid)
			}
		}
	}

	return maxGID + 1
}
