package platform

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Multi-provider model integration.
//
// All three local services speak the OpenAI-compatible protocol:
//   - Chat/Reasoning : POST {base}/chat/completions  (qwen3.6-35b-a3b)
//   - Embedding      : POST {base}/embeddings        (bge-m3, 1024 dims)
//   - OCR            : POST {base}/chat/completions  (deepseek-ocr, vision)
//
// Every provider degrades gracefully: chat falls back to the rule engine when
// misconfigured, embeddings fall back to the built-in hashed sparse vector,
// and OCR reports a clear error to the caller.
// ---------------------------------------------------------------------------

// modelConf is one OpenAI-compatible endpoint's resolved settings. Values come
// from the database settings table first, then process/.env variables.
type modelConf struct {
	BaseURL string
	Model   string
	APIKey  string
	Dim     int
}

// modelSetting resolves a single key: DB setting > environment > default.
func (a *App) modelSetting(key, envKey, def string) string {
	if v := a.setting(key); v != "" {
		return v
	}
	return env(envKey, def)
}

func (a *App) chatConf() modelConf {
	return modelConf{
		BaseURL: strings.TrimRight(a.modelSetting("llm.base_url", "LLM_BASE_URL", "https://api.deepseek.com/v1"), "/"),
		Model:   a.modelSetting("llm.model", "LLM_MODEL", "deepseek-chat"),
		APIKey:  a.modelSetting("llm.api_key", "LLM_API_KEY", env("DEEPSEEK_API_KEY", "")),
	}
}

func (a *App) embeddingConf() modelConf {
	return modelConf{
		BaseURL: strings.TrimRight(a.modelSetting("embedding.base_url", "EMBEDDING_BASE_URL", ""), "/"),
		Model:   a.modelSetting("embedding.model", "EMBEDDING_MODEL", ""),
		APIKey:  a.modelSetting("embedding.api_key", "EMBEDDING_API_KEY", ""),
		Dim:     int(num(a.modelSetting("embedding.dim", "EMBEDDING_DIM", "1024"))),
	}
}

func (a *App) ocrConf() modelConf {
	return modelConf{
		BaseURL: strings.TrimRight(a.modelSetting("ocr.base_url", "OCR_BASE_URL", ""), "/"),
		Model:   a.modelSetting("ocr.model", "OCR_MODEL", ""),
		APIKey:  a.modelSetting("ocr.api_key", "OCR_API_KEY", ""),
	}
}

// modelHTTPClient bypasses any system HTTP proxy. Local/private model
// endpoints (10.x.x.x) must be reached directly; going through a corporate
// proxy returns an HTML error page instead of the model's JSON.
var modelHTTPClient = &http.Client{Transport: &http.Transport{Proxy: nil}}

// modelPost performs a generic POST to an OpenAI-compatible endpoint and
// returns the decoded JSON object. HTTP/transport errors are surfaced as
// DSH-style apiError values so route handlers get consistent status codes.
func modelPost(ctx context.Context, conf modelConf, path string, body any, timeout time.Duration) M {
	if conf.BaseURL == "" {
		fail(422, "模型服务地址未配置")
	}
	raw, e := json.Marshal(body)
	if e != nil {
		fail(500, "请求序列化失败: "+e.Error())
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, e := http.NewRequestWithContext(cctx, http.MethodPost, conf.BaseURL+path, bytes.NewReader(raw))
	if e != nil {
		fail(500, "模型请求构造失败: "+e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if conf.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+conf.APIKey)
	}
	resp, e := modelHTTPClient.Do(req)
	if e != nil {
		fail(502, "模型服务连接失败: "+e.Error())
	}
	defer resp.Body.Close()
	data, e := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if e != nil {
		fail(502, "读取模型响应失败: "+e.Error())
	}
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 500 {
			msg = msg[:500]
		}
		fail(502, fmt.Sprintf("模型服务返回 HTTP %d: %s", resp.StatusCode, msg))
	}
	var out M
	if e = json.Unmarshal(data, &out); e != nil {
		fail(502, "模型响应不是有效 JSON: "+e.Error())
	}
	return out
}

