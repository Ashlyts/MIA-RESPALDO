// admonUsers/login.go
package admonUsers

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/global"
	"Proyecto/comandos/utils"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// LoginExecute maneja el comando login
func LoginExecute(comando string, parametros map[string]string) (string, bool) {
	// Verificar que no haya sesión activa
	if global.SesionActiva != nil {
		return "[LOGIN]: Ya hay una sesión activa. Use LOGOUT primero", true
	}

	// Validar parámetros obligatorios
	usuario := strings.TrimSpace(parametros["user"])
	if usuario == "" {
		return "[LOGIN]: Parámetro -user es obligatorio", true
	}

	password := strings.TrimSpace(parametros["pass"])
	if password == "" {
		return "[LOGIN]: Parámetro -pass es obligatorio", true
	}

	idParticion := strings.TrimSpace(parametros["id"])
	if idParticion == "" {
		return "[LOGIN]: Parámetro -id es obligatorio", true
	}

	return iniciarSesion(usuario, password, idParticion)
}

func iniciarSesion(usuario string, password string, idParticion string) (string, bool) {
	particionMontada, err := admonDisk.GetMountedPartitionByID(idParticion)
	if err != nil {
		return fmt.Sprintf("[LOGIN]: Partición con ID '%s' no encontrada o no montada", idParticion), true
	}

	// Abrir el disco
	file, errOpen := os.OpenFile(particionMontada.DiskPath, os.O_RDONLY, 0666)
	if errOpen != nil {
		return "[LOGIN]: Error al abrir el disco", true
	}
	defer file.Close()

	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	// Buscar la partición
	var particion *structures.Partition
	for i := 0; i < 4; i++ {
		partName := utils.ConvertirByteAString(mbr.Mbr_partitions[i].Part_name[:])
		if partName == particionMontada.PartName {
			particion = &mbr.Mbr_partitions[i]
			break
		}
	}

	if particion == nil {
		return "[LOGIN]: Partición no encontrada en el MBR", true
	}

	// Leer el SuperBloque
	sb, errSB := utils.LeerSuperBloque(file, particion.Part_start)
	if errSB != nil {
		return "[LOGIN]: Partición no formateada o error al leer SuperBloque", true
	}

	// Leer el archivo users.txt
	contenidoUsers, errUsers := utils.LeerArchivoDesdeRuta(file, &sb, "/users.txt")
	if errUsers != nil {
		return "[LOGIN]: Error al leer archivo users.txt: " + errUsers.Error(), true
	}

	// Parsear y validar usuario
	uid, gid, encontrado := ValidarCredenciales(contenidoUsers, usuario, password)
	if !encontrado {
		return "[LOGIN]: Usuario o contraseña incorrectos", true
	}

	// Crear la sesión
	global.SesionActiva = &global.SesionUsuario{
		UsuarioActual: usuario,
		UID:           uid,
		GID:           gid,
		IDParticion:   idParticion,
		PathDisco:     particionMontada.DiskPath,
		Particion:     particion,
	}

	// ✅ Generar mensaje formateado (SIN colores, para frontend)
	salida := fmt.Sprintf(`═══════════════════════════════════════════════════════════
SESIÓN INICIADA EXITOSAMENTE
═══════════════════════════════════════════════════════════
  Usuario:        %s
  UID:            %d
  GID:            %d
  Partición:      %s
  ID:             %s
  Disco:          %s
═══════════════════════════════════════════════════════════`,
		usuario, uid, gid, particionMontada.PartName, idParticion, particionMontada.DiskName)

	// ✅ Opcional: seguir imprimiendo en backend con colores (solo para logs)
	color.Green("═══════════════════════════════════════════════════════════")
	color.Green("SESIÓN INICIADA EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  Usuario:        %s", usuario)
	color.Cyan("  UID:            %d", uid)
	color.Cyan("  GID:            %d", gid)
	color.Cyan("  Partición:      %s", particionMontada.PartName)
	color.Cyan("  ID:             %s", idParticion)
	color.Cyan("  Disco:          %s", particionMontada.DiskName)
	color.Green("═══════════════════════════════════════════════════════════")

	// ✅ Devolver el mensaje limpio (sin ANSI codes)
	return salida, false
}

func ValidarCredenciales(contenido string, usuario string, password string) (int32, int32, bool) {
	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		// Formato: UID,U,grupo,usuario,password
		// Formato: GID,G,nombre_grupo
		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue
		}

		tipo := strings.TrimSpace(partes[1])
		if tipo != "U" {
			continue
		}

		if len(partes) < 5 {
			continue
		}

		nombreUsuario := strings.TrimSpace(partes[3])
		passUsuario := strings.TrimSpace(partes[4])

		if nombreUsuario == usuario && passUsuario == password {
			var uid int32
			fmt.Sscanf(partes[0], "%d", &uid)

			grupo := strings.TrimSpace(partes[2])
			gid := BuscarGIDEnContenido(contenido, grupo)

			return uid, gid, true
		}
	}

	return 0, 0, false
}

// BuscarGIDEnContenido busca el GID de un grupo
func BuscarGIDEnContenido(contenido string, nombreGrupo string) int32 {
	lineas := strings.Split(contenido, "\n")

	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}

		partes := strings.Split(linea, ",")
		if len(partes) < 3 {
			continue
		}

		tipo := strings.TrimSpace(partes[1])
		grupo := strings.TrimSpace(partes[2])

		if tipo == "G" && grupo == nombreGrupo {
			var gid int32
			fmt.Sscanf(partes[0], "%d", &gid)
			return gid
		}
	}

	return 0
}
