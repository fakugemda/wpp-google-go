package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	contactsKey    = "contacts"
	DefaultTimeout = 10 * time.Second
)

type upstashResp struct {
	Result interface{} `json:"result"`
}

func exec(ctx context.Context, cmd []interface{}) (interface{}, error) {
	url := os.Getenv("UPSTASH_REDIS_REST_URL")
	token := os.Getenv("UPSTASH_REDIS_REST_TOKEN")
	if url == "" || token == "" {
		return nil, fmt.Errorf("UPSTASH_REDIS_REST_URL o UPSTASH_REDIS_REST_TOKEN no configurados")
	}
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	body, _ := json.Marshal(cmd)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: DefaultTimeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("redis rest: %d - %s", resp.StatusCode, string(b))
	}
	var out upstashResp
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

func GetContacts(ctx context.Context) (map[string]string, error) {
	result, err := exec(ctx, []interface{}{"HGETALL", contactsKey})
	if err != nil {
		return nil, err
	}
	arr, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("redis: HGETALL devolvió %T", result)
	}
	contacts := make(map[string]string)
	for i := 0; i+1 < len(arr); i += 2 {
		if name, _ := arr[i].(string); name != "" {
			contacts[name], _ = arr[i+1].(string)
		}
	}
	return contacts, nil
}

func SetContact(ctx context.Context, name, email string) error {
	_, err := exec(ctx, []interface{}{"HSET", contactsKey, name, email})
	return err
}

func Ping(ctx context.Context) error {
	_, err := exec(ctx, []interface{}{"PING"})
	return err
}

func StartKeepAlive(interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Hour
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
			err := Ping(ctx)
			cancel()

			if err != nil {
				fmt.Printf("⚠️ KeepAlive Redis falló: %v\n", err)
				continue
			}

			fmt.Println("✅ KeepAlive Redis: PING enviado")
		}
	}()
}
