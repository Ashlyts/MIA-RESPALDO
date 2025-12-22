package admonDisk

import (
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func RmdiskExecute(comando string, props map[string]string) (string, bool) {
	diskName := props["diskname"]

	if diskName == "" {
		msg := "[RMDISK ERROR]: Parámetro 'diskname' vacío"
		color.Red(msg)
		return msg, true
	}

	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		msg := "[RMDISK ERROR]: El disco debe tener extensión .mia"
		color.Red(msg)
		return msg, true
	}

	path := utils.DirectorioDisco + diskName

	if _, err := os.Stat(path); os.IsNotExist(err) {
		msg := fmt.Sprintf("[RMDISK ERROR]: Disco no encontrado: %s", diskName)
		color.Red(msg)
		return msg, true
	}

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("[RMDISK ERROR]: No se pudo eliminar '%s': %v", diskName, err)
		color.Red(msg)
		return msg, true
	}

	msg := fmt.Sprintf("[RMDISK]: Disco '%s' eliminado correctamente", diskName)
	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("✓ %s", msg)
	color.Green("═══════════════════════════════════════════════════════════")
	return msg, false
}
