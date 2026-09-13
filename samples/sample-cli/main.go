package main

import (
	"context"
	"fmt"
	"os"

	parcelemais "github.com/Twila-Digital/twila-parcelemais-go-sdk"
)

func main() {
	clientID := os.Getenv("PARCELEMAIS_CLIENT_ID")
	clientSecret := os.Getenv("PARCELEMAIS_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Defina PARCELEMAIS_CLIENT_ID e PARCELEMAIS_CLIENT_SECRET no ambiente.")
		os.Exit(1)
	}

	client, err := parcelemais.NewClient(parcelemais.ClientOptions{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Environment:  parcelemais.EnvironmentStaging,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	parcelas, err := client.Simulations.SimulateInstallments(context.Background(), parcelemais.SimulateInstallmentsRequest{
		RequestedAmount: 1500.0,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, parcela := range parcelas {
		fmt.Printf("%dx de %.2f (total %.2f)\n", parcela.Term, parcela.InstallmentAmount, parcela.TotalAmount)
	}
}
