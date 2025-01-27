package snyk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIResponse struct {
	ID  string `json:"id"`
	Org string `json:"org"`
}

func MonitorImage(ctx context.Context, image, orgID, token string) error {
	url := fmt.Sprintf(#TODO, orgID)
	payload, err := json.Marshal(map[string]string{"image": image})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
