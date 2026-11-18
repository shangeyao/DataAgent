package platform

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func nativeTools() any { return metadata["tool_catalog"] }
func periodWindow(period string) M {
	loc, e := time.LoadLocation(env("DATABUDDY_TIMEZONE", "Asia/Shanghai"))
	if e != nil {
		fail(422, "业务时区无效")
	}
	at := time.Now().In(loc)
	today := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, loc)
	start, end := today, today
	key := period
	label := "上月"
	switch key {
	case "today":
		label = "今天"
	case "yesterday":
		start = today.AddDate(0, 0, -1)
		end = start
		label = "昨天"
	case "last_7d", "last_7_days", "last7":
		key = "last_7d"
		start = today.AddDate(0, 0, -6)
		label = "近7天"
	case "last_30d", "last_30_days":
		key = "last_30d"
		start = today.AddDate(0, 0, -29)
		label = "近30天"
	case "last_90d", "last_90_days":
		key = "last_90d"
		start = today.AddDate(0, 0, -89)
		label = "近90天"
	case "this_month":
		start = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, loc)
		label = "本月"
	case "this_year":
		start = time.Date(today.Year(), 1, 1, 0, 0, 0, 0, loc)
		label = "本年"
	default:
		key = "last_month"
		first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, loc)
		start = first.AddDate(0, -1, 0)
		end = first.AddDate(0, 0, -1)
	}
	span := int(end.Sub(start).Hours()/24) + 1
	prevStart, prevEnd := start.AddDate(0, 0, -span), start.AddDate(0, 0, -1)
	if key == "last_month" || key == "this_month" {
		prevStart = start.AddDate(0, -1, 0)
		prevEnd = start.AddDate(0, 0, -1)
	}
	if key == "this_year" {
		prevStart = start.AddDate(-1, 0, 0)
		prevEnd = end.AddDate(-1, 0, 0)
	}
	return M{"key": key, "label": label, "start": start.Format("2006-01-02"), "end": end.Format("2006-01-02"), "prev_start": prevStart.Format("2006-01-02"), "prev_end": prevEnd.Format("2006-01-02")}
}
func (a *App) metric(ws, name string) M {
	for _, m := range a.list("metrics", ws) {
		if strings.EqualFold(str(m["name"]), name) || strings.EqualFold(str(m["display_name"]), name) {
			return decoded(m)
		}
	}
	fail(404, "指标不存在："+name)
	return nil
}
// metricDateCol returns the time dimension column for a metric. It prefers a
// dimension explicitly named the metric's date the semantic layer uses and falls
// back to any dimension name ending in "_date", so the analysis engine is not
// coupled to a specific business schema (retail order_date, semi wip_date, ...).
func metricDateCol(m M) string {
	for _, d := range arr(m["dimensions"]) {
		n := str(d)
		if n == "wip_date" || n == "order_date" || strings.HasSuffix(n, "_date") {
			return n
		}
	}
	return "wip_date"
}
func (a *App) metricSQL(ws string, m M, args M) string {
	t := a.resolve(ws, str(m["table_ref"]))
	joins := arr(m["joins"])
	from := physical(t) + " f"
	dimExpr := map[string]string{}
	for _, c := range a.rows("SELECT name FROM columns WHERE table_id=?", t["id"]) {
		dimExpr[str(c["name"])] = "f." + quote(str(c["name"]))
	}
	for _, v := range joins {
		j := obj(v)
		alias := validName(str(j["alias"]))
		on := str(j["on"])
		guard("SELECT 1 WHERE " + on)
		from += " LEFT JOIN " + physical(a.resolve(ws, str(j["table"]))) + " " + alias + " ON " + on
		for _, v := range arr(j["columns"]) {
			dimExpr[str(v)] = alias + "." + quote(str(v))
		}
	}
	allowed := map[string]bool{}
	for _, d := range arr(m["dimensions"]) {
		allowed[str(d)] = true
	}
	dateCol := metricDateCol(m)
	dims := []string{}
	for _, d := range arr(args["dimensions"]) {
		name := str(d)
		if name == "__date__" {
			name = dateCol
		}
		if !allowed[name] || dimExpr[name] == "" {
			fail(422, "指标不支持维度："+name)
		}
		expr := dimExpr[name]
		if name == dateCol && args["granularity"] != nil {
			grain := str(args["granularity"])
			if grain != "day" && grain != "week" && grain != "month" && grain != "year" {
				fail(422, "无效时间粒度")
			}
			expr = "date_trunc('" + grain + "'," + expr + ")"
		}
		dims = append(dims, expr+" AS "+quote(name))
	}
	expr := str(m["expr"])
	guard("SELECT " + expr)
	selects := append(append([]string{}, dims...), expr+" AS "+quote(str(m["name"])))
	q := "SELECT " + strings.Join(selects, ",") + " FROM " + from
	where := []string{}
	if str(m["filters"]) != "" {
		guard("SELECT 1 WHERE " + str(m["filters"]))
		where = append(where, str(m["filters"]))
	}
	for _, k := range []string{"start", "end"} {
		if v := str(args[k]); v != "" {
			if _, e := time.Parse("2006-01-02", v); e != nil {
				fail(422, "日期格式应为 YYYY-MM-DD")
			}
			if dimExpr[dateCol] != "" {
				op := ">="
				if k == "end" {
					op = "<="
				}
				where = append(where, dimExpr[dateCol]+op+"DATE '"+v+"'")
			}
		}
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	if len(dims) > 0 {
		positions := []string{}
		for i := range dims {
			positions = append(positions, fmt.Sprint(i+1))
		}
		q += " GROUP BY " + strings.Join(positions, ",") + " ORDER BY 1"
	}
	return q
}
func (a *App) queryMetric(ws string, args M) M {
	m := a.metric(ws, str(args["metric"]))
	out := a.query(ws, a.metricSQL(ws, m, args), 500)
	out["metric"] = m
	out["chart"] = chart(out)
	return out
}
func (a *App) ask(ws, question, period string) M {
	if question == "" {
		fail(422, "请输入问题")
	}
	picked := ""
	for _, m := range a.list("metrics", ws) {
		if strings.Contains(strings.ToLower(question), strings.ToLower(str(m["name"]))) || strings.Contains(question, str(m["display_name"])) {
			picked = str(m["name"])
			break
		}
	}
	if picked == "" {
		for _, pair := range [][2]string{{"营收", "revenue"}, {"收入", "revenue"}, {"金额", "revenue"}, {"晶圆", "wafer_out"}, {"出货", "wafer_out"}, {"良率", "yield_rate"}, {"良品", "good_die"}, {"客户", "customer_cnt"}, {"单片", "avg_revenue_per_wafer"}} {
			if strings.Contains(question, pair[0]) {
				picked = pair[1]
				break
			}
		}
	}
	if picked == "" {
		fail(422, "未匹配注册指标，请在问题中指定指标名")
	}
	dims := []string{}
	for _, p := range [][2]string{{"制程", "process_node"}, {"节点", "process_node"}, {"品类", "product_category"}, {"客户行业", "segment"}, {"行业", "segment"}, {"区域", "region"}, {"晶圆厂", "fab_name"}, {"Fab", "fab_name"}, {"趋势", "__date__"}, {"每日", "__date__"}} {
		if strings.Contains(question, p[0]) {
			dims = append(dims, p[1])
			break
		}
	}
	if period == "" {
		for _, p := range [][2]string{{"近7", "last_7d"}, {"近 7", "last_7d"}, {"近30", "last_30d"}, {"近 30", "last_30d"}, {"本月", "this_month"}, {"今天", "today"}, {"昨天", "yesterday"}} {
			if strings.Contains(question, p[0]) {
				period = p[1]
				break
			}
		}
	}
	win := periodWindow(period)
	args := M{"metric": picked, "dimensions": dims, "start": win["start"], "end": win["end"]}
	out := a.queryMetric(ws, args)
	out["engine"] = "metric"
	out["question"] = question
	out["period"] = win
	fallbackAnswer := fmt.Sprintf("按注册口径 %s 查询，共返回 %v 行；时间范围为 %s 至 %s。", picked, out["row_count"], win["start"], win["end"])
	// Use the configured model to turn the deterministic query result into a
	// natural-language answer; fall back to the template on any failure.
	user := fmt.Sprintf("用户问题：%s\n注册指标：%s（%s）\n查询时间范围：%s 至 %s\n返回行数：%v\n",
		question, picked, str(def(a.row("SELECT display_name FROM metrics WHERE workspace_id=? AND name=?", ws, picked), "display_name", picked)),
		win["start"], win["end"], out["row_count"])
	if cols, ok := out["columns"].([]string); ok && len(cols) > 0 {
		user += "结果列：" + strings.Join(cols, "、") + "\n"
	}
	if rows, ok := out["rows"].([][]any); ok {
		for i, row := range rows {
			if i >= 10 {
				user += "…（更多行已省略）\n"
				break
			}
			parts := []string{}
			for _, v := range row {
				parts = append(parts, fmt.Sprint(v))
			}
			user += "- " + strings.Join(parts, " | ") + "\n"
		}
	}
	answer := a.llmTextBackground(
		"你是 OpenDataBuddy 的数据分析师。基于给定的注册指标查询结果，用简洁中文直接回答用户问题。"+
			"只能使用提供的数据，不要编造数字或口径；若数据不足以判断，明确说明。不要输出 Markdown 标题。",
		user, fallbackAnswer)
	if answer == fallbackAnswer {
		out["engine"] = "metric+template"
	} else {
		out["engine"] = "metric+llm"
	}
	out["answer"] = answer
	out["reasoning"] = "匹配注册指标与业务维度，通过确定性语义层编译 SQL 并执行只读校验；自然语言由模型生成。"
	out["chart_data"] = M{"spec": out["chart"], "columns": out["columns"], "rows": out["rows"]}
	return out
}
func (a *App) attribute(ws string, m M) M {
	metric := a.metric(ws, str(m["metric"]))
	dim := str(m["dimension"])
	if dim == "" {
		for _, v := range arr(metric["dimensions"]) {
			if v != metricDateCol(metric) {
				dim = str(v)
				break
			}
		}
	}
	win := periodWindow(str(def(m, "period", "last_month")))
	cur := a.queryMetric(ws, M{"metric": metric["name"], "dimensions": []string{dim}, "start": win["start"], "end": win["end"]})
	prev := a.queryMetric(ws, M{"metric": metric["name"], "dimensions": []string{dim}, "start": win["prev_start"], "end": win["prev_end"]})
	current, previous := map[string]float64{}, map[string]float64{}
	keys := map[string]bool{}
	// Overall ratios and distinct counts cannot be summed across dimension groups.
	currentTotal := a.queryMetric(ws, M{"metric": metric["name"], "start": win["start"], "end": win["end"]})
	previousTotal := a.queryMetric(ws, M{"metric": metric["name"], "start": win["prev_start"], "end": win["prev_end"]})
	totalCur := num(currentTotal["rows"].([][]any)[0][0])
	totalPrev := num(previousTotal["rows"].([][]any)[0][0])
	for _, r := range cur["rows"].([][]any) {
		current[str(r[0])] = num(r[1])
		keys[str(r[0])] = true
	}
	for _, r := range prev["rows"].([][]any) {
		previous[str(r[0])] = num(r[1])
		keys[str(r[0])] = true
	}
	delta := totalCur - totalPrev
	cc := []M{}
	for k := range keys {
		diff := current[k] - previous[k]
		share := float64(0)
		if delta != 0 {
			share = diff / delta
		}
		var growth any
		if previous[k] != 0 {
			growth = diff / previous[k]
		}
		cc = append(cc, M{"value": k, "dimension_value": k, "dim_value": k, "label": k, "current": current[k], "previous": previous[k], "delta": diff, "share": share, "growth": growth, "contribution": share, "contribution_rate": share})
	}
	sort.Slice(cc, func(i, j int) bool { return math.Abs(num(cc[i]["delta"])) > math.Abs(num(cc[j]["delta"])) })
	for i, contribution := range cc {
		contribution["rank"] = i + 1
	}
	driver := ""
	if len(cc) > 0 {
		driver = str(cc[0]["value"])
	}
	rr := [][]any{}
	for _, c := range cc {
		rr = append(rr, []any{c["value"], c["current"], c["previous"], c["delta"], c["contribution"]})
	}
	cols := []string{dim, "本期", "上期", "差值", "贡献度"}
	var growth any
	if totalPrev != 0 {
		growth = delta / totalPrev
	}
	fallbackConclusion := fmt.Sprintf("本期 %g，上期 %g，差值 %g；主要贡献维度为 %s。", totalCur, totalPrev, delta, driver)
	contribLines := []string{}
	for i, c := range cc {
		if i >= 5 {
			break
		}
		contribLines = append(contribLines, fmt.Sprintf("%s：本期 %g，上期 %g，差值 %g，贡献度 %.1f%%", str(c["value"]), num(c["current"]), num(c["previous"]), num(c["delta"]), num(c["share"])*100))
	}
	conclusion := a.llmTextBackground(
		"你是数据分析师。基于给定的指标本期/上期对比和维度贡献数据，用一段简洁中文说明变化驱动因素。"+
			"只使用提供的数据，不要编造；指出主要贡献维度及方向。不要输出 Markdown 标题。",
		fmt.Sprintf("指标：%s\n维度：%s\n周期：%s 至 %s\n本期合计：%g\n上期合计：%g\n差值：%g\n增长率：%v\n主要贡献：%s\n",
			str(metric["display_name"]), dim, win["start"], win["end"], totalCur, totalPrev, delta, growth, strings.Join(contribLines, "；")),
		fallbackConclusion)
	return M{"metric": metric, "dimension": dim, "dimension_label": dim, "period": win, "total_current": totalCur, "total_previous": totalPrev, "total_delta": delta, "growth": growth, "driver": driver, "contributions": cc, "all_contributions": cc, "conclusion": conclusion, "table": M{"columns": cols, "rows": rr}, "chart_data": M{"columns": cols, "rows": rr, "spec": M{"type": "bar", "x": 0, "y": 3}}}
}
func (a *App) report(ws string, m M) M {
	names := arr(m["metrics"])
	if len(names) == 0 {
		for _, r := range a.list("metrics", ws) {
			names = append(names, r["name"])
			if len(names) == 3 {
				break
			}
		}
	}
	win := periodWindow(str(def(m, "period", "last_month")))
	table := [][]any{}
	picked := []M{}
	for _, n := range names {
		metric := a.metric(ws, str(n))
		r := a.queryMetric(ws, M{"metric": n, "start": win["start"], "end": win["end"]})
		value := any(nil)
		if rr := r["rows"].([][]any); len(rr) > 0 {
			value = rr[0][0]
		}
		table = append(table, []any{metric["display_name"], value, metric["unit"]})
		picked = append(picked, metric)
	}
	sections := []M{{"title": "经营概览", "body": "依据注册指标口径计算本期实际数据。", "table": M{"columns": []string{"指标", "本期值", "单位"}, "rows": table}}, {"title": "数据口径", "body": "所有结果来自当前工作空间的本地 DuckDB，指标口径以统一语义层为准。"}, {"title": "质量与风险", "body": fmt.Sprintf("当前共有 %v 条质量规则，最近失败 %v 条。", a.qualitySummary(ws)["rule_count"], a.qualitySummary(ws)["failing_count"])}, {"title": "经营建议", "body": "结合渠道、品类和城市维度检查贡献变化；对异常指标先进行数据质量校验。"}}
	if len(sections) > 0 {
		sections[0]["chart"] = M{"spec": M{"type": "table"}, "columns": []string{"指标", "本期值", "单位"}, "rows": table}
	}
	// Let the model rewrite the narrative sections and overall summary from the
	// real query results. Each section keeps its original text as a fallback.
	metricLines := []string{}
	for _, row := range table {
		metricLines = append(metricLines, fmt.Sprintf("%v：%v %v", row[0], row[1], row[2]))
	}
	reportContext := fmt.Sprintf("报告主题：%s\n周期：%s 至 %s\n指标结果：\n- %s\n",
		def(m, "topic", "经营分析"), win["start"], win["end"], strings.Join(metricLines, "\n- "))
	for _, s := range sections {
		s["heading"] = s["title"]
		orig := str(s["body"])
		s["body"] = a.llmTextBackground(
			"你是经营分析报告撰写助手。根据提供的真实指标数据，为报告章节「"+str(s["title"])+"」写一段简短专业的中文分析。"+
				"只使用提供的数据和口径，不要编造数字；不要输出 Markdown 标题或章节名。",
			reportContext+"\n章节原有要点："+orig,
			orig)
	}
	fallbackSummary := fmt.Sprintf("已完成 %d 个指标的本期统计。", len(picked))
	summary := a.llmTextBackground(
		"你是经营分析报告撰写助手。基于提供的真实指标结果，用一段简短中文总结本期经营情况与需要关注的方向。"+
			"只使用提供的数据，不要编造数字；不要输出 Markdown 标题。",
		reportContext,
		fallbackSummary)
	return M{"title": str(def(m, "topic", "经营分析")) + " · " + str(win["label"]), "topic": def(m, "topic", "经营分析"), "period": win, "metrics": picked, "sections": sections, "summary": summary, "generated_at": now()}
}
// reportChannel turns a schedule channel descriptor into a channel the
// delivery path understands. A schedule stores either a configured channel id
// or an inline {type, target} pair.
func (a *App) reportChannel(ws string, v any) M {
	if s, ok := v.(string); ok {
		return a.row("SELECT * FROM alert_channels WHERE id=? AND workspace_id=?", s, ws)
	}
	d := obj(v)
	kind := str(d["type"])
	target := str(d["target"])
	cfg := M{}
	switch kind {
	case "email":
		cfg["to"] = target
	case "webhook":
		cfg["url"] = target
	case "wecom":
		if strings.HasPrefix(target, "http") {
			cfg["webhook_url"] = target
		} else {
			cfg["key"] = target
		}
	default:
		cfg["webhook_url"] = target
	}
	return M{"id": d["id"], "name": firstNonempty(str(d["name"]), kind), "type": kind, "config": cfg, "target": target}
}

func (a *App) runReport(ws string, r M, trigger string) M {
	out := a.report(ws, M{"topic": def(r, "topic", r["name"]), "metrics": arr(r["metrics"]), "period": r["period"]})
	channels := arr(r["_channels"])
	if len(channels) == 0 {
		// A manual run has no channel list of its own, so the schedules of the
		// report provide the push targets.
		for _, s := range a.rows("SELECT channels FROM report_schedules WHERE workspace_id=? AND report_id=? AND enabled=1", ws, str(r["id"])) {
			channels = append(channels, arr(decoded(s)["channels"])...)
		}
	}
	deliveries := []M{}
	for _, v := range channels {
		channel := a.reportChannel(ws, v)
		if channel == nil {
			deliveries = append(deliveries, M{"channel": str(v), "ok": false, "simulated": false, "detail": "渠道不存在", "target": nil})
			continue
		}
		result := a.deliverAlert(channel, M{"event": "report", "subject": str(out["title"]), "severity": "info", "message": str(out["summary"])})
		target := str(channel["target"])
		if target == "" {
			cfg := obj(channel["config"])
			target = firstNonempty(str(cfg["url"]), firstNonempty(str(cfg["webhook_url"]), firstNonempty(str(cfg["to"]), str(cfg["key"]))))
		}
		deliveries = append(deliveries, M{
			"channel":   firstNonempty(str(channel["name"]), str(channel["type"])),
			"ok":        boolOf(result["ok"]),
			"simulated": boolOf(result["simulated"]),
			"detail":    str(result["message"]),
			"target":    target,
		})
	}
	run := M{"id": id(), "workspace_id": ws, "report_id": r["id"], "status": "success", "trigger": trigger, "title": out["title"], "summary": out["summary"], "sections": out["sections"], "snapshot": out, "channels": deliveries, "started_at": now(), "finished_at": now()}
	a.insert("report_runs", run)
	a.update("reports", str(r["id"]), M{"sections": out["sections"], "summary": out["summary"], "status": "ready", "updated_at": now()})
	return run
}
func (a *App) embed(ws, space string, rotate bool) M {
	a.owned("analysis_spaces", space, ws)
	if !rotate {
		if r := a.row("SELECT * FROM analysis_embeds WHERE workspace_id=? AND space_id=? AND revoked=0 ORDER BY created_at DESC LIMIT 1", ws, space); r != nil {
			r["iframe"] = `<iframe src="` + str(r["readonly_url"]) + `" width="100%" height="720"></iframe>`
			return r
		}
	}
	token := id() + id()
	r := a.insert("analysis_embeds", M{"id": id(), "workspace_id": ws, "space_id": space, "token": token, "readonly_url": "/api/embed/" + token, "scope": "readonly", "revoked": false, "created_at": now()})
	r["iframe"] = `<iframe src="/api/embed/` + token + `"></iframe>`
	return r
}
func (a *App) analysisRoutes() []route {
	r := []route{
		{"POST", "/api/metrics/query", false, false, func(c *request) any { return a.queryMetric(c.WS, c.Body) }},
		{"GET", "/api/analysis/meta", false, false, func(c *request) any {
			mm := a.list("metrics", c.WS)
			dims := []M{}
			seen := map[string]bool{}
			for _, m := range mm {
				decoded(m)
				for _, v := range arr(m["dimensions"]) {
					if !seen[str(v)] {
						dims = append(dims, M{"name": v, "label": v})
						seen[str(v)] = true
					}
				}
			}
			return M{"metrics": mm, "dimensions": dims, "periods": metadata["periods"], "summary": M{"reports": len(a.list("reports", c.WS)), "schedules": len(a.list("report_schedules", c.WS))}}
		}},
		{"POST", "/api/analysis/ask", false, false, func(c *request) any { return a.ask(c.WS, c.s("question"), c.s("period")) }},
		{"POST", "/api/analysis/attribute", false, false, func(c *request) any { return a.attribute(c.WS, c.Body) }},
		{"POST", "/api/analysis/report/generate", false, false, func(c *request) any { return a.report(c.WS, c.Body) }},
		{"POST", "/api/analysis/dashboard/generate", false, false, func(c *request) any {
			question := c.s("prompt")
			if question == "" {
				fail(422, "请输入要生成的看板描述")
			}
			out := a.ask(c.WS, question, c.s("period"))
			out["title"] = question
			description := "由语义层生成"
			reasoning := "按指标口径生成查询与可视化卡片"
			if a.llm("api_key", "") != "" {
				description = a.llmText(c.R.Context(),
					"你是数据分析师。根据用户想看的看板描述，用一句话说明这个看板回答什么问题。不要输出 Markdown 标题。",
					"看板描述："+question, description)
				reasoning = a.llmText(c.R.Context(),
					"你是数据分析师。用一句话说明此看板卡片为何这样设计（指标口径与图表选择）。不要输出 Markdown 标题。",
					fmt.Sprintf("看板描述：%s\n使用指标：%v\n返回行数：%v", question, out["metric"], out["row_count"]), reasoning)
			}
			return M{"name": question, "description": description, "cards": []M{out}, "metrics": []M{}, "dimensions": []M{}, "window": periodWindow(c.s("period"))["label"], "reasoning": reasoning}
		}},
		{"POST", "/api/reports/{report_id}/run", false, false, func(c *request) any {
			return M{"run": a.runReport(c.WS, a.owned("reports", c.p("report_id"), c.WS), "manual")}
		}},
		{"GET", "/api/reports/{report_id}/runs", false, false, func(c *request) any {
			a.owned("reports", c.p("report_id"), c.WS)
			rr := a.rows("SELECT * FROM report_runs WHERE workspace_id=? AND report_id=? ORDER BY started_at DESC", c.WS, c.p("report_id"))
			for _, r := range rr {
				decoded(r)
			}
			return M{"runs": rr}
		}},
		{"GET", "/api/analysis-spaces/{space_id}/embed", false, false, func(c *request) any { return M{"embed": a.embed(c.WS, c.p("space_id"), false)} }},
		{"POST", "/api/analysis-spaces/{space_id}/embed", false, false, func(c *request) any { return M{"embed": a.embed(c.WS, c.p("space_id"), true)} }},
		{"DELETE", "/api/analysis-spaces/{space_id}/embed", false, false, func(c *request) any {
			a.owned("analysis_spaces", c.p("space_id"), c.WS)
			a.exec("UPDATE analysis_embeds SET revoked=1 WHERE workspace_id=? AND space_id=?", c.WS, c.p("space_id"))
			return M{"ok": true}
		}},
		{"GET", "/api/embed/{token}", true, true, func(c *request) any {
			r := a.row("SELECT * FROM analysis_embeds WHERE token=? AND revoked=0 AND (expires_at IS NULL OR expires_at='' OR expires_at>?)", c.p("token"), now())
			if r == nil {
				fail(404, "嵌出链接无效或已撤销")
			}
			space := decoded(a.owned("analysis_spaces", str(r["space_id"]), str(r["workspace_id"])))
			delete(r, "token")
			r["space"] = space
			return M{"embed": r}
		}},
	}
	return r
}
