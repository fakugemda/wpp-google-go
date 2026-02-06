package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// GetClient - Obtiene un cliente HTTP autenticado
func GetClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromEnvOrFile("GOOGLE_TOKEN", "resources/token.json")
	if err != nil {
		tok = getTokenFromWeb(config)
		if err := saveToken("resources/token.json", tok); err != nil {
			fmt.Printf("No se pudo guardar token localmente: %s\n", err.Error())
		}
	}
	return config.Client(context.Background(), tok)
}

// tokenFromEnvOrFile - Intenta leer el token desde variable de entorno o archivo local
func tokenFromEnvOrFile(envName, fileName string) (*oauth2.Token, error) {
	envContent := os.Getenv(envName)
	if envContent != "" {
		fmt.Println("Token cargado desde Variable de Entorno.")
		tok := &oauth2.Token{}
		err := json.Unmarshal([]byte(envContent), tok)
		return tok, err
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fmt.Println("Token cargado desde archivo local.")
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// getTokenFromWeb - Obtiene un token desde la web (solo funciona en local)
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	var authCode string
	fmt.Print("Paste the authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %s", err.Error())
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to exchange authorization code: %s", err.Error())
	}

	return tok
}

// saveToken - Guarda el token en un archivo local
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving token in %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error abriendo archivo: %s", err.Error())
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("error codificando token: %s", err.Error())
	}
	return nil
}
