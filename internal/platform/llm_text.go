package platform

import (
	"context"
	"strings"
	"time"
)

// llmText asks the configured chat model to produce a short piece of narrative
// text (answer / conclusion / summary / advice). It is deliberately defensive:
// when the model is not configured or anything fails, it returns the supplied
// fallback so every module keeps working exactly as before. The model is told
// to answer in one short paragraph and to avoid repeating raw numbers verbatim
// unless they are part of the sentence.
func (a *App) llmText(ctx context.Context, system, user, fallback string) string {
	if a.llm("api_key", "") == "" || a.chatConf().BaseURL == "" {
		return fallback
	}
	if strings.TrimSpace(user) == "" {
		return fallback
	}
	if ctx == nil {
		ctx = context.Background()
	}
	body := M{
		"model": a.chatConf().Model,
		"messages": []any{
			M{"role": "system", "content": system},
			M{"role": "user", "content": user},
		},
		"temperature": 0.3,
	}
	if maxTokens := int(num(a.modelSetting("llm.max_tokens", "LLM_MAX_TOKENS", "0"))); maxTokens > 0 {
		body["max_tokens"] = maxTokens
	}
	if strings.EqualFold(strings.TrimSpace(a.modelSetting("llm.enable_thinking", "LLM_ENABLE_THINKING", "")), "false") {
		body["chat_template_kwargs"] = M{"enable_thinking": false}
	}
	text := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Model failure must never break a business module.
				text = ""
			}
		}()
		conf := a.chatConf()
		out := modelPost(ctx, conf, "/chat/completions", body, 90*time.Second)
		choices := arr(out["choices"])
		if len(choices) == 0 {
			return
		}
		msg := obj(obj(choices[0])["message"])
		text = strings.TrimSpace(str(msg["content"]))
		if text == "" {
			text = strings.TrimSpace(str(msg["reasoning"]))
		}
	}()
	if text == "" {
		return fallback
	}
	// Reasoning models sometimes wrap the answer in markdown code fences.
	text = strings.TrimPrefix(text, "```markdown")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

// llmTextBackground is a convenience wrapper for callers without a request.
func (a *App) llmTextBackground(system, user, fallback string) string {
	return a.llmText(context.Background(), system, user, fallback)
}
