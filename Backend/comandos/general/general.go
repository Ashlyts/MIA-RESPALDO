package general

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var NamePath = "VDIC-MIA"
var ReportPath = "VDIC-MIA/Rep"
var DiskPath = "VDIC-MIA/Disks"

func ObtenerParametros(comando string) map[string]string {
	parametros := make(map[string]string)
	re := regexp.MustCompile(`-([a-zA-Z]\w*)=("[^"]*"|\S+)`)
	matches := re.FindAllStringSubmatch(comando, -1)

	for _, match := range matches {
		nombre := strings.ToLower(match[1])
		valor := match[2]

		if len(valor) >= 2 && valor[0] == '"' && valor[len(valor)-1] == '"' {
			valor = valor[1 : len(valor)-1]
		}

		parametros[nombre] = valor
	}

	return parametros
}

func CrearCarpeta() {
	// nombre := "VDIC-MIA"
	// reportes := "VDIC-MIA/Rep"
	// discos := "VDIC-MIA/Disks"
	nombreArchivo := "VDIC-MIA/Carpeta.txt"
	// git1 := "Rep/.gitignore"
	// git2 := "VDIC-MIA/PFINAL/Disks/.gitignore"
	// if _, err := os.Stat(nombre); os.IsNotExist(err) {
	if _, err := os.Stat(NamePath); os.IsNotExist(err) {
		// err := os.MkdirAll(nombre, 0777)
		err := os.MkdirAll(NamePath, 0777)
		if err != nil {
			color.Red("Error al crear carpeta", err)
			return
		}

		color.Green("\t\t\t\t\tCarpeta VDIC-MIA creada correctamente")
	} else {
		color.Yellow("\t\t\t\t\tCarpeta VDIC-MIA ya existente")
	}

	// if _, err := os.Stat(reportes); os.IsNotExist(err) {
	if _, err := os.Stat(ReportPath); os.IsNotExist(err) {
		// err := os.Mkdir(reportes, 0777)
		err := os.Mkdir(ReportPath, 0777)
		if err != nil {
			color.Red("Error al crear carpeta", err)
			return
		}
		color.Green("\t\t\t\t\tCarpeta Rep creada correctamente")
	} else {
		color.Yellow("\t\t\t\t\tCarpeta Rep ya existente")
	}

	// if _, err := os.Stat(discos); os.IsNotExist(err) {
	if _, err := os.Stat(DiskPath); os.IsNotExist(err) {
		// err := os.Mkdir(discos, 0777)
		err := os.Mkdir(DiskPath, 0777)
		if err != nil {
			color.Red("Error al crear carpeta", err)
			return
		}
		color.Green("\t\t\t\t\tCarpeta VDIC-MIA/Disks creada correctamente")
	} else {
		color.Yellow("\t\t\t\t\tCarpeta VDIC-MIA/Disks ya existente")
	}

	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		archivo, err := os.Create(nombreArchivo)
		if err != nil {
			fmt.Println("Error al crear archivo")
			return
		}
		defer archivo.Close()

		content := []byte("Proyecto Único\t\t\t\tCreated by ASHLYTS")
		_, err = archivo.Write(content)
		if err != nil {
			color.Red("Error escribiendo archivo:", err)
			return
		}
		color.Green("\t\t\t\t\tArchivo creado correctamente")
	} else {
		color.Yellow("\t\t\t\t\tArchivo existente")
	}
	color.Green("Finalizada la creación de carpetas")
}

func TienePath(x string) string {
	y := strings.Split(x, "=")
	fmt.Print("\t\t\t\t\t\t\tBuscando:")
	color.Yellow(y[1])
	if _, err := os.Stat(y[1]); os.IsNotExist(err) {
		color.Red("Archivo No Encontrado")
		return "nil"
	} else {
		color.Green("Archivo Encontrado")
		return y[1]
	}
}

// func ExecuteCommandList(comandos []string) (string, bool, Resultado) {
func ExecuteCommandList(comandos []string) Resultado {
	var lineas []string
	// _ -> índice
	for _, comando := range comandos {
		linea := strings.TrimSpace(comando)
		if len(linea) > 0 && !strings.HasPrefix(linea, "#") {
			lineas = append(lineas, linea)
		}
	}

	var exportar []string
	reg := regexp.MustCompile(`(.*?)\s*(?:#.*|$)`)
	for _, y := range lineas {
		match := reg.FindStringSubmatch(y)
		//fmt.Println(y, "asdf")
		if len(match) > 1 {
			exportar = append(exportar, match[1])
			//fmt.Println(match[0], "///", match[1])
		}
	}

	return Resultado{"", false, SalidaComandoEjecutado{LstComandos: exportar}}
}
