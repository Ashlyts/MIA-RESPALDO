// Proyecto/global/session.go
package global

import (
	"Proyecto/Estructuras/structures"
)

// SesionUsuario representa la sesión de usuario actual.
type SesionUsuario struct {
	UsuarioActual string
	UID           int32
	GID           int32
	IDParticion   string
	PathDisco     string
	Particion     *structures.Partition
}

// SesionActiva almacena la sesión de usuario actual.
var SesionActiva *SesionUsuario
