package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func main() {
	client, err := parcelemais.NewClient(parcelemais.ClientOptions{
		ClientID:     os.Getenv("PARCELEMAIS_CLIENT_ID"),
		ClientSecret: os.Getenv("PARCELEMAIS_CLIENT_SECRET"),
		Environment:  parcelemais.EnvironmentStaging,
	})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/simulacoes", func(w http.ResponseWriter, r *http.Request) {
		parcelas, err := client.Simulations.SimulateInstallments(r.Context(), parcelemais.SimulateInstallmentsRequest{
			RequestedAmount: 1500.0,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"erro": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(parcelas)
	})

	log.Println("Ouvindo em :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
