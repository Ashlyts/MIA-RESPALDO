package admonUsers

import (
	"github.com/fatih/color"
)

// LogoutExecute maneja el comando logout
func LogoutExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que haya sesión activa
	if SesionActiva == nil {
		return "[LOGOUT]: No hay sesión activa", true
	}

	usuarioSaliente := SesionActiva.UsuarioActual

	// Cerrar sesión
	SesionActiva = nil

	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("✓ SESIÓN CERRADA EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  Usuario:        %s", usuarioSaliente)
	color.Yellow("  La sesión ha sido cerrada correctamente")
	color.Green("═══════════════════════════════════════════════════════════")

	return "", false
}
