package admonDisk

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Estructura para espacios libres
type EspacioLibre struct {
	Inicio  int32
	Tamanio int32
}

func FdiskExecute(comando string, parametros map[string]string) (string, bool) {
	tamanio, er, strError := utils.TieneSize(comando, parametros["size"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	unidad, er, strError := utils.TieneUnit(comando, parametros["unit"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	diskName, er, strError := utils.TieneDiskName(parametros["diskname"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	tipo, er, strError := utils.TieneType(parametros["type"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	fit, er, strError := utils.TieneFit("fdisk", parametros["fit"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	nombreParticion, er, strError := utils.TieneName(parametros["name"])
	if er {
		errMsg := fmt.Sprintf("[FDISK ERROR]: %s", strError)
		color.Red(errMsg)
		return errMsg, er
	}

	return fdiskCreate(tamanio, unidad, diskName, tipo, fit, nombreParticion)
}

func fdiskCreate(tamanio int32, unidad byte, diskName string, tipo byte, tipoFit byte, nombreParticion string) (string, bool) {
	var nombreSinExtension = ""
	var extensionArchivo = ""
	if strings.Contains(diskName, ".") {
		nombreSinExtension = strings.Split(diskName, ".")[0]
		extensionArchivo = strings.Split(diskName, ".")[1]
	}

	if strings.ToLower(extensionArchivo) != "mia" {
		msg := "Extensión del archivo no válida. Debe ser .mia"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	nombreCompleto := nombreSinExtension + ".mia"
	path := utils.DirectorioDisco + nombreCompleto

	switch tipo {
	case 'P':
		color.Cyan("→ Creando Partición Primaria...")
		return particionPrimaria(path, nombreParticion, tipo, tamanio, tipoFit, unidad)
	case 'E':
		color.Cyan("→ Creando Partición Extendida...")
		return particionExtendida(path, nombreParticion, tipo, tamanio, tipoFit, unidad)
	case 'L':
		color.Cyan("→ Creando Partición Lógica...")
		return particionLogica(path, nombreParticion, tamanio, tipoFit, unidad)
	default:
		msg := "Tipo de partición desconocido. Use P, E o L"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}
}

func particionPrimaria(ubicacionArchivo string, nombreParticion string, tipo byte, tamanioDisco int32, tipoFit byte, unidad byte) (string, bool) {
	if !utils.ExisteArchivo("FDISK", ubicacionArchivo) {
		msg := fmt.Sprintf("Disco no encontrado: %s", ubicacionArchivo)
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	mbr, er, strError := utils.ObtenerEstructuraMBR(ubicacionArchivo)
	if er {
		color.Red("[FDISK ERROR]: %s", strError)
		return strError, er
	}

	if existe, msg := utils.ExisteNombreParticion(ubicacionArchivo, nombreParticion); existe {
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	countParticiones := 0
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_s > 0 {
			countParticiones++
		}
	}

	if countParticiones >= 4 {
		msg := "No se pueden crear más particiones (máximo 4 primarias/extendidas)"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	tamanioBytes := utils.ObtenerTamanioDisco(tamanioDisco, unidad)
	if int64(tamanioBytes) <= 0 {
		msg := "Tamaño de partición inválido"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	espaciosLibres := encontrarEspaciosLibres(&mbr)
	espacioSeleccionado := aplicarAlgoritmoAjuste(espaciosLibres, tamanioBytes, tipoFit)
	if espacioSeleccionado == nil {
		msg := fmt.Sprintf("No hay espacio suficiente para crear partición de %d bytes", tamanioBytes)
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	posSlot := -1
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_s <= 0 {
			posSlot = i
			break
		}
	}

	if posSlot == -1 {
		msg := "No hay slots disponibles en el MBR"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	particion := utils.NuevaPartitionVacia()
	particion.Part_type = tipo
	particion.Part_fit = tipoFit
	copy(particion.Part_name[:], []byte(nombreParticion))
	particion.Part_s = tamanioBytes
	particion.Part_start = espacioSeleccionado.Inicio
	particion.Part_correlative = -1
	particion.Part_status = int8(-1)

	mbr.Mbr_partitions[posSlot] = particion

	file, err := os.OpenFile(ubicacionArchivo, os.O_RDWR, 0666)
	if err != nil {
		msg := "[FDISK]: Error al abrir el archivo para escritura"
		color.Red(msg)
		return msg, true
	}
	defer file.Close()

	if _, err := file.Seek(0, 0); err != nil {
		msg := "[FDISK]: Error al posicionar puntero"
		color.Red(msg)
		return msg, true
	}

	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		msg := "[FDISK]: Error al escribir MBR"
		color.Red(msg)
		return msg, true
	}

	llenarParticionConCeros(file, espacioSeleccionado.Inicio, tamanioBytes)

	detalles := fmt.Sprintf(`  Nombre:         %s
  Tipo:           Primaria (P)
  Inicio:         %d bytes
  Tamaño:         %d bytes (%.2f KB)
  Ajuste:         %s (%c)
  Slot MBR:       %d`,
		nombreParticion,
		espacioSeleccionado.Inicio,
		tamanioBytes,
		float64(tamanioBytes)/1024.0,
		map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}[tipoFit],
		tipoFit,
		posSlot)

	salida := utils.SuccessBanner("PARTICIÓN PRIMARIA CREADA EXITOSAMENTE", detalles)

	color.Green("===========================================================")
	color.Green("PARTICIÓN PRIMARIA CREADA EXITOSAMENTE")
	color.Green("===========================================================")
	color.Cyan("  Nombre:         %s", nombreParticion)
	color.Cyan("  Tipo:           Primaria (P)")
	color.Cyan("  Inicio:         %d bytes", espacioSeleccionado.Inicio)
	color.Cyan("  Tamaño:         %d bytes (%.2f KB)", tamanioBytes, float64(tamanioBytes)/1024.0)
	fitNombre := map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}
	color.Cyan("  Ajuste:         %s (%c)", fitNombre[tipoFit], tipoFit)
	color.Cyan("  Slot MBR:       %d", posSlot)
	color.Green("============================================================")

	return salida, false
}

func particionExtendida(ubicacionArchivo string, nombreParticion string, tipo byte, tamanioDisco int32, tipoFit byte, unidad byte) (string, bool) {
	if !utils.ExisteArchivo("FDISK", ubicacionArchivo) {
		msg := fmt.Sprintf("Disco no encontrado: %s", ubicacionArchivo)
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	mbr, er, strError := utils.ObtenerEstructuraMBR(ubicacionArchivo)
	if er {
		color.Red("[FDISK ERROR]: %s", strError)
		return strError, er
	}

	if existe, msg := utils.ExisteNombreParticion(ubicacionArchivo, nombreParticion); existe {
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	// Verificar que no exista ya una partición extendida
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_type == 'E' && mbr.Mbr_partitions[i].Part_s > 0 {
			msg := "Ya existe una partición extendida en el disco"
			color.Red("[FDISK ERROR]: %s", msg)
			return msg, true
		}
	}

	countParticiones := 0
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_s > 0 {
			countParticiones++
		}
	}

	if countParticiones >= 4 {
		msg := "No se pueden crear más particiones (máximo 4 primarias/extendidas)"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	tamanioBytes := utils.ObtenerTamanioDisco(tamanioDisco, unidad)
	if int64(tamanioBytes) <= 0 {
		msg := "Tamaño de partición inválido"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	espaciosLibres := encontrarEspaciosLibres(&mbr)
	espacioSeleccionado := aplicarAlgoritmoAjuste(espaciosLibres, tamanioBytes, tipoFit)
	if espacioSeleccionado == nil {
		msg := fmt.Sprintf("No hay espacio suficiente para crear partición de %d bytes", tamanioBytes)
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	posSlot := -1
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_s <= 0 {
			posSlot = i
			break
		}
	}

	if posSlot == -1 {
		msg := "No hay slots disponibles en el MBR"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	particion := utils.NuevaPartitionVacia()
	particion.Part_type = 'E'
	particion.Part_fit = tipoFit
	copy(particion.Part_name[:], []byte(nombreParticion))
	particion.Part_s = tamanioBytes
	particion.Part_start = espacioSeleccionado.Inicio
	particion.Part_correlative = -1
	particion.Part_status = int8(-1)

	mbr.Mbr_partitions[posSlot] = particion

	file, err := os.OpenFile(ubicacionArchivo, os.O_RDWR, 0666)
	if err != nil {
		msg := "[FDISK]: Error al abrir el archivo para escritura"
		color.Red(msg)
		return msg, true
	}
	defer file.Close()

	if _, err := file.Seek(0, 0); err != nil {
		msg := "[FDISK]: Error al posicionar puntero"
		color.Red(msg)
		return msg, true
	}

	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		msg := "[FDISK]: Error al escribir MBR"
		color.Red(msg)
		return msg, true
	}

	// Crear el primer EBR vacío al inicio de la partición extendida
	primerEBR := crearEBRVacio()
	if _, err := file.Seek(int64(espacioSeleccionado.Inicio), 0); err != nil {
		msg := "[FDISK]: Error al posicionar puntero para EBR"
		color.Red(msg)
		return msg, true
	}
	if err := binary.Write(file, binary.LittleEndian, &primerEBR); err != nil {
		msg := "[FDISK]: Error al escribir EBR inicial"
		color.Red(msg)
		return msg, true
	}

	detalles := fmt.Sprintf(`  Nombre:         %s
  Tipo:           Extendida (E)
  Inicio:         %d bytes
  Tamaño:         %d bytes (%.2f KB)
  Ajuste:         %s (%c)
  Slot MBR:       %d
  Nota:           Puede contener particiones lógicas`,
		nombreParticion,
		espacioSeleccionado.Inicio,
		tamanioBytes,
		float64(tamanioBytes)/1024.0,
		map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}[tipoFit],
		tipoFit,
		posSlot)

	salida := utils.SuccessBanner("PARTICIÓN EXTENDIDA CREADA EXITOSAMENTE", detalles)

	color.Green("==========================================================")
	color.Green("PARTICIÓN EXTENDIDA CREADA EXITOSAMENTE")
	color.Green("==========================================================")
	color.Cyan("  Nombre:         %s", nombreParticion)
	color.Cyan("  Tipo:           Extendida (E)")
	color.Cyan("  Inicio:         %d bytes", espacioSeleccionado.Inicio)
	color.Cyan("  Tamaño:         %d bytes (%.2f KB)", tamanioBytes, float64(tamanioBytes)/1024.0)
	fitNombre := map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}
	color.Cyan("  Ajuste:         %s (%c)", fitNombre[tipoFit], tipoFit)
	color.Cyan("  Slot MBR:       %d", posSlot)
	color.Yellow("  Nota:           Puede contener particiones lógicas")
	color.Green("==========================================================")

	// ✅ Devolver el banner formateado
	return salida, false
}

func particionLogica(ubicacionArchivo string, nombreParticion string, tamanioDisco int32, tipoFit byte, unidad byte) (string, bool) {
	if !utils.ExisteArchivo("FDISK", ubicacionArchivo) {
		msg := fmt.Sprintf("Disco no encontrado: %s", ubicacionArchivo)
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	mbr, er, strError := utils.ObtenerEstructuraMBR(ubicacionArchivo)
	if er {
		color.Red("[FDISK ERROR]: %s", strError)
		return strError, er
	}

	if existe, msg := utils.ExisteNombreParticion(ubicacionArchivo, nombreParticion); existe {
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	var particionExtendida *structures.Partition
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_type == 'E' && mbr.Mbr_partitions[i].Part_s > 0 {
			particionExtendida = &mbr.Mbr_partitions[i]
			break
		}
	}

	if particionExtendida == nil {
		msg := "No existe partición extendida. Créela primero con -type=E"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	tamanioBytes := utils.ObtenerTamanioDisco(tamanioDisco, unidad)
	if int64(tamanioBytes) <= 0 {
		msg := "Tamaño de partición inválido"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	// Abrir archivo
	file, err := os.OpenFile(ubicacionArchivo, os.O_RDWR, 0666)
	if err != nil {
		msg := "[FDISK]: Error al abrir el archivo"
		color.Red(msg)
		return msg, true
	}
	defer file.Close()

	// Encontrar espacio libre dentro de la extendida
	espaciosLibres := encontrarEspaciosLibresEnExtendida(file, particionExtendida)
	espacioSeleccionado := aplicarAlgoritmoAjusteLogica(espaciosLibres, tamanioBytes+size.SizeEBR(), tipoFit)

	if espacioSeleccionado == nil {
		msg := "No hay espacio suficiente en la partición extendida"
		color.Red("[FDISK ERROR]: %s", msg)
		return msg, true
	}

	// Crear nuevo EBR
	nuevoEBR := structures.EBR{}
	nuevoEBR.Part_mount = int8(0)
	nuevoEBR.Part_fit = tipoFit
	nuevoEBR.Part_start = espacioSeleccionado.Inicio + size.SizeEBR()
	nuevoEBR.Part_s = tamanioBytes
	nuevoEBR.Part_next = -1
	copy(nuevoEBR.Name[:], []byte(nombreParticion))

	// Escribir el nuevo EBR
	if _, err := file.Seek(int64(espacioSeleccionado.Inicio), 0); err != nil {
		msg := "[FDISK]: Error al posicionar puntero"
		color.Red(msg)
		return msg, true
	}
	if err := binary.Write(file, binary.LittleEndian, &nuevoEBR); err != nil {
		msg := "[FDISK]: Error al escribir EBR"
		color.Red(msg)
		return msg, true
	}

	// Si no es el primer EBR, actualizar el anterior
	if espacioSeleccionado.Inicio != particionExtendida.Part_start {
		actualizarEBRAnterior(file, particionExtendida, espacioSeleccionado.Inicio)
	}

	detalles := fmt.Sprintf(`  Nombre:         %s
  Tipo:           Lógica (L)
  Inicio:         %d bytes
  Tamaño:         %d bytes (%.2f KB)
  Ajuste:         %s (%c)`,
		nombreParticion,
		nuevoEBR.Part_start,
		tamanioBytes,
		float64(tamanioBytes)/1024.0,
		map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}[tipoFit],
		tipoFit)

	salida := utils.SuccessBanner("PARTICIÓN LÓGICA CREADA EXITOSAMENTE", detalles)

	color.Green("===========================================================")
	color.Green("PARTICIÓN LÓGICA CREADA EXITOSAMENTE")
	color.Green("===========================================================")
	color.Cyan("  Nombre:         %s", nombreParticion)
	color.Cyan("  Tipo:           Lógica (L)")
	color.Cyan("  Inicio:         %d bytes", nuevoEBR.Part_start)
	color.Cyan("  Tamaño:         %d bytes (%.2f KB)", tamanioBytes, float64(tamanioBytes)/1024.0)
	fitNombre := map[byte]string{'B': "Best Fit", 'F': "First Fit", 'W': "Worst Fit"}
	color.Cyan("  Ajuste:         %s (%c)", fitNombre[tipoFit], tipoFit)
	color.Green("===========================================================")

	return salida, false
}

func crearEBRVacio() structures.EBR {
	var ebr structures.EBR
	ebr.Part_mount = int8(0)
	ebr.Part_fit = 'W'
	ebr.Part_start = -1
	ebr.Part_s = 0
	ebr.Part_next = -1
	return ebr
}

func encontrarEspaciosLibresEnExtendida(file *os.File, extendida *structures.Partition) []EspacioLibre {
	espacios := []EspacioLibre{}
	inicioExtendida := extendida.Part_start
	finExtendida := extendida.Part_start + extendida.Part_s

	// Leer primer EBR
	var ebr structures.EBR
	if _, err := file.Seek(int64(inicioExtendida), 0); err != nil {
		return espacios
	}
	if err := binary.Read(file, binary.LittleEndian, &ebr); err != nil {
		return espacios
	}

	// Si el primer EBR está vacío, toda la extendida está libre
	if ebr.Part_s == 0 {
		espacios = append(espacios, EspacioLibre{
			Inicio:  inicioExtendida,
			Tamanio: extendida.Part_s,
		})
		return espacios
	}

	// Recolectar todas las particiones lógicas
	var logicas []structures.EBR
	logicas = append(logicas, ebr)

	for ebr.Part_next != -1 {
		if _, err := file.Seek(int64(ebr.Part_next), 0); err != nil {
			break
		}
		if err := binary.Read(file, binary.LittleEndian, &ebr); err != nil {
			break
		}
		logicas = append(logicas, ebr)
	}

	// Calcular espacios libres
	ultimoFin := inicioExtendida
	for _, logica := range logicas {
		inicioLogica := logica.Part_start - size.SizeEBR()
		if inicioLogica > ultimoFin {
			espacios = append(espacios, EspacioLibre{
				Inicio:  ultimoFin,
				Tamanio: inicioLogica - ultimoFin,
			})
		}
		ultimoFin = logica.Part_start + logica.Part_s
	}

	// Espacio después de la última lógica
	if ultimoFin < finExtendida {
		espacios = append(espacios, EspacioLibre{
			Inicio:  ultimoFin,
			Tamanio: finExtendida - ultimoFin,
		})
	}

	return espacios
}

func aplicarAlgoritmoAjusteLogica(espacios []EspacioLibre, tamanioRequerido int32, tipoFit byte) *EspacioLibre {
	switch tipoFit {
	case 'F':
		return firstFit(espacios, tamanioRequerido)
	case 'B':
		return bestFit(espacios, tamanioRequerido)
	case 'W':
		return worstFit(espacios, tamanioRequerido)
	default:
		return worstFit(espacios, tamanioRequerido)
	}
}

func actualizarEBRAnterior(file *os.File, extendida *structures.Partition, nuevaPosicion int32) {
	var ebr structures.EBR
	posicionActual := extendida.Part_start

	if _, err := file.Seek(int64(posicionActual), 0); err != nil {
		return
	}
	if err := binary.Read(file, binary.LittleEndian, &ebr); err != nil {
		return
	}

	// Navegar hasta el último EBR
	for ebr.Part_next != -1 {
		posicionActual = ebr.Part_next
		if _, err := file.Seek(int64(posicionActual), 0); err != nil {
			return
		}
		if err := binary.Read(file, binary.LittleEndian, &ebr); err != nil {
			return
		}
	}

	// Actualizar Part_next del último EBR
	ebr.Part_next = nuevaPosicion
	if _, err := file.Seek(int64(posicionActual), 0); err != nil {
		return
	}
	binary.Write(file, binary.LittleEndian, &ebr)
}

// Funciones auxiliares existentes

func encontrarEspaciosLibres(mbr *structures.MBR) []EspacioLibre {
	espacios := []EspacioLibre{}
	var particionesOrdenadas []structures.Partition

	for i := 0; i < 4; i++ {
		if mbr.Mbr_partitions[i].Part_s > 0 {
			particionesOrdenadas = append(particionesOrdenadas, mbr.Mbr_partitions[i])
		}
	}

	for i := 0; i < len(particionesOrdenadas)-1; i++ {
		for j := 0; j < len(particionesOrdenadas)-i-1; j++ {
			if particionesOrdenadas[j].Part_start > particionesOrdenadas[j+1].Part_start {
				particionesOrdenadas[j], particionesOrdenadas[j+1] = particionesOrdenadas[j+1], particionesOrdenadas[j]
			}
		}
	}

	mbrSize := size.SizeMBR()
	if len(particionesOrdenadas) == 0 {
		espacios = append(espacios, EspacioLibre{
			Inicio:  mbrSize,
			Tamanio: mbr.Mbr_tamano - mbrSize,
		})
		return espacios
	}

	if particionesOrdenadas[0].Part_start > mbrSize {
		espacios = append(espacios, EspacioLibre{
			Inicio:  mbrSize,
			Tamanio: particionesOrdenadas[0].Part_start - mbrSize,
		})
	}

	for i := 0; i < len(particionesOrdenadas)-1; i++ {
		finActual := particionesOrdenadas[i].Part_start + particionesOrdenadas[i].Part_s
		inicioSiguiente := particionesOrdenadas[i+1].Part_start

		if inicioSiguiente > finActual {
			espacios = append(espacios, EspacioLibre{
				Inicio:  finActual,
				Tamanio: inicioSiguiente - finActual,
			})
		}
	}

	ultimaParticion := particionesOrdenadas[len(particionesOrdenadas)-1]
	finUltima := ultimaParticion.Part_start + ultimaParticion.Part_s
	if finUltima < mbr.Mbr_tamano {
		espacios = append(espacios, EspacioLibre{
			Inicio:  finUltima,
			Tamanio: mbr.Mbr_tamano - finUltima,
		})
	}

	return espacios
}

func aplicarAlgoritmoAjuste(espacios []EspacioLibre, tamanioRequerido int32, tipoFit byte) *EspacioLibre {
	switch tipoFit {
	case 'F':
		return firstFit(espacios, tamanioRequerido)
	case 'B':
		return bestFit(espacios, tamanioRequerido)
	case 'W':
		return worstFit(espacios, tamanioRequerido)
	default:
		return worstFit(espacios, tamanioRequerido)
	}
}

func firstFit(espacios []EspacioLibre, tamanio int32) *EspacioLibre {
	for i := range espacios {
		if espacios[i].Tamanio >= tamanio {
			return &espacios[i]
		}
	}
	return nil
}

func bestFit(espacios []EspacioLibre, tamanio int32) *EspacioLibre {
	var mejor *EspacioLibre
	menorDiferencia := int32(0x7FFFFFFF)

	for i := range espacios {
		if espacios[i].Tamanio >= tamanio {
			diferencia := espacios[i].Tamanio - tamanio
			if diferencia < menorDiferencia {
				menorDiferencia = diferencia
				mejor = &espacios[i]
			}
		}
	}
	return mejor
}

func worstFit(espacios []EspacioLibre, tamanio int32) *EspacioLibre {
	var peor *EspacioLibre
	mayorTamanio := int32(0)

	for i := range espacios {
		if espacios[i].Tamanio >= tamanio && espacios[i].Tamanio > mayorTamanio {
			mayorTamanio = espacios[i].Tamanio
			peor = &espacios[i]
		}
	}
	return peor
}

func llenarParticionConCeros(file *os.File, inicio int32, tamanio int32) error {
	file.Seek(int64(inicio), 0)
	buffer := make([]byte, 1024)
	restante := tamanio

	for restante > 0 {
		escribir := int32(1024)
		if restante < escribir {
			escribir = restante
		}
		_, err := file.Write(buffer[:escribir])
		if err != nil {
			return err
		}
		restante -= escribir
	}
	return nil
}
