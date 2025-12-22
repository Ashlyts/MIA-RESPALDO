package admonUsers

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// MkusrExecute maneja el comando mkusr
func MkusrExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar sesión activa
	if SesionActiva == nil {
		return "[MKUSR]: No hay sesión activa. Use LOGIN primero", true
	}

	// Solo root puede crear usuarios
	if SesionActiva.UsuarioActual != "root" {
		return "[MKUSR]: Solo el usuario root puede crear usuarios", true
	}

	// Validar parámetros
	nombreUsuario := strings.TrimSpace(parametros["user"])
	if nombreUsuario == "" {
		return "[MKUSR]: Parámetro -user es obligatorio", true
	}

	password := strings.TrimSpace(parametros["pass"])
	if password == "" {
		return "[MKUSR]: Parámetro -pass es obligatorio", true
	}

	grupo := strings.TrimSpace(parametros["grp"])
	if grupo == "" {
		return "[MKUSR]: Parámetro -grp es obligatorio", true
	}

	// Validar longitudes
	if len(nombreUsuario) > 10 {
		return "[MKUSR]: El nombre de usuario no puede exceder 10 caracteres", true
	}

	if len(password) > 10 {
		return "[MKUSR]: La contraseña no puede exceder 10 caracteres", true
	}

	if len(grupo) > 10 {
		return "[MKUSR]: El nombre del grupo no puede exceder 10 caracteres", true
	}

	return crearUsuario(nombreUsuario, password, grupo)
}

func crearUsuario(nombreUsuario string, password string, grupo string) (string, bool) {
	// Abrir el disco
	file, err := os.OpenFile(SesionActiva.PathDisco, os.O_RDWR, 0666)
	if err != nil {
		return "[MKUSR]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer SuperBloque
	sb, errSB := leerSuperBloque(file, SesionActiva.Particion.Part_start)
	if errSB != nil {
		return "[MKUSR]: Error al leer SuperBloque", true
	}

	// Leer contenido actual de users.txt
	contenidoActual, errRead := leerArchivoDesdeRuta(file, &sb, "/users.txt")
	if errRead != nil {
		return "[MKUSR]: Error al leer users.txt", true
	}

	// Verificar que el usuario no exista
	if ExisteUsuario(contenidoActual, nombreUsuario) {
		return fmt.Sprintf("[MKUSR]: El usuario '%s' ya existe", nombreUsuario), true
	}

	// Verificar que el grupo exista
	if !ExisteGrupo(contenidoActual, grupo) {
		return fmt.Sprintf("[MKUSR]: El grupo '%s' no existe", grupo), true
	}

	// Calcular nuevo UID
	nuevoUID := ObtenerSiguienteUID(contenidoActual)

	// Agregar nueva línea: UID,U,grupo,usuario,password
	nuevaLinea := fmt.Sprintf("%d,U,%s,%s,%s\n", nuevoUID, grupo, nombreUsuario, password)
	nuevoContenido := contenidoActual + nuevaLinea

	// Escribir el nuevo contenido
	if err := EscribirArchivoUsersText(file, &sb, nuevoContenido); err != nil {
		return "[MKUSR]: Error al escribir en users.txt: " + err.Error(), true
	}

	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("✓ USUARIO CREADO EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  Usuario:        %s", nombreUsuario)
	color.Cyan("  UID:            %d", nuevoUID)
	color.Cyan("  Grupo:          %s", grupo)
	color.Cyan("  Password:       %s", password)
	color.Green("═══════════════════════════════════════════════════════════")

	return "", false
}

// ExisteUsuario verifica si un usuario ya existe
func ExisteUsuario(contenido string, nombreUsuario string) bool {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		if len(partes) >= 4 && partes[1] == "U" && strings.TrimSpace(partes[3]) == nombreUsuario {
			return true
		}
	}
	return false
}

// ObtenerSiguienteUID calcula el siguiente UID disponible
func ObtenerSiguienteUID(contenido string) int32 {
	maxUID := int32(0)
	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		partes := strings.Split(strings.TrimSpace(linea), ",")
		if len(partes) >= 2 && partes[1] == "U" {
			uid, _ := strconv.Atoi(strings.TrimSpace(partes[0]))
			if int32(uid) > maxUID {
				maxUID = int32(uid)
			}
		}
	}

	return maxUID + 1
}
