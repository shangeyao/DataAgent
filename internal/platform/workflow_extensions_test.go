package platform

import (
	"strings"
	"testing"
	"time"
)

// §6.11 补充比对闭环测试：系统内置参数 / IF-ELSE / ForEach / Ask 模式 /
// Buddy 定时任务字段 / 波动率质量规则。

func TestBuiltinParameterScope(t *testing.T) {
	start := time.Date(2026, 11, 16, 9, 30, 15, 0, time.UTC)
	f := M{"id": "flow-1", "name": "内置参数测试", "timezone": "Asia/Shanghai"}
	run := M{"id": "run-1", "started_at": start.Format(time.RFC3339), "repair_count": 2}
	scope := renderScope(f, run, M{"custom": "x"}, M{}, start)

	wf := obj(scope["workflow"])
	if str(wf["run_id"]) != "run-1" || num(wf["repair_count"]) != 2 {
		t.Fatal("workflow.run_id / repair_count", wf)
	}
	st := obj(wf["start_time"])
	if str(st["iso_date"]) != "2026-11-16" || num(st["iso_weekday"]) != 1 {
		t.Fatal("workflow.start_time", st)
	}
	if int64(num(st["timestamp_ms"])) != start.UnixMilli() {
		t.Fatal("workflow.start_time.timestamp_ms", st)
	}
	// Legacy alias: custom params remain reachable under workflow.*
	if str(wf["custom"]) != "x" {
		t.Fatal("legacy workflow.<param> alias broken", wf)
	}
	trig := obj(obj(wf["trigger"])["time"])
	if str(trig["iso_datetime"]) != "2026-11-16T09:30:15" {
		t.Fatal("workflow.trigger.time", trig)
	}
}

func TestBuiltinPlaceholderRenderingWithUpstream(t *testing.T) {
	outputs := M{"q": M{"id": "q", "name": "查询", "status": "success", "output": M{"rows": []any{M{"amount": 128}}, "first_row": M{"amount": 128}}}}
	scope := renderScope(M{"id": "f"}, M{"id": "r"}, M{}, outputs, time.Now().UTC())
	got := renderValue("{{tasks.q.output.first_row.amount}}", scope)
	if num(got) != 128 {
		t.Fatal("tasks.<id>.output.first_row.<col>", got)
	}
}

func TestIfElseAndForEachNodes(t *testing.T) {
	a := testApp(t)
	ws := str(a.row("SELECT id FROM workspaces LIMIT 1")["id"])

	// IF-ELSE: expression chooses a branch.
	ifFlow := a.newFlow(ws, M{"name": "ifelse", "nodes": []M{
		{"id": "chk", "type": "ifelse", "config": M{"expression": "1 > 0"}},
	}, "edges": []any{}})
	run := a.runFlow(ws, ifFlow, M{}, "test", 0)
	if run["status"] != "success" {
		t.Fatal("ifelse run failed", run)
	}
	chk := obj(arr(run["nodes"])[0])
	if str(obj(chk["output"])["branch"]) != "true" {
		t.Fatal("ifelse branch", chk)
	}

	// ForEach: iterate a literal list, expose {{loop.item}} to a child SQL node.
	feFlow := a.newFlow(ws, M{"name": "foreach", "nodes": []M{
		{"id": "loop", "type": "foreach", "config": M{
			"items": []any{"a", "b", "c"},
			"nodes": []any{M{"id": "child", "type": "sql", "config": M{"sql": "SELECT 1 AS n"}}},
			"edges": []any{},
		}},
	}, "edges": []any{}})
	feRun := a.runFlow(ws, feFlow, M{}, "test", 0)
	if feRun["status"] != "success" {
		t.Fatal("foreach run failed", feRun)
	}
	loop := obj(arr(feRun["nodes"])[0])
	if num(obj(loop["output"])["count"]) != 3 {
		t.Fatal("foreach count", loop)
	}
	iterations := arr(obj(loop["output"])["iterations"])
	if len(iterations) != 3 || num(obj(iterations[2])["index"]) != 2 {
		t.Fatal("foreach iterations", iterations)
	}
}

