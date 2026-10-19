package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  [2]interface{}    `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func QueryPrometheus(ctx context.Context, query, endpoint string) (map[string]float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req) //#nosec:G704
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response: %s - %s", resp.Status, body)
	}

	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, err
	}

	results := make(map[string]float64)
	for _, r := range promResp.Data.Result {
		instance := r.Metric["instance"]
		valueStr, ok := r.Value[1].(string)
		if !ok {
			continue
		}
		var val float64
		_, err := fmt.Sscanf(valueStr, "%f", &val)
		if err != nil {
			return nil, err
		}
		results[instance] = val
	}

	return results, nil
}
