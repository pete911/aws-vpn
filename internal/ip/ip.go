package ip

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ipApiEndpoint = "http://checkip.amazonaws.com"

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func MyIp() (string, error) {
	resp, err := httpClient.Get(ipApiEndpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	return strings.TrimSpace(string(b)), nil
}
