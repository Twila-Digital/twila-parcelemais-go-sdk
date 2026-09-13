package parcelemais

// Environment seleciona o ambiente padrão da API do Parcele+ quando BaseURL não é informado.
type Environment string

const (
	EnvironmentStaging    Environment = "staging"
	EnvironmentProduction Environment = "production"
)

var environmentBaseURLs = map[Environment]string{
	EnvironmentStaging:    "https://api.staging.parcelemais.com.br/integration/",
	EnvironmentProduction: "https://api.parcelemais.com.br/integration/",
}
