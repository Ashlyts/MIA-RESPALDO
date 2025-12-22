package admonUsers

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/admonDisk"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"
)

// mkfsExecute maneja el comando mkfs
func MkfsExecute(comando string, parametros map[string]string) (string, bool) {
	// Validar parámetro obligatorio: id
	idParam := parametros["id"]
	if idParam == "" {
		return "[MKFS]: Parámetro -id es obligatorio", true
	}

	// Limpiar el ID (NO convertir a entero, es una cadena)
	id := strings.TrimSpace(idParam)

	// Validar longitud del ID
	if len(id) < 3 || len(id) > 5 {
		return fmt.Sprintf("[MKFS]: ID inválido '%s'", id), true
	}

	// Validar parámetro opcional: type (por defecto FULL)
	tipoFormateo := strings.ToUpper(strings.TrimSpace(parametros["type"]))
	if tipoFormateo == "" {
		tipoFormateo = "FULL"
	}

	if tipoFormateo != "FULL" {
		return "Solo se soporta formateo FULL", true
	}

	return formatearParticion(id, tipoFormateo)
}

func formatearParticion(id string, tipoFormateo string) (string, bool) {
	// Buscar la partición montada por ID
	particionMontada, err := admonDisk.GetMountedPartitionByID(id)
	if err != nil {
		return fmt.Sprintf("Partición con ID '%s' no encontrada o no montada", id), true
	}

	// Abrir el disco
	file, errOpen := os.OpenFile(particionMontada.DiskPath, os.O_RDWR, 0666)
	if errOpen != nil {
		return "[MKFS]: Error al abrir el disco", true
	}
	defer file.Close()

	// Leer el MBR
	mbr, er, strError := utils.ObtenerEstructuraMBR(particionMontada.DiskPath)
	if er {
		return strError, er
	}

	// Buscar la partición en el MBR
	partIndex := -1
	for i := 0; i < 4; i++ {
		partName := utils.ConvertirByteAString(mbr.Mbr_partitions[i].Part_name[:])
		if partName == particionMontada.PartName {
			partIndex = i
			break
		}
	}

	if partIndex == -1 {
		return "Partición no encontrada en el MBR", true
	}

	particion := mbr.Mbr_partitions[partIndex]

	// Verificar que sea partición primaria
	if particion.Part_type != 'P' {
		return "Solo se pueden formatear particiones primarias", true
	}

	// Realizar el formateo EXT2
	return formatearEXT2(file, &particion, id, particionMontada.PartName)
}

