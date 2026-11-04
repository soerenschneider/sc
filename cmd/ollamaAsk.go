package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/huh/v2"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/sc/internal/ollama"
	"github.com/soerenschneider/sc/internal/tui"
	"github.com/soerenschneider/sc/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	ollamaModel    = "model"
	ollamaContinue = "continue"
	ollamaSystem   = "system"
	ollamaThink    = "think"
	ollamaRaw      = "raw"

	// ollamaMaxWidth caps the width rendered markdown is wrapped at.
	ollamaMaxWidth = 100

	// ollamaMaxInput caps the input read from stdin.
	ollamaMaxInput = 4 << 20
)

var ollamaAskCmd = &cobra.Command{
	Use:     "ask [question]...",
	Aliases: []string{"a", "q"},
	Short:   "Ask a model a quick question",
	Long: `Ask a model a question and stream the answer to stdout. Input piped to the
command is passed to the model along with the question, or is the question
itself if no other is given. Without either, the question is read in a form
that accepts multiple lines and pasted text: enter starts a new line, ctrl+d
sends the question and ctrl+e opens $EDITOR.

The model is given by --model, which is best set in a profile. Without it, the
model of the continued conversation is used or, in a terminal, picked
interactively.

Every exchange is saved, and --continue asks a follow-up question in the last
conversation. Without --continue, a new conversation is started.

In a terminal, the markdown of the answer is rendered as it streams, one
paragraph, list or code block at a time. --raw prints the markdown as is,
which is also what is printed if the output is piped.

Thinking models use their default unless --think is given. With --think, the
reasoning is printed to stderr, --think=false turns it off for models that
support it.

Examples:
  sc ollama ask -m qwen3:8b "what's the difference between SIGTERM and SIGKILL"
  sc ollama ask -c "and SIGHUP?"
  git diff --staged | sc ollama ask "write a conventional commit message"
  kubectl logs pod/foo | sc ollama ask "why is this crashing"`,
	Run: func(cmd *cobra.Command, args []string) {
		prompt, err := ollamaAskPrompt(args)
		if errors.Is(err, errOllamaNoQuestion) && term.IsTerminal(int(os.Stdin.Fd())) {
			prompt, err = tui.ReadMultilineInput(cmd.Context(), "Question", "Type or paste your question")
			if errors.Is(err, huh.ErrUserAborted) {
				return
			}
		}
		if err != nil {
			log.Fatal().Err(err).Msg("could not read the question")
		}

		convFile, err := ollama.DefaultConversationFile()
		if err != nil {
			log.Fatal().Err(err).Msg("could not determine conversation file")
		}

		conv := &ollama.Conversation{}
		if pkg.MustGetBool(cmd, ollamaContinue) {
			conv, err = ollama.LoadConversation(convFile)
			if errors.Is(err, os.ErrNotExist) {
				log.Fatal().Msg("No conversation to continue, ask without --continue first")
			}
			if err != nil {
				log.Fatal().Err(err).Msg("could not load conversation")
			}
			if cmd.Flags().Changed(ollamaSystem) {
				log.Warn().Msg("Ignoring the system prompt, the conversation keeps its own")
			}
		} else if system := pkg.GetString(cmd, ollamaSystem); system != "" {
			conv.Messages = append(conv.Messages, ollama.Message{Role: "system", Content: system})
		}

		client := mustOllamaClient(cmd)
		if model := pkg.GetString(cmd, ollamaModel); model != "" {
			conv.Model = model
		}
		if conv.Model == "" {
			conv.Model = pickOllamaModels(cmd.Context(), client, "Model", false)[0]
		}
		conv.Messages = append(conv.Messages, ollama.Message{Role: "user", Content: prompt})

		req := ollama.ChatRequest{Model: conv.Model, Messages: conv.Messages}
		showThinking := false
		if cmd.Flags().Changed(ollamaThink) {
			think := pkg.MustGetBool(cmd, ollamaThink)
			req.Think = &think
			showThinking = think
		}

		out := newOllamaAnswerPrinter(showThinking, !pkg.MustGetBool(cmd, ollamaRaw))
		answer, err := client.Chat(cmd.Context(), req, out.write)
		out.done()
		if err != nil {
			if cmd.Context().Err() != nil {
				os.Exit(130)
			}
			log.Fatal().Err(err).Msgf("could not ask %s", conv.Model)
		}

		// the reasoning is not needed to continue the conversation
		answer.Thinking = ""
		conv.Messages = append(conv.Messages, answer)
		if err := conv.Save(convFile); err != nil {
			log.Warn().Err(err).Msg("could not save conversation")
		}
	},
}

