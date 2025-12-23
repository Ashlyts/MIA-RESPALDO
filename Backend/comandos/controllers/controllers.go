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
		Comando *string `json:"Comando"` // ← CORREGIDO: singular
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestBody); err != nil {
		json.NewEncoder(w).Encode(general.ResultadoSalida("JSON inválido o campos no permitidos", true, nil))
		return
	}

	if requestBody.Comando == nil || strings.TrimSpace(*requestBody.Comando) == "" {
		json.NewEncoder(w).Encode(general.ResultadoSalida("El campo 'Comando' es obligatorio y no puede ser nulo", true, nil))
		return
	}

	// Soportar múltiples comandos separados por salto de línea
	comandosLista := strings.Split(strings.TrimSpace(*requestBody.Comando), "\n")
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

	errores, salidas, _ := general.GlobalCom(comandosValidos)

	if len(errores) > 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida(errores[0], true, nil))
	} else {
		json.NewEncoder(w).Encode(general.ResultadoSalida("", false, salidas))
	}
}

/*func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var params struct {
		User string `json:"user"`
		Pass string `json:"pass"`
		ID   string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validar que no falten campos
	if params.User == "" || params.Pass == "" || params.ID == "" {
		http.Error(w, "Faltan parámetros: user, pass o id", http.StatusBadRequest)
		return
	}

	// Construir el comando completo
	cmd := fmt.Sprintf("login -user=%s -pass=%s -id=%s", params.User, params.Pass, params.ID)

	// Ejecutar usando el mismo flujo que /commands
	errores, salidas, _ := general.GlobalCom([]string{cmd})

	// Responder igual que HandleCommand
	if len(errores) > 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida(errores[0], true, nil))
	} else {
		json.NewEncoder(w).Encode(general.ResultadoSalida("", false, salidas))
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	errores, salidas, _ := general.GlobalCom([]string{"logout"})
	if len(errores) > 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida(errores[0], true, nil))
	} else {
		json.NewEncoder(w).Encode(general.ResultadoSalida("", false, salidas))
	}
}

func HandleCat(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID    string `json:"id"`
		File1 string `json:"file1"`
		File2 string `json:"file2"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	cmd := "cat"
	if params.ID != "" {
		cmd += " -id=" + params.ID
	}
	if params.File1 != "" {
		cmd += " -file1=" + params.File1
	}
	if params.File2 != "" {
		cmd += " -file2=" + params.File2
	}

	errores, salidas, _ := general.GlobalCom([]string{cmd})
	if len(errores) > 0 {
		json.NewEncoder(w).Encode(general.ResultadoSalida(errores[0], true, nil))
	} else {
		json.NewEncoder(w).Encode(general.ResultadoSalida("", false, salidas))
	}
}*/
