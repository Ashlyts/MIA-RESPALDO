// filecomands/mkdir.go
package filecomands

import (
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func MkdirExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar sesión activa
	if global.SesionActiva == nil {
		return "[MKDIR]: No hay sesión activa. Use LOGIN primero", true
	}

	// Validar parámetros obligatorios
	path := strings.TrimSpace(parametros["path"])
	if path == "" {
		return "[MKDIR]: Parámetro -path es obligatorio", true
	}

	// Parámetro opcional
	rParam := strings.TrimSpace(parametros["p"]) // Cambiado de "r" a "p" según el enunciado
	crearRecursivo := rParam != ""

	return crearDirectorio(path, crearRecursivo)
}

func crearDirectorio(path string, crearRecursivo bool) (string, bool) {
	// Abrir el disco
	file, err := os.OpenFile(global.SesionActiva.PathDisco, os.O_RDWR, 0666)
	if err != nil {
		return "[MKDIR]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer SuperBloque
	sb, errSB := utils.LeerSuperBloque(file, global.SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[MKDIR]: Error al leer SuperBloque", true
	}

	// Dividir la ruta
	ruta := strings.TrimSpace(path)
	if !strings.HasPrefix(ruta, "/") {
		ruta = "/" + ruta
	}
	partes := strings.Split(strings.Trim(ruta, "/"), "/")
	if len(partes) == 0 || (len(partes) == 1 && partes[0] == "") {
		return "[MKDIR]: Ruta inválida para directorio", true
	}

	// Manejar creación recursiva si es necesario
	if crearRecursivo {
		// Verificar que exista la carpeta padre (si no es la raíz)
		if len(partes) > 1 {
			rutaDirectorioPadre := "/" + strings.Join(partes[:len(partes)-1], "/")
			_, _, errDirPadre := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre)
			if errDirPadre != nil {
				// Intentar crear directorios intermedios recursivamente
				// Usamos el primer inodo como punto de partida (raíz)
				errRec := utils.CrearDirectoriosRecursivos(file, &sb, partes, 0, sb.S_inode_start)
				if errRec != nil {
					return fmt.Sprintf("[MKDIR]: Error al crear directorios recursivamente: %v", errRec), true
				}
			} else {
				// Si el padre existe, crear solo el último directorio
				nombreDirectorio := partes[len(partes)-1]
				// Obtener inodo y posición del padre existente
				inodoPadre, posInodoPadre, errDir := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre)
				if errDir != nil {
					return fmt.Sprintf("[MKDIR]: Error al acceder al directorio padre '%s': %v", rutaDirectorioPadre, errDir), true
				}

				// Verificar permisos de escritura en el directorio padre
				if !utils.TienePermisoEscritura(&inodoPadre, global.SesionActiva, "") {
					return fmt.Sprintf("[MKDIR]: No tiene permisos de escritura en el directorio '%s'", rutaDirectorioPadre), true
				}

				// Verificar si el directorio ya existe
				_, existe, errBusqueda := utils.BuscarEnCarpeta(file, &sb, &inodoPadre, nombreDirectorio)
				if errBusqueda != nil {
					return fmt.Sprintf("[MKDIR]: Error buscando directorio en padre '%s': %v", rutaDirectorioPadre, errBusqueda), true
				}
				if existe {
					return fmt.Sprintf("[MKDIR]: El directorio '%s' ya existe en '%s'", nombreDirectorio, rutaDirectorioPadre), true
				}

				// Crear el nuevo directorio
				errCrear := utils.CrearDirectorio(file, &sb, &inodoPadre, posInodoPadre, nombreDirectorio)
				if errCrear != nil {
					return fmt.Sprintf("[MKDIR]: Error al crear directorio '%s': %v", nombreDirectorio, errCrear), true
				}
			}
		} else {
			// Caso donde la ruta es directamente en la raíz, e.g., mkdir -p /nueva_raiz
			// Solo crea el directorio si no existe ya en la raíz
			nombreDirectorio := partes[0] // No hay partes[:len(partes)-1], es directo
			rutaDirectorioPadre := "/"
			inodoPadre, posInodoPadre, errDir := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre)
			if errDir != nil {
				return fmt.Sprintf("[MKDIR]: Error al acceder al directorio raíz: %v", errDir), true
			}

			// Verificar permisos de escritura en el directorio raíz
			if !utils.TienePermisoEscritura(&inodoPadre, global.SesionActiva, "") {
				return fmt.Sprintf("[MKDIR]: No tiene permisos de escritura en el directorio raíz"), true
			}

			// Verificar si el directorio ya existe
			_, existe, errBusqueda := utils.BuscarEnCarpeta(file, &sb, &inodoPadre, nombreDirectorio)
			if errBusqueda != nil {
				return fmt.Sprintf("[MKDIR]: Error buscando directorio en raíz: %v", errBusqueda), true
			}
			if existe {
				// Si ya existe, no hacer nada (según comportamiento típico de mkdir -p)
				color.Green("===========================================================")
				color.Green("DIRECTORIO YA EXISTE, NO SE CREÓ NUEVAMENTE (modo -p)")
				color.Green("===========================================================")
				color.Cyan("  Ruta:           %s", ruta)
				color.Cyan("  Modo:           Recursivo (-p)")
				color.Green("============================================================")
				return "", false
			}

			// Crear el nuevo directorio
			errCrear := utils.CrearDirectorio(file, &sb, &inodoPadre, posInodoPadre, nombreDirectorio)
			if errCrear != nil {
				return fmt.Sprintf("[MKDIR]: Error al crear directorio '%s': %v", nombreDirectorio, errCrear), true
			}
		}
	} else {
		// Creación no recursiva
		// Verificar que el directorio padre exista
		if len(partes) > 1 {
			rutaDirectorioPadre := "/" + strings.Join(partes[:len(partes)-1], "/")
			_, _, errDirPadre := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre)
			if errDirPadre != nil {
				return fmt.Sprintf("[MKDIR]: Directorio padre '%s' no existe", rutaDirectorioPadre), true
			}
		}

		// Crear solo el último directorio en la ruta
		nombreDirectorio := partes[len(partes)-1]
		rutaDirectorioPadre := "/" + strings.Join(partes[:len(partes)-1], "/")
		if rutaDirectorioPadre == "/" {
			rutaDirectorioPadre = "/" // Asegurar que sea la raíz
		}

		// Obtener el inodo Y su posición del directorio padre
		inodoPadre, posInodoPadre, errDir := utils.LeerInodoDesdeRuta(file, &sb, rutaDirectorioPadre) // Llamada CORRECTA
		if errDir != nil {
			return fmt.Sprintf("[MKDIR]: Error al acceder al directorio padre '%s': %v", rutaDirectorioPadre, errDir), true
		}

		// Verificar permisos de escritura en el directorio padre
		if !utils.TienePermisoEscritura(&inodoPadre, global.SesionActiva, "") {
			return fmt.Sprintf("[MKDIR]: No tiene permisos de escritura en el directorio '%s'", rutaDirectorioPadre), true
		}

		// Verificar si el directorio ya existe en el directorio padre
		_, existe, errBusqueda := utils.BuscarEnCarpeta(file, &sb, &inodoPadre, nombreDirectorio)
		if errBusqueda != nil {
			return fmt.Sprintf("[MKDIR]: Error buscando directorio en padre '%s': %v", rutaDirectorioPadre, errBusqueda), true
		}
		if existe {
			return fmt.Sprintf("[MKDIR]: El directorio '%s' ya existe en '%s'", nombreDirectorio, rutaDirectorioPadre), true
		}

		// Crear el nuevo directorio
		errCrear := utils.CrearDirectorio(file, &sb, &inodoPadre, posInodoPadre, nombreDirectorio) // Uso correcto de posInodoPadre
		if errCrear != nil {
			return fmt.Sprintf("[MKDIR]: Error al crear directorio '%s': %v", nombreDirectorio, errCrear), true
		}
	}

	// Actualizar SuperBloque (ya se actualiza dentro de CrearDirectorio o CrearDirectoriosRecursivos)
	// Escribir SuperBloque actualizado
	if _, err := file.Seek(int64(global.SesionActiva.Particion.Part_start), 0); err != nil {
		return "[MKDIR]: Error al posicionar puntero para escribir SuperBloque", true
	}
	if err := binary.Write(file, binary.LittleEndian, &sb); err != nil {
		return "[MKDIR]: Error al escribir SuperBloque actualizado", true
	}

	color.Green("===========================================================")
	color.Green("DIRECTORIO CREADO EXITOSAMENTE")
	color.Green("===========================================================")
	color.Cyan("  Ruta:           %s", ruta)
	if crearRecursivo {
		color.Cyan("  Modo:           Recursivo (-p)")
	}
	color.Green("============================================================")

	return "", false
}
