package filecomands

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strings"
)

func CatExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que haya sesión activa
	if global.SesionActiva == nil {
		return "[CAT]: No hay sesión activa. Use el comando LOGIN", true
	}

	// Obtener archivos a mostrar (file1, file2, ..., file10)
	var archivos []string
	for i := 1; i <= 10; i++ {
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

func mostrarContenidoArchivos(rutas []string) (string, bool) {

	var salidaStrings []string
	salidaStrings = append(salidaStrings, "===========================================================")
	salidaStrings = append(salidaStrings, "                    CONTENIDO DE ARCHIVO(S)")
	salidaStrings = append(salidaStrings, "===========================================================\n")

	// Abrir el disco (solo lectura)
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

	fmt.Println("\033[32m===========================================================\033[0m")
	fmt.Println("\033[32m                    CONTENIDO DE ARCHIVO(S)\033[0m")
	fmt.Println("\033[32m===========================================================\n\033[0m")

	// Procesar cada archivo
	for idx, ruta := range rutas {
		if idx > 0 {
			salidaStrings = append(salidaStrings, "")
		}

		salidaStrings = append(salidaStrings, "---------------------------------------------------------")
		salidaStrings = append(salidaStrings, fmt.Sprintf("Archivo: %s", ruta))
		salidaStrings = append(salidaStrings, "---------------------------------------------------------")

		fmt.Printf("\033[36m---------------------------------------------------------\033[0m\n")
		fmt.Printf("\033[33m Archivo: %s\033[0m\n", ruta)
		fmt.Printf("\033[36m---------------------------------------------------------\033[0m\n")

		// Leer contenido
		contenido, errCat := utils.LeerArchivoDesdeRuta(file, &sb, ruta)
		if errCat != nil {

			errorMsg := fmt.Sprintf("Error: %s", errCat.Error())
			salidaStrings = append(salidaStrings, errorMsg)

			fmt.Printf("\033[31m%s\033[0m\n", errorMsg)
			continue
		}

		if contenido == "" {
			salidaStrings = append(salidaStrings, "(archivo vacío)")
		} else {
			salidaStrings = append(salidaStrings, contenido)
		}

		if contenido == "" {
			fmt.Println("(archivo vacío)")
		} else {
			fmt.Println(contenido)
		}
	}

	salidaStrings = append(salidaStrings, "")
	salidaStrings = append(salidaStrings, "===========================================================")

	fmt.Println("\n\033[32m=============================================\033[0m")

	return strings.Join(salidaStrings, "\n"), false
}
