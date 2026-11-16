package platform

import (
	"strings"
	"time"
)

// This file implements the built-in parameter scope for Workflow, matching the
// official DataBuddy "系统内置参数" contract (doc 1835/138048):
//
//   {{workflow.id}} / {{workflow.name}} / {{workflow.run_id}} / {{workflow.repair_count}}
//   {{workflow.start_time.*}}   (year, month, day, hour, minute, second,
//                                is_weekday, iso_date, iso_datetime, iso_weekday, timestamp_ms)
//   {{workflow.trigger.time.*}} (same shape as start_time; the scheduled time)
//   {{tasks.<node_id>.output.*}} (upstream node output, incl. first_row.<col>)
//   {{params.<key>}}             (user supplied run parameters)
//
// The legacy scope keys "workflow" (= user params) and "nodes" are preserved for
// backward compatibility; see builtinRenderScope / renderScope.

// timeParts expands a time into the official built-in parameter shape. Times are
// normalised to UTC+0 as documented by the official page.
func timeParts(t time.Time) M {
	t = t.UTC()
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return M{
		"year":         t.Year(),
		"month":        t.Format("01"),
		"day":          t.Format("02"),
		"hour":         t.Format("15"),
		"minute":       t.Format("04"),
		"second":       t.Format("05"),
		"is_weekday":   t.Weekday() != time.Saturday && t.Weekday() != time.Sunday,
		"iso_date":     t.Format("2006-01-02"),
		"iso_datetime": t.Format("2006-01-02T15:04:05"),
		"iso_weekday":  wd,
		"timestamp_ms": t.UnixMilli(),
	}
}

// triggerTime resolves the scheduled trigger time of a run. For manual runs the
// start time is used. run_at (single-shot) and cron-derived next fire times are
// honoured through the f["trigger_time"] hint when present.
func triggerTime(f M, run M, start time.Time) time.Time {
	if raw := str(def(f, "trigger_time", "")); raw != "" {
		if t, e := parseRunAt(raw, str(f["timezone"])); e == nil {
			return t
		}
	}
	if raw := str(def(run, "trigger_time", "")); raw != "" {
		if t, e := parseRunAt(raw, str(f["timezone"])); e == nil {
			return t
		}
	}
	if raw := str(def(f, "run_at", "")); raw != "" {
		if t, e := parseRunAt(raw, str(f["timezone"])); e == nil {
			return t
		}
	}
	return start
}

// builtinWorkflowScope builds the official workflow.* built-in parameters.
func builtinWorkflowScope(f M, run M, start time.Time) M {
	repair := 0
	if run != nil {
		repair = int(num(def(run, "repair_count", 0)))
	}
	return M{
		"id":           str(f["id"]),
		"name":         str(f["name"]),
		"run_id":       str(def(run, "id", "")),
		"repair_count": repair,
		"start_time":   timeParts(start),
		"trigger":      M{"time": timeParts(triggerTime(f, run, start))},
	}
}

// builtinTasksScope expands upstream node outputs into tasks.<node_id>.*.
// The official contract exposes output.first_row.<col> for SQL nodes; we expose
// the full output map plus a first_row alias derived from output.rows[0].
func builtinTasksScope(outputs M) M {
	tasks := M{}
	for key, raw := range outputs {
		node := obj(raw)
		out := obj(node["output"])
		if len(out) == 0 {
			out = node
		}
		if fr := obj(out["first_row"]); len(fr) == 0 {
			if rows := arr(out["rows"]); len(rows) > 0 {
				out["first_row"] = obj(rows[0])
			}
		}
		tasks[key] = M{
			"id":     str(node["id"]),
			"name":   str(node["name"]),
			"type":   str(node["type"]),
			"status": str(node["status"]),
			"output": out,
		}
	}
	return tasks
}

// renderScope assembles the full render scope passed to renderValue.
//
// Compatibility note: historically "workflow" held user params and "nodes"
// held upstream outputs. We keep those keys so existing flows keep working,
// while adding the official built-ins under their documented names.
func renderScope(f M, run M, params M, outputs M, start time.Time) M {
	// Merge user params into workflow.* so legacy flows that used
	// {{workflow.<custom_param>}} keep working alongside the new built-ins.
	wf := builtinWorkflowScope(f, run, start)
	for k, v := range params {
		if _, exists := wf[k]; !exists {
			wf[k] = v
		}
	}
	return M{
		"workflow": wf,
		"tasks":    builtinTasksScope(outputs),
		"params":   params,
		"nodes":    outputs,
		// legacy aliases used by older flows
		"flow":    f,
		"run":     run,
		"context": params,
	}
}

// builtinRenderScope is a convenience wrapper that records the run start time
// from the run record itself when available.
func builtinRenderScope(f M, run M, params M, outputs M) M {
	start := time.Now().UTC()
	if run != nil {
		if t, e := time.Parse(time.RFC3339, str(def(run, "started_at", ""))); e == nil {
			start = t
		}
	}
	return renderScope(f, run, params, outputs, start)
}

// firstRowValue extracts {{tasks.<id>.output.first_row.<col>}} from a render
// scope. It is used by IF-ELSE / ForEach node evaluation where we need a typed
// value rather than the string produced by renderValue.
func firstRowValue(scope M, expr string) any {
	expr = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expr, "{{"), "}}"))
	parts := strings.Split(expr, ".")
	var cur any = scope
	for _, p := range parts {
		cur = obj(cur)[p]
	}
	return cur
}

// evalBranchExpression renders built-in placeholders in an IF-ELSE expression
// and evaluates it. Numeric / boolean values are rendered natively so that
// comparisons like "amount > 100" work on real numbers rather than strings.
func evalBranchExpression(expr string, scope M) bool {
	expr = placeholder.ReplaceAllStringFunc(expr, func(s string) string {
		v := firstRowValue(scope, s)
		switch v.(type) {
		case float64, int, bool:
			return str(v)
		}
		return quote(str(v))
	})
	return evalCondition(expr, M{})
}

// replaceWithVars replaces every {{path}} placeholder in expr with the string
// returned by fn (typically a generated variable name) and returns the result.
func replaceWithVars(expr string, fn func(path string) string) string {
	return placeholder.ReplaceAllStringFunc(expr, func(s string) string {
		path := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, "{{"), "}}"))
		return fn(path)
	})
}

// loopValues normalises a ForEach items configuration into a concrete slice.
// Accepts a literal array, a {{...}} placeholder resolving to an array, or a
// comma separated string.
func loopValues(raw any, scope M) []any {
	switch x := raw.(type) {
	case []any:
		return x
	case []M:
		out := []any{}
		for _, m := range x {
			out = append(out, m)
		}
		return out
	case string:
		s := strings.TrimSpace(x)
		if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") {
			if v := firstRowValue(scope, s); v != nil {
				if got := loopValues(v, scope); len(got) > 0 {
					return got
				}
			}
		}
		out := []any{}
		for _, part := range strings.Split(s, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	return []any{}
}
