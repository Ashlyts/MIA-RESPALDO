package admonUsers

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"fmt"

	"github.com/fatih/color"
)

// LogoutExecute maneja el comando logout
func LogoutExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que haya sesión activa
	if global.SesionActiva == nil {
		return "[LOGOUT]: No hay sesión activa", true
	}

	usuarioSaliente := global.SesionActiva.UsuarioActual

	global.SesionActiva = nil

	salida := utils.SuccessBanner(
		"SESIÓN CERRADA EXITOSAMENTE",
		fmt.Sprintf("  Usuario:        %s\n  La sesión ha sido cerrada correctamente", usuarioSaliente),
	)

	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("SESIÓN CERRADA EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  Usuario:        %s", usuarioSaliente)
	color.Yellow("  La sesión ha sido cerrada correctamente")
	color.Green("═══════════════════════════════════════════════════════════")

	return salida, false
}
