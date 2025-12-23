// filecomands/cat.go
package filecomands

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// CatExecute maneja el comando cat
func CatExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que haya sesión activa
	if global.SesionActiva == nil {
		return "[CAT]: No hay sesión activa. Use el comando LOGIN", true
	}

	// Obtener archivos a mostrar (file1, file2, file3, ...)
	var archivos []string
	for i := 1; i <= 10; i++ { // Soportar hasta 10 archivos
		key := fmt.Sprintf("file%d", i)
		if archivo, existe := parametros[key]; existe && archivo != "" {
			archivos = append(archivos, strings.TrimSpace(archivo))
		}
	}

	if len(archivos) == 0 {
		return "[CAT]: Debe especificar al menos un archivo con -file1=ruta", true
	}

	return mostrarContenidoArchivos(archivos)
}

// mostrarContenidoArchivos lee y muestra el contenido de los archivos especificados
func mostrarContenidoArchivos(rutas []string) (string, bool) {
	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("                    CONTENIDO DE ARCHIVO(S)")
	color.Green("═══════════════════════════════════════════════════════════\n")

	// Abrir el disco
	file, err := os.OpenFile(global.SesionActiva.PathDisco, os.O_RDONLY, 0666)
	if err != nil {
		return "[CAT]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer el SuperBloque
	sb, errSB := utils.LeerSuperBloque(file, global.SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[CAT]: Error al leer SuperBloque: " + errSB.Error(), true
	}

	// Mostrar contenido de cada archivo
	for idx, ruta := range rutas {
		if idx > 0 {
			fmt.Println() // Separador entre archivos
		}

		color.Cyan("─────────────────────────────────────────────────────────")
		color.Yellow("Archivo: %s", ruta)
		color.Cyan("─────────────────────────────────────────────────────────")

		contenido, errCat := utils.LeerArchivoDesdeRuta(file, &sb, ruta) // Usa utils.LeerArchivoDesdeRuta
		if errCat != nil {
			color.Red("✗ Error: %s", errCat.Error())
			continue
		}

		// Mostrar el contenido
		if contenido == "" {
			color.White("(archivo vacío)")
		} else {
			color.White("%s", contenido)
		}
	}

	color.Green("\n═══════════════════════════════════════════════════════════")
	return "", false
}