func init() {
	ollamaCmd.AddCommand(ollamaAskCmd)

	ollamaAskCmd.Flags().StringP(ollamaModel, "m", "", "Model to ask")
	ollamaAskCmd.Flags().BoolP(ollamaContinue, "c", false, "Continue the last conversation")
	ollamaAskCmd.Flags().StringP(ollamaSystem, "s", "", "System prompt for a new conversation")
	ollamaAskCmd.Flags().Bool(ollamaRaw, false, "Print the answer as raw markdown")
	ollamaAskCmd.Flags().Bool(ollamaThink, false, "Enable thinking and print the reasoning to stderr, or disable it with --think=false")
}

var errOllamaNoQuestion = errors.New("no question given")

// ollamaAskPrompt builds the prompt from the arguments and piped input.
func ollamaAskPrompt(args []string) (string, error) {
	question := strings.TrimSpace(strings.Join(args, " "))

	var input string
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, ollamaMaxInput+1))
		if err != nil {
			return "", err
		}
		if len(data) > ollamaMaxInput {
			return "", fmt.Errorf("input exceeds %d MiB", ollamaMaxInput>>20)
		}
		input = strings.TrimSpace(string(data))
	}

	switch {
	case question == "" && input == "":
		return "", errOllamaNoQuestion
	case input == "":
		return question, nil
	case question == "":
		return input, nil
	}
	return question + "\n\n" + input, nil
}

// ollamaAnswerPrinter streams the answer to stdout, rendering its markdown if
// stdout is a terminal, and, if enabled, the reasoning dimmed to stderr.
type ollamaAnswerPrinter struct {
	showThinking bool
	dim          bool
	thinking     bool
	markdown     *tui.MarkdownStream
	lastByte     byte
}

func newOllamaAnswerPrinter(showThinking, renderMarkdown bool) *ollamaAnswerPrinter {
	p := &ollamaAnswerPrinter{showThinking: showThinking, dim: term.IsTerminal(int(os.Stderr.Fd()))}

	stdout := int(os.Stdout.Fd())
	if renderMarkdown && term.IsTerminal(stdout) {
		width := ollamaMaxWidth
		if w, _, err := term.GetSize(stdout); err == nil && w > 0 {
			width = min(w, ollamaMaxWidth)
		}
		md, err := tui.NewMarkdownStream(os.Stdout, width)
		if err != nil {
			log.Warn().Err(err).Msg("could not render markdown, printing it raw")
		}
		p.markdown = md
	}
	return p
}

func (p *ollamaAnswerPrinter) write(msg ollama.Message) {
	if p.showThinking && msg.Thinking != "" {
		p.thinking = true
		if p.dim {
			_, _ = fmt.Fprint(os.Stderr, "\033[2m"+msg.Thinking+"\033[0m")
		} else {
			_, _ = fmt.Fprint(os.Stderr, msg.Thinking)
		}
	}
	if msg.Content == "" {
		return
	}
	if p.thinking {
		p.thinking = false
		_, _ = fmt.Fprint(os.Stderr, "\n\n")
	}
	if p.markdown != nil {
		_ = p.markdown.Write(msg.Content)
		p.lastByte = ' '
		return
	}
	_, _ = fmt.Fprint(os.Stdout, msg.Content)
	p.lastByte = msg.Content[len(msg.Content)-1]
}

func (p *ollamaAnswerPrinter) done() {
	if p.thinking {
		_, _ = fmt.Fprintln(os.Stderr)
	}
	if p.markdown != nil {
		_ = p.markdown.Close()
	}
	if p.lastByte != 0 && p.lastByte != '\n' {
		_, _ = fmt.Fprintln(os.Stdout)
	}
}
