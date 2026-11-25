package platform



// modelRoutes exposes configuration + live test endpoints for the embedding
// and OCR providers, and an OCR-to-knowledge-document ingest helper.
func (a *App) modelRoutes() []route {
	return []route{
		// ---- Embedding ----
		{"POST", "/api/settings/embedding/test", false, true, func(c *request) any {
			c.admin()
			conf := a.embeddingConf()
			vecs := a.embedTexts(c.R.Context(), []string{"DataAgent 语义检索测试", "semantic search test"})
			if len(vecs) == 0 {
				fail(422, "Embedding 服务未配置")
			}
			return M{"ok": true, "model": conf.Model, "base_url": conf.BaseURL, "dim": len(vecs[0]), "configured": conf.BaseURL != "" && conf.Model != ""}
		}},
		{"POST", "/api/embedding", false, false, func(c *request) any {
			inputs := arr(c.Body["input"])
			texts := []string{}
			for _, v := range inputs {
				texts = append(texts, str(v))
			}
			if len(texts) == 0 {
				if s := c.s("input"); s != "" {
					texts = append(texts, s)
				}
			}
			if len(texts) == 0 {
				fail(422, "input 不能为空")
			}
			vecs := a.embedTexts(c.R.Context(), texts)
			if vecs == nil {
				fail(422, "Embedding 服务未配置")
			}
			return M{"model": a.embeddingConf().Model, "data": vecs, "dim": len(vecs[0])}
		}},

		// ---- OCR ----
		{"POST", "/api/settings/ocr/test", false, true, func(c *request) any {
			c.admin()
			conf := a.ocrConf()
			if conf.BaseURL == "" || conf.Model == "" {
				fail(422, "OCR 服务未配置")
			}
			// A tiny generated PNG with the text "HELLO OCR 2026".
			img := "iVBORw0KGgoAAAANSUhEUgAAAPAAAAA8CAYAAABYfzddAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAASSSURBVHhe7ZptkiIhDIY9ngfyON7Fq3gTV12nCXSAEOgaM/U8VfyxOx+EvEDv7OkBAGFBwACBQcAAgUHAAIFBwACBQcAAgUHAAIFBwACBQcAAgUHAAIFBwACBQcAAgUHAAIEZEPD9cT2fHqfTc1xun9863K+P8+v959ibCH9D4/y43j8u3jjy2pix1bldPv7Kcb4+ox3DETFrPs958RXG17Xv08jtovq3Le3tcdnZXp6/GpiKq9faWhMEvEjAVRGVY6GQ18e0r0m9XN51NYpFxRCzUYP79azbfMbUXFu1F/rQR78m3yHgKfHM+FmRQ7mIetFzsZWb0ChHxCxOIK0eRcPpp8RYTbMcnZub9FHmlIlTy0eentlzWWO9dlNxs1rm69e1FSDgyRzkIvZdSJH4T5z1MWWz9vKS72qNPV7T1LCOja3ZY/+p++/l2ng+FVeuoV5va00Q8EwOYve2mwtBfUnMcQG1/HlqmvyNfg9vuTdP74p/gwhTvSun5FFxhW2rjAj4N2y3phg9cY6I6fNZP0E8/pLNqIC3PJqxdP9p42rdOnQhLYnr/GSQIGCvrWkXreE8cY6IOeVTw1NTZz3M6P43ETaFlOYznlsn7oKCI2CnrW33rmNrnpwjYs763DNe05TD6I3ERvIv+9Ca54oe0eNuohafRT/DulkgYKetR4CSbXEH7I+I6cmjzUhNxbum9z2kUzCfoz1P34lZi5t+f7nbfGvDsCY+ATvGfu4+f00/QwV+4bddJqaBk++ImLM+9zj7ZFl8SZ5LvsTpWe+0GxdwK664Vp8/61H4Teu0f1aCgJ22CLjG6Loec20u89iLND1fK+BeXHEyN3xKEbfCcoV22v4VAW+/HSFgraaiJ0bmPoZFJPa1twvYEle+05q/eK8RFwE7bT0ClHjEeEjMyt85/VhqKpt48QmcbRCtHKxrb3zPHFfMvbP2lh5BwF7bbW6eBkyLaP3XxjdHxGyuUYsff6XwPQ2/aPPIfA5cjZtCSvOp+huKK+qDgF/M+Pkl2z/wHznSbcDvL/kYi61S/CnG4s52o0kbn+rTEdd6+9reazhFwDM5iMWzm4sr1LfErIq7hvC3a8KRmop3n8NTjjfFCWj2s9k15r3VRhG5N66p3qnGrRMdAU/msO2S6hxLRONPXBvXx5RC6uUl312wrmJD6p1IOnJ+AyJ608u19XwmrrCtzLl+w8lBwNM5CB/voQtAis5+0tU4ImbekGo9pNieQz8Zxmsq8+x9t5bIa/io7ZvqBiJrvK/dcXFz370afoeAB0desFE/cjFmbHOyoreG65TRWR+zEHFj1FvA0SdZ3JHNzZ7vNpScenXcm6yJW26Iu2GoHwJeJOAfqs2wULgl62PWG7R/2ngE/MRzlS6+QU2jmpMy51oeS+NqPWjfxAYEDADfBgIGCAwCBggMAgYIDAIGCAwCBggMAgYIDAIGCAwCBggMAgYIDAIGCAwCBggMAgYIDAIGCAwCBggMAgYIDAIGCAwCBggMAgYIDAIGCAwCBgjL4/EPQm3XLZf77jUAAAAASUVORK5CYII="
			text := a.ocrImage(c.R.Context(), img, "")
			return M{"ok": true, "model": conf.Model, "base_url": conf.BaseURL, "text": text}
		}},
		{"POST", "/api/ocr", false, false, func(c *request) any {
			image := c.s("image")
			if image == "" {
				image = c.s("url")
			}
			if image == "" {
				// Also accept a file uploaded by the ingest page.
				image = c.s("data_url")
			}
			if image == "" {
				fail(422, "image / url / data_url 至少提供一个")
			}
			text := a.ocrImage(c.R.Context(), image, c.s("prompt"))
			return M{"text": text, "model": a.ocrConf().Model}
		}},
		{"POST", "/api/knowledge/ocr-ingest", false, false, func(c *request) any {
			image := c.s("image")
			if image == "" {
				image = c.s("url")
			}
			if image == "" {
				image = c.s("data_url")
			}
			if image == "" {
				fail(422, "image / url / data_url 至少提供一个")
			}
			text := a.ocrImage(c.R.Context(), image, c.s("prompt"))
			title := c.s("title")
			if title == "" {
				title = "OCR 识别文档"
			}
			doc := M{
				"id":           id(),
				"workspace_id": c.WS,
				"title":        title,
				"content":      text,
				"source_type":  "ocr",
				"created_at":   now(),
			}
			a.insert("knowledge_docs", doc)
			a.indexKnowledgeDoc(c.WS, str(doc["id"]))
			return M{"document": a.owned("knowledge_docs", str(doc["id"]), c.WS), "text": text}
		}},

		// ---- Combined settings view ----
		{"GET", "/api/settings/models", false, true, func(c *request) any {
			return M{
				"chat":      a.chatConf().masked(),
				"embedding": a.embeddingConf().masked(),
				"ocr":       a.ocrConf().masked(),
			}
		}},
		{"PUT", "/api/settings/models", false, true, func(c *request) any {
			c.admin()
			groups := map[string][]string{
				"embedding": {"base_url", "model", "api_key", "dim"},
				"ocr":       {"base_url", "model", "api_key"},
				"llm":       {"max_tokens", "enable_thinking"},
			}
			for prefix, keys := range groups {
				for _, k := range keys {
					full := prefix + "." + k
					if v, ok := c.Body[full]; ok && str(v) != "" {
						a.setSetting(full, v)
					}
				}
			}
			return M{"ok": true}
		}},
	}
}

// masked returns a copy with the API key hidden.
func (c modelConf) masked() M {
	key := ""
	if c.APIKey != "" {
		key = "已配置"
	}
	return M{"base_url": c.BaseURL, "model": c.Model, "api_key_masked": key, "configured": c.BaseURL != "" && c.Model != ""}
}
