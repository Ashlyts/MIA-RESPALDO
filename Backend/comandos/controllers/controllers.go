package controllers

import (
	"Proyecto/comandos/general"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func HandleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	var requestBody struct {
		Comandos *string `json:"Comandos"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestBody); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("JSON inválido o campos no permitidos", true, nil))
		return
	}

	if requestBody.Comandos == nil || strings.TrimSpace(*requestBody.Comandos) == "" {
		json.NewEncoder(w).Encode(general.ResultadoSalida("El campo 'Comandos' es obligatorio y no puede ser nulo", true, nil))
		return
	}

	comandosLista := strings.Split(strings.TrimSpace(*requestBody.Comandos), "\n")
	var comandosValidos []string
	for _, cmd := range comandosLista {
		if trimmed := strings.TrimSpace(cmd); trimmed != "" {
			comandosValidos = append(comandosValidos, trimmed)
		}
	}

	if len(comandosValidos) == 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida("No hay comandos válidos para ejecutar", true, nil))
		return
	}

	_, todasSalidas, _ := general.GlobalCom(comandosValidos)

	//Devolver todas las salidas al frontend
	if err := json.NewEncoder(w).Encode(general.ResultadoSalida("", false, todasSalidas)); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("Error interno al generar respuesta", true, nil))
	}
}

func HandleReportsObtener(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	var requestBody struct {
		Comandos *string `json:"Comandos"` // O el nombre que uses en el frontend
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestBody); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("JSON inválido o campos no permitidos", true, nil))
		return
	}

	if requestBody.Comandos == nil || strings.TrimSpace(*requestBody.Comandos) == "" {
		json.NewEncoder(w).Encode(general.ResultadoSalida("El campo 'Comandos' es obligatorio y no puede ser nulo", true, nil))
		return
	}

	comandosLista := strings.Split(strings.TrimSpace(*requestBody.Comandos), "\n")
	var comandosRep []string // Solo comandos 'rep'
	for _, cmd := range comandosLista {
		trimmed := strings.TrimSpace(cmd)
		if trimmed != "" {
			// Verificar si es un comando 'rep'
			parts := strings.Fields(trimmed)
			if len(parts) > 0 && strings.ToLower(parts[0]) == "rep" {
				comandosRep = append(comandosRep, trimmed)
			} else if strings.HasPrefix(trimmed, "#") || trimmed == "" {
				// Añadir comentarios y líneas vacías también
				comandosRep = append(comandosRep, trimmed)
			}
			// Opcional: puedes manejar errores para comandos no 'rep' aquí si lo deseas
		}
	}

	if len(comandosRep) == 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida("No hay comandos 'rep' válidos para ejecutar", true, nil))
		return
	}

	_, todasSalidas, _ := general.GlobalCom(comandosRep)

	// Devolver todas las salidas al frontend
	if err := json.NewEncoder(w).Encode(general.ResultadoSalida("", false, todasSalidas)); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("Error interno al generar respuesta", true, nil))
	}
}

func HandleListReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	repDir := "VDIC-MIA/Rep"
	var reports []string

	if _, err := os.Stat(repDir); os.IsNotExist(err) {
		// Si no existe la carpeta, devolver lista vacía
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": false,
			"data":  []string{},
		})
		return
	}

	files, err := os.ReadDir(repDir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   true,
			"message": "Error al leer la carpeta de reportes",
			"data":    []string{},
		})
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			// Solo incluir archivos HTML o TXT (reportes)
			if strings.HasSuffix(strings.ToLower(name), ".html") || strings.HasSuffix(strings.ToLower(name), ".txt") {
				reports = append(reports, name)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": false,
		"data":  reports,
	})
}
