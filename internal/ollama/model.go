package ollama

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type Model struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
}

type RunningModel struct {
	Name          string       `json:"name"`
	Model         string       `json:"model"`
	Size          int64        `json:"size"`
	SizeVRAM      int64        `json:"size_vram"`
	Digest        string       `json:"digest"`
	Details       ModelDetails `json:"details"`
	ExpiresAt     time.Time    `json:"expires_at"`
	ContextLength int          `json:"context_length"`
}

type ModelInfo struct {
	Modelfile    string         `json:"modelfile"`
	Parameters   string         `json:"parameters"`
	Template     string         `json:"template"`
	System       string         `json:"system"`
	License      string         `json:"license"`
	Details      ModelDetails   `json:"details"`
	ModelInfo    map[string]any `json:"model_info"`
	Capabilities []string       `json:"capabilities"`
	ModifiedAt   time.Time      `json:"modified_at"`
}

type Message struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	// Think enables or disables the reasoning of thinking models. If nil,
	// the default of the model applies.
	Think *bool `json:"think,omitempty"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
}

func shortDigest(digest string) string {
	digest = strings.TrimPrefix(digest, "sha256:")
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

func ModelsAsTable(models []Model) ([]string, [][]string) {
	headers := []string{"Name", "Size", "Params", "Quant", "Family", "ID", "Modified"}

	sort.Slice(models, func(i, j int) bool { return models[i].Name < models[j].Name })
	rows := make([][]string, 0, len(models))
	for _, m := range models {
		rows = append(rows, []string{
			m.Name,
			humanize.Bytes(uint64(max(m.Size, 0))),
			m.Details.ParameterSize,
			m.Details.QuantizationLevel,
			m.Details.Family,
			shortDigest(m.Digest),
			humanize.Time(m.ModifiedAt),
		})
	}
	return headers, rows
}

func RunningModelsAsTable(models []RunningModel) ([]string, [][]string) {
	headers := []string{"Name", "Size", "Processor", "Context", "Until"}

	rows := make([][]string, 0, len(models))
	for _, m := range models {
		rows = append(rows, []string{
			m.Name,
			humanize.Bytes(uint64(max(m.Size, 0))),
			m.Processor(),
			fmt.Sprint(m.ContextLength),
			humanize.Time(m.ExpiresAt),
		})
	}
	return headers, rows
}

// Processor describes how the model is split between CPU and GPU, like
// "ollama ps" does.
func (m RunningModel) Processor() string {
	switch {
	case m.Size <= 0 || m.SizeVRAM <= 0:
		return "100% CPU"
	case m.SizeVRAM >= m.Size:
		return "100% GPU"
	}
	gpu := m.SizeVRAM * 100 / m.Size
	return fmt.Sprintf("%d%%/%d%% CPU/GPU", 100-gpu, gpu)
}

// Describe renders the details of a model as text.
func (i ModelInfo) Describe(name string) string {
	var b strings.Builder

	field := func(key, value string) {
		if value != "" {
			fmt.Fprintf(&b, "  %-16s %s\n", key, value)
		}
	}

	b.WriteString("Model\n")
	field("name", name)
	field("family", i.Details.Family)
	field("parameters", i.Details.ParameterSize)
	field("quantization", i.Details.QuantizationLevel)
	field("format", i.Details.Format)
	if arch, ok := i.ModelInfo["general.architecture"].(string); ok {
		field("architecture", arch)
		if ctx, ok := i.ModelInfo[arch+".context_length"].(float64); ok {
			field("context length", fmt.Sprintf("%.0f", ctx))
		}
		if dim, ok := i.ModelInfo[arch+".embedding_length"].(float64); ok {
			field("embedding length", fmt.Sprintf("%.0f", dim))
		}
	}
	if !i.ModifiedAt.IsZero() {
		field("modified", humanize.Time(i.ModifiedAt))
	}

	if len(i.Capabilities) > 0 {
		b.WriteString("\nCapabilities\n")
		for _, c := range i.Capabilities {
			fmt.Fprintf(&b, "  %s\n", c)
		}
	}

	section := func(title, content string) {
		content = strings.TrimSpace(content)
		if content == "" {
			return
		}
		fmt.Fprintf(&b, "\n%s\n", title)
		for _, line := range strings.Split(content, "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	section("Parameters", i.Parameters)
	section("System", i.System)
	section("Template", i.Template)
	section("License", i.License)

	return b.String()
}