func TestBranchExpressionWithStringComparison(t *testing.T) {
	// IF-ELSE branch values are strings ("true"/"false"); edge expressions
	// must compare them as strings rather than coercing to numbers.
	vars := M{"__p0": "true"}
	if !evalCondition(`__p0 == "true"`, vars) {
		t.Fatal("string equality failed")
	}
	if evalCondition(`__p0 == "false"`, vars) {
		t.Fatal("string inequality failed")
	}
}

func TestAskModeRefusesWriteIntent(t *testing.T) {
	for _, q := range []string{
		"帮我创建一张订单表",
		"请帮我新建一个工作流",
		"删除这张表的数据",
		"帮我把这张表加一个字段",
	} {
		if !detectWriteIntent(q) {
			t.Fatal("write intent not detected:", q)
		}
	}
	for _, q := range []string{
		"有哪些表？",
		"这张表有哪些字段",
		"什么是 GMV 口径？",
		"如何创建表？",
		"查询昨天的订单数",
	} {
		if detectWriteIntent(q) {
			t.Fatal("read-only question misclassified:", q)
		}
	}
	// Artefact whitelist drops write-type artefacts.
	arts := []any{M{"type": "query"}, M{"type": "flow_run"}, M{"type": "dashboard"}}
	filtered := filterAskArtifacts(arts)
	if len(filtered) != 1 || str(obj(filtered[0])["type"]) != "query" {
		t.Fatal("ask artefact filter", filtered)
	}
}

func TestBuddyTaskFrequencyAndFields(t *testing.T) {
	cases := []struct {
		t    M
		want string
	}{
		{M{"frequency": "daily", "time_of_day": "07:30"}, "30 7 * * *"},
		{M{"frequency": "interval", "interval_minutes": 30}, "*/30 * * * *"},
		{M{"frequency": "interval", "interval_minutes": 120}, "0 */2 * * *"},
		{M{"cron": "0 9 * * *", "frequency": "daily"}, "0 9 * * *"},
	}
	for _, c := range cases {
		if got := buddyTaskCron(c.t); got != c.want {
			t.Fatalf("buddyTaskCron(%v) = %s, want %s", c.t, got, c.want)
		}
	}
	if buddyTaskDue(M{"valid_from": "2026-11-01", "valid_until": "2026-11-30"}, time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("valid_until not honoured")
	}
}

func TestVolatilityRule(t *testing.T) {
	a := testApp(t)
	ws := str(a.row("SELECT id FROM workspaces LIMIT 1")["id"])
	ref := "demo.semi.fact_wafer_out"
	col := "revenue"

	r := a.newRule(ws, M{"table_ref": ref, "column_name": col, "rule_type": "volatility", "config": M{"max_deviation_rate": 0.3}})
	if str(r["rule_type"]) != "volatility" {
		t.Fatal("volatility rule not created", r)
	}
	// First run captures a baseline and passes.
	first := a.runRule(ws, a.owned("quality_rules", str(r["id"]), ws))
	if str(first["status"]) != "pass" || !strings.Contains(str(first["detail"]), "基线") {
		t.Fatal("first volatility run", first)
	}
	// A wildly different forced baseline should make the next run fail.
	a.exec("DELETE FROM profile_baselines WHERE workspace_id=? AND table_ref=?", ws, ref)
	a.insert("profile_baselines", M{"id": id(), "workspace_id": ws, "table_ref": ref, "mode": "volatility", "column_name": col, "stats": M{"columns": M{col: M{"avg": 999999999}}}, "captured_at": "2026-12-01T00:00:00Z"})
	second := a.runRule(ws, a.owned("quality_rules", str(r["id"]), ws))
	if str(second["status"]) != "fail" {
		t.Fatal("volatility did not detect deviation", second)
	}
}