func formatearEXT2(file *os.File, particion *structures.Partition, id string, nombrePart string) (string, bool) {
	tamanioParticion := particion.Part_s
	inicioParticion := particion.Part_start

	// Calcular número de estructuras según fórmula del proyecto
	// tamaño_particion = sizeof(superblock) + n + 3*n + n*sizeof(inodos) + 3*n*sizeof(block)
	// Despejando n:
	n := float64(tamanioParticion-size.SizeSuperBloque()) /
		float64(4+size.SizeTablaInodo()+3*size.SizeBloqueArchivo())

	numeroEstructuras := int32(math.Floor(n))
	numeroBloques := 3 * numeroEstructuras

	color.Cyan("\n→ Formateando partición como EXT2...")
	color.Yellow("  Calculando estructuras:")
	color.White("    • Inodos: %d", numeroEstructuras)
	color.White("    • Bloques: %d", numeroBloques)

	// ==================== PASO 1: LIMPIAR PARTICIÓN ====================
	color.Cyan("\n→ Limpiando partición...")
	if err := limpiarParticion(file, inicioParticion, tamanioParticion); err != nil {
		return "[MKFS]: Error al limpiar partición", true
	}

	// ==================== PASO 2: CREAR SUPERBLOQUE ====================
	color.Cyan("→ Creando SuperBloque...")
	sb := crearSuperBloque(numeroEstructuras, numeroBloques, inicioParticion)

	// Escribir SuperBloque
	if _, err := file.Seek(int64(inicioParticion), 0); err != nil {
		return "[MKFS]: Error al posicionar puntero en SuperBloque", true
	}
	if err := binary.Write(file, binary.LittleEndian, &sb); err != nil {
		return "[MKFS]: Error al escribir SuperBloque", true
	}

	// ==================== PASO 3: INICIALIZAR BITMAPS ====================
	color.Cyan("→ Inicializando Bitmaps...")
	if err := inicializarBitmaps(file, &sb, numeroEstructuras, numeroBloques); err != nil {
		return "[MKFS]: Error al inicializar bitmaps", true
	}

	// ==================== PASO 4: CREAR INODO RAÍZ ====================
	color.Cyan("→ Creando inodo raíz (/)...")
	inodoRaiz := crearInodoRaiz(&sb)

	if _, err := file.Seek(int64(sb.S_inode_start), 0); err != nil {
		return "[MKFS]: Error al posicionar puntero en inodo raíz", true
	}
	if err := binary.Write(file, binary.LittleEndian, &inodoRaiz); err != nil {
		return "[MKFS]: Error al escribir inodo raíz", true
	}

	// ==================== PASO 5: CREAR BLOQUE CARPETA RAÍZ ====================
	color.Cyan("→ Creando bloque carpeta raíz...")
	bloqueCarpetaRaiz := crearBloqueCarpetaRaiz(&sb)

	if _, err := file.Seek(int64(sb.S_block_start), 0); err != nil {
		return "[MKFS]: Error al posicionar puntero en bloque carpeta", true
	}
	if err := binary.Write(file, binary.LittleEndian, &bloqueCarpetaRaiz); err != nil {
		return "[MKFS]: Error al escribir bloque carpeta raíz", true
	}

	// ==================== PASO 6: CREAR users.txt ====================
	color.Cyan("Creando archivo users.txt...")
	if err := crearArchivoUsers(file, &sb); err != nil {
		return "[MKFS]: Error al crear users.txt", true
	}

	// ==================== PASO 7: ACTUALIZAR MBR ====================
	color.Cyan("→ Actualizando MBR...")
	// La partición ya está formateada, no necesitamos cambiar nada en el MBR
	// El estado de montado ya se maneja en mount/mounted

	// ==================== RESULTADO ====================
	color.Green("\n═══════════════════════════════════════════════════════════")
	color.Green("✓ FORMATEO COMPLETADO EXITOSAMENTE")
	color.Green("═══════════════════════════════════════════════════════════")
	color.Cyan("  ID:                %s", id)
	color.Cyan("  Partición:         %s", nombrePart)
	color.Cyan("  Sistema Archivos:  EXT2")
	color.Cyan("  Tipo Formateo:     FULL")
	color.Green("  ───────────────────────────────────────────────────────")
	color.Cyan("  Total Inodos:      %d", numeroEstructuras)
	color.Cyan("  Inodos Libres:     %d", numeroEstructuras-2)
	color.Cyan("  Total Bloques:     %d", numeroBloques)
	color.Cyan("  Bloques Libres:    %d", numeroBloques-2)
	color.Green("  ───────────────────────────────────────────────────────")
	color.Yellow("  Archivos creados:")
	color.White("    • / (raíz)")
	color.White("    • /users.txt")
	color.Green("═══════════════════════════════════════════════════════════\n")

	return "", false
}

// limpiarParticion llena la partición con ceros
func limpiarParticion(file *os.File, inicio int32, tamanio int32) error {
	buffer := make([]byte, 1024)
	restante := tamanio

	if _, err := file.Seek(int64(inicio), 0); err != nil {
		return err
	}

	for restante > 0 {
		escribir := int32(1024)
		if restante < escribir {
			escribir = restante
		}

		if _, err := file.Write(buffer[:escribir]); err != nil {
			return err
		}

		restante -= escribir
	}

	return nil
}

// crearSuperBloque crea e inicializa el SuperBloque
func crearSuperBloque(numeroInodos int32, numeroBloques int32, inicioParticion int32) structures.SuperBloque {
	var sb structures.SuperBloque

	sb.S_filesistem_type = 2 // EXT2
	sb.S_inodes_count = numeroInodos
	sb.S_blocks_count = numeroBloques
	sb.S_free_blocks_count = numeroBloques - 2 // -2 por carpeta raíz y users.txt
	sb.S_free_inodes_count = numeroInodos - 2  // -2 por inodo raíz y users.txt
	sb.S_mtime = utils.ObFechaInt()
	sb.S_umtime = 0
	sb.S_mnt_count = 1
	sb.S_magic = 0xEF53
	sb.S_inode_s = size.SizeTablaInodo()
	sb.S_block_s = size.SizeBloqueArchivo()
	sb.S_first_ino = 2 // Primer inodo libre (0 y 1 están usados)
	sb.S_first_blo = 2 // Primer bloque libre (0 y 1 están usados)

	// Calcular posiciones de las estructuras
	sb.S_bm_inode_start = inicioParticion + size.SizeSuperBloque()
	sb.S_bm_block_start = sb.S_bm_inode_start + numeroInodos
	sb.S_inode_start = sb.S_bm_block_start + numeroBloques
	sb.S_block_start = sb.S_inode_start + (numeroInodos * size.SizeTablaInodo())

	return sb
}

