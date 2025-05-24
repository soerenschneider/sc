package healthcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type LogEntry struct {
	unix      int64
	Timestamp string `json:"_time"`
	Message   string `json:"_msg"`
}

func TransformLogs(data []LogEntry) ([]string, [][]string) {
	headers := []string{
		"Timestamp",
		"Message",
	}

	ret := make([][]string, 0, len(data))
	for _, logline := range data {
		ret = append(ret, []string{logline.Timestamp, logline.Message})
	}

	return headers, ret
}

func QueryVictorialogs(ctx context.Context, query, endpoint string) ([]LogEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("query", query)
	q.Set("limit", "10")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response: %s - %s", resp.Status, body)
	}

	body, _ := io.ReadAll(resp.Body)
	lines := bytes.Split(body, []byte("\n"))
	logs := make([]LogEntry, 0, len(lines))

	for _, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		t, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err == nil {
			entry.unix = t.Unix()
			entry.Timestamp = t.Format("01-02 15:04:05")
		}

		logs = append(logs, entry)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].unix < logs[j].unix
	})

	return logs, nil
}
