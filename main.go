package main

import (
	"log"
	"os"
)

func main() {
	executeAmanar()
}

func processVaultAddress(token string, ce AmanarConfigurationElement) {
	log.Printf("\n\n\n\n =========================== [VAULT ADDRESS %s] =========================== \n\n", ce.VaultAddress)

	ghc := &VaultGithubAuthClient{
		VaultAddress: ce.VaultAddress,
	}
	ghc.loginWithToken(token)

	for _, configItem := range ce.VaultConfiguration {
		secret, err := ghc.getCredential(configItem.VaultPath, configItem.VaultRole)
		if err != nil {
			log.Printf("[VAULT AUTH] Could not retrieve secret for vault path %s and vault role %s because %s. Skipping.", configItem.VaultPath, configItem.VaultRole, err)
			continue
		}

		credentials, err := CreateCredentialsFromSecret(secret)

		if err != nil {
			log.Printf("[VAULT AUTH] Could not convert Vault secret into Amanar credentials because %s. Skipping.", err)
			continue
		}

		log.Printf("[VAULT CONFIGURATION] %v:%v", configItem.VaultPath, configItem.VaultRole)
		ProcessConfigItem(&configItem.Configurables, credentials)
	}
}

//go:generate go-bindata amanar_config_schema.json
func executeAmanar() {
	configurationElements, err, resultErrors := LoadConfiguration(os.Getenv("CONFIG_FILEPATH"), "amanar_config_schema.json")

	if err != nil {
		log.Fatalf("[CONFIG] Could not load configuration file: %s", err)
		return
	}

	if resultErrors != nil {
		log.Println("[CONFIG SCHEMA] The provided configuration JSON did not conform to the structure required by the JSON Schema.")
		for _, resultErr := range resultErrors {
			log.Printf("[CONFIG SCHEMA] At JSON location %s: %s", resultErr.Context().String(), resultErr.Description())
		}
		return
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		log.Fatalln("[AUTH] Please provide a valid Vault token as the environment variable VAULT_TOKEN.")
		return
	}

	for _, configurationElement := range configurationElements {
		processVaultAddress(token, configurationElement)
	}
}