// inicializarBitmaps inicializa los bitmaps de inodos y bloques
func inicializarBitmaps(file *os.File, sb *structures.SuperBloque, numeroInodos int32, numeroBloques int32) error {
	var bit0 byte = '0'
	var bit1 byte = '1'

	// Inicializar bitmap de inodos (todos en 0)
	if _, err := file.Seek(int64(sb.S_bm_inode_start), 0); err != nil {
		return err
	}
	for i := int32(0); i < numeroInodos; i++ {
		if err := binary.Write(file, binary.LittleEndian, &bit0); err != nil {
			return err
		}
	}

	// Marcar primeros 2 inodos como usados (raíz y users.txt)
	if _, err := file.Seek(int64(sb.S_bm_inode_start), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &bit1); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &bit1); err != nil {
		return err
	}

	// Inicializar bitmap de bloques (todos en 0)
	if _, err := file.Seek(int64(sb.S_bm_block_start), 0); err != nil {
		return err
	}
	for i := int32(0); i < numeroBloques; i++ {
		if err := binary.Write(file, binary.LittleEndian, &bit0); err != nil {
			return err
		}
	}

	// Marcar primeros 2 bloques como usados (carpeta raíz y archivo users.txt)
	if _, err := file.Seek(int64(sb.S_bm_block_start), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &bit1); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &bit1); err != nil {
		return err
	}

	return nil
}

// crearInodoRaiz crea el inodo de la carpeta raíz
func crearInodoRaiz(sb *structures.SuperBloque) structures.TablaInodo {
	var inodo structures.TablaInodo

	inodo.I_uid = 1 // Usuario root
	inodo.I_gid = 1 // Grupo root
	inodo.I_s = 0   // Tamaño 0 para carpetas
	inodo.I_atime = utils.ObFechaInt()
	inodo.I_ctime = utils.ObFechaInt()
	inodo.I_mtime = utils.ObFechaInt()

	// Inicializar bloques en -1
	for i := range inodo.I_block {
		inodo.I_block[i] = -1
	}

	inodo.I_block[0] = sb.S_block_start // Primer bloque apunta al bloque carpeta
	inodo.I_type[0] = '0'               // '0' = Carpeta
	inodo.I_perm[0] = '6'               // Permisos 664
	inodo.I_perm[1] = '6'
	inodo.I_perm[2] = '4'

	return inodo
}

// crearBloqueCarpetaRaiz crea el bloque de carpeta raíz
func crearBloqueCarpetaRaiz(sb *structures.SuperBloque) structures.BloqueCarpeta {
	var bloque structures.BloqueCarpeta

	// Entrada 0: . (punto - referencia a sí mismo)
	copy(bloque.B_content[0].B_name[:], ".")
	bloque.B_content[0].B_inodo = sb.S_inode_start

	// Entrada 1: .. (punto punto - referencia al padre, en este caso sí mismo)
	copy(bloque.B_content[1].B_name[:], "..")
	bloque.B_content[1].B_inodo = sb.S_inode_start

	// Entrada 2: users.txt
	copy(bloque.B_content[2].B_name[:], "users.txt")
	bloque.B_content[2].B_inodo = sb.S_inode_start + size.SizeTablaInodo()

	// Entrada 3: vacía
	bloque.B_content[3].B_inodo = -1

	return bloque
}

// crearArchivoUsers crea el archivo users.txt con el contenido inicial
func crearArchivoUsers(file *os.File, sb *structures.SuperBloque) error {
	// Contenido inicial del archivo users.txt
	contenido := "1,G,root\n1,U,root,root,123\n"

	// Crear inodo para users.txt
	var inodoUsers structures.TablaInodo
	inodoUsers.I_uid = 1
	inodoUsers.I_gid = 1
	inodoUsers.I_s = int32(len(contenido))
	inodoUsers.I_atime = utils.ObFechaInt()
	inodoUsers.I_ctime = utils.ObFechaInt()
	inodoUsers.I_mtime = utils.ObFechaInt()

	// Inicializar bloques en -1
	for i := range inodoUsers.I_block {
		inodoUsers.I_block[i] = -1
	}

	inodoUsers.I_block[0] = sb.S_block_start + size.SizeBloqueCarpeta()
	inodoUsers.I_type[0] = '1' // '1' = Archivo
	inodoUsers.I_perm[0] = '6' // Permisos 664
	inodoUsers.I_perm[1] = '6'
	inodoUsers.I_perm[2] = '4'

	// Escribir inodo de users.txt
	if _, err := file.Seek(int64(sb.S_inode_start+size.SizeTablaInodo()), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &inodoUsers); err != nil {
		return err
	}

	// Crear bloque de archivo con el contenido
	var bloqueArchivo structures.BloqueArchivo
	copy(bloqueArchivo.B_content[:], contenido)

	// Escribir bloque de archivo
	if _, err := file.Seek(int64(sb.S_block_start+size.SizeBloqueCarpeta()), 0); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, &bloqueArchivo); err != nil {
		return err
	}

	return nil
}
