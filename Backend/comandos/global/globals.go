package global

import "Proyecto/Estructuras/structures"

type ParticionesMontadas struct {
	Path           string
	DriveLetter    byte
	ID_Particion   [4]byte
	Type           byte
	Es_Particion_P bool
	Es_Particion_L bool
	Particion_P    structures.Partition
}

var ParticionMontadaDefault = ParticionesMontadas{
	Path:           "",
	DriveLetter:    0,
	ID_Particion:   Global_ID(""),
	Type:           0,
	Es_Particion_P: false,
	Particion_P:    structures.Partition{},
}

var Mounted_Partitions []ParticionesMontadas

type Grupo struct {
	GID    int32
	Tipo   byte
	Nombre string
}

type Usuario struct {
	UID          int32
	GID          int32
	Tipo         byte
	Grupo        [10]byte
	User         [10]byte
	Password     [10]byte
	ID_Particion [4]byte
	Logged_in    bool
	Mounted      ParticionesMontadas
}

var UsuarioLogeado = Usuario{
	UID:          1,
	GID:          1,
	Tipo:         'U',
	Grupo:        Global_Data("root"),
	User:         Global_Data("root"),
	Password:     Global_Data("123"),
	ID_Particion: Global_ID(""),
	Logged_in:    false,
}

var DefaultUser = Usuario{
	UID:          -1,
	GID:          -1,
	Tipo:         0,
	Grupo:        Global_Data(""),
	User:         Global_Data(""),
	Password:     Global_Data(""),
	ID_Particion: Global_ID(""),
	Logged_in:    false,
	Mounted:      ParticionMontadaDefault,
}

var GrupoUsuarioLoggeado = Grupo{
	GID:    1,
	Tipo:   'G',
	Nombre: "root",
}

var DefaultGrupoUsuario = Grupo{
	GID:    -1,
	Tipo:   0,
	Nombre: "",
}
