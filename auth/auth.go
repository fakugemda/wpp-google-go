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

// GetClient: Obtiene un cliente HTTP autenticado
func GetClient(config *oauth2.Config) *http.Client {
	// 1. Intentamos cargar el token (Nube o Local)
	tok, err := tokenFromEnvOrFile("GOOGLE_TOKEN", "resources/token.json")
	if err != nil {
		// Si no encontramos token por ningún lado, hay que loguearse.
		// OJO: Esto solo funciona en TU PC. En la nube fallará si no configuraste la variable.
		tok = getTokenFromWeb(config)
		saveToken("resources/token.json", tok) // Guardamos copia local por si acaso
	}
	return config.Client(context.Background(), tok)
}

// tokenFromEnvOrFile: La lógica "anfibia" para el token
func tokenFromEnvOrFile(envName, fileName string) (*oauth2.Token, error) {
	tok := &oauth2.Token{}

	// A. Intentar leer desde Variable de Entorno (Nube)
	envContent := os.Getenv(envName)
	if envContent != "" {
		fmt.Println("☁️ Token cargado desde Variable de Entorno.")
		err := json.Unmarshal([]byte(envContent), tok)
		return tok, err
	}

	// B. Intentar leer desde Archivo (Local)
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fmt.Println("🏠 Token cargado desde archivo local.")
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// getTokenFromWeb: Obtiene un token desde la web (solo funciona en local)
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	var authCode string
	fmt.Print("Paste the authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %s", err.Error())
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to exchange authorization code: %s", err.Error())
	}

	return tok
}

// saveToken: Guarda el token en un archivo local
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving token in %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Error saving token: %s", err.Error())
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
