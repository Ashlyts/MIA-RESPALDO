package controllers

import (
	"Proyecto/comandos/general"
	"encoding/json"
	"net/http"
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

	// ✅ Ejecutar y obtener SALIDAS (no solo errores)
	_, todasSalidas, _ := general.GlobalCom(comandosValidos)

	// ✅ Devolver todas las salidas al frontend
	if err := json.NewEncoder(w).Encode(general.ResultadoSalida("", false, todasSalidas)); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("Error interno al generar respuesta", true, nil))
	}
}