// ---------------------------------------------------------------------------
// Embedding (bge-m3)
// ---------------------------------------------------------------------------

// embedTexts returns one dense vector per input string. Returns nil when the
// embedding endpoint is not configured, so callers can fall back to the local
// hash vector. Errors on a configured-but-failing endpoint are surfaced so the
// caller can decide whether to degrade or fail.
func (a *App) embedTexts(ctx context.Context, texts []string) [][]float64 {
	conf := a.embeddingConf()
	if conf.BaseURL == "" || conf.Model == "" || len(texts) == 0 {
		return nil
	}
	out := modelPost(ctx, conf, "/embeddings", M{"model": conf.Model, "input": texts}, 60*time.Second)
	items := arr(out["data"])
	if len(items) != len(texts) {
		fail(502, fmt.Sprintf("Embedding 返回数量不匹配: 期望 %d 实际 %d", len(texts), len(items)))
	}
	// Sort by index to match input order regardless of response ordering.
	vecs := make([][]float64, len(texts))
	for _, raw := range items {
		item := obj(raw)
		idx := int(num(item["index"]))
		if idx < 0 || idx >= len(vecs) {
			fail(502, "Embedding 返回了非法的 index")
		}
		floats := arr(item["embedding"])
		vec := make([]float64, 0, len(floats))
		for _, f := range floats {
			vec = append(vec, num(f))
		}
		vecs[idx] = vec
	}
	for i, v := range vecs {
		if len(v) == 0 {
			fail(502, fmt.Sprintf("Embedding 第 %d 条为空", i))
		}
	}
	return vecs
}

// denseCosine computes the cosine similarity of two dense vectors.
func denseCosine(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// ---------------------------------------------------------------------------
// OCR (deepseek-ocr) — OpenAI-compatible vision chat completion
// ---------------------------------------------------------------------------

// ocrImage sends an image (local path or data URL) to the OCR model and returns
// the recognised markdown/plain text.
func (a *App) ocrImage(ctx context.Context, image, prompt string) string {
	conf := a.ocrConf()
	if conf.BaseURL == "" || conf.Model == "" {
		fail(422, "OCR 服务未配置")
	}
	dataURL, e := asDataURL(image)
	if e != nil {
		fail(422, "图片读取失败: "+e.Error())
	}
	if prompt == "" {
		prompt = "请识别图片中的全部文字，保持原有结构与排版，仅输出识别结果。"
	}
	content := []any{
		M{"type": "text", "text": prompt},
		M{"type": "image_url", "image_url": M{"url": dataURL}},
	}
	body := M{
		"model":      conf.Model,
		"messages":   []any{M{"role": "user", "content": content}},
		"max_tokens": 4096,
	}
	out := modelPost(ctx, conf, "/chat/completions", body, 180*time.Second)
	choices := arr(out["choices"])
	if len(choices) == 0 {
		fail(502, "OCR 服务没有返回结果")
	}
	msg := obj(obj(choices[0])["message"])
	text := str(msg["content"])
	if text == "" {
		// Some reasoning OCR models emit content in the reasoning field.
		text = str(msg["reasoning"])
	}
	return text
}

// asDataURL converts a local path, http(s) URL, raw base64, or an existing
// data URL into a data URL suitable for the vision API.
func asDataURL(image string) (string, error) {
	s := strings.TrimSpace(image)
	if s == "" {
		return "", fmt.Errorf("图片内容为空")
	}
	if strings.HasPrefix(s, "data:") {
		return s, nil
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s, nil // vision endpoints generally accept remote URLs too
	}
	if strings.HasPrefix(s, "base64,") {
		return "data:image/png;base64," + strings.TrimPrefix(s, "base64,"), nil
	}
	// Treat as raw base64 if it decodes cleanly and is long enough.
	if len(s) > 128 {
		if _, e := base64.StdEncoding.DecodeString(s); e == nil {
			return "data:image/png;base64," + s, nil
		}
	}
	// Otherwise read as a local file path.
	clean := filepath.Clean(s)
	data, e := os.ReadFile(clean)
	if e != nil {
		return "", e
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(clean)))
	if mimeType == "" {
		mimeType = "image/png"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
