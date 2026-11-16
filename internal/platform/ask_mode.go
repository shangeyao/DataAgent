package platform

import (
	"regexp"
	"strings"
)

// This file implements the official "ASK 模式" contract (doc 1835/140860):
//   - Ask only produces read-only artefacts (query results, schema, knowledge,
//     analysis summaries); it must not create workflows, ingest tasks, metrics,
//     tables or any other persisted asset.
//   - When the user asks for a mutating action, Buddy must refuse and prompt the
//     user to switch to Agent mode.
//
// The runtime already clears the tool list in agentLoop for chat_mode=ask;
// this file adds the missing explicit refusal, prompt-level guidance and an
// artefact whitelist so the semantics hold on every code path (LLM + heuristic).

const askModePrompt = `当前处于 Ask（只读）模式。你只能读取、查询、分析和总结，不得创建、修改或删除任何资产（表、数据、指标、工作流、接入任务、看板、知识库、记忆、权限），不得执行任何写操作或调用写类工具。
若用户请求属于写操作或修改类需求，不要执行，也不要声称已完成；请明确拒绝并提示："这是修改类请求，请切换到 Agent 模式后再试。"
只读类请求（查询数据、查看表结构、口径解释、知识库检索、生成只读报告/Markdown 总结）正常回答。`

// askReadOnlyArtifactTypes is the whitelist of artefact kinds allowed in Ask
// mode. Everything else (flow_run, dashboard entities, etc.) is filtered out.
var askReadOnlyArtifactTypes = map[string]bool{
	"query":               true,
	"table_schema":        true,
	"table_list":          true,
	"sensitive_scan":      true,
	"quality_results":     true,
	"quality_suggestions": true,
	"profile":             true,
	"knowledge":           true,
	"notice":              true,
}

// askModeRefusal detects a mutating request when the runtime has no LLM tools
// (heuristic path or unconfigured model) and returns an explicit refusal.
// It uses conservative Chinese + English verb patterns; read-only requests
// that merely mention these words in a data value are unlikely to match.
func detectWriteIntent(question string) bool {
	q := strings.TrimSpace(question)
	lower := strings.ToLower(q)

	// Purely informational questions ("how do I create a table?") are not
	// mutation requests — let them through as read-only answers.
	informational := regexp.MustCompile(`^(如何|怎么|怎样|什么是|为什么|能否|可不可以|能不能|是否|what|how|why|can|could)\b|吗？?$|呢？?$`)
	if informational.MatchString(q) || informational.MatchString(lower) {
		// Still refuse if it is an explicit "help me ..." command.
		if !regexp.MustCompile(`帮我|请帮我|帮忙|给我|帮我们`).MatchString(q) {
			return false
		}
	}

	// An explicit imperative addressed to the assistant is the strongest signal.
	if regexp.MustCompile(`(帮我|请帮我|帮忙|给我|帮我们)`).MatchString(q) {
		write := regexp.MustCompile(`创建|新建|新增|添加|加一个|加个|建表|删除|移除|去掉|修改|更新|改动|变更|编辑|重命名|写入|插入|导入|同步|落地|发布|部署|启动|停止|暂停|赋权|授权|打标|打标签|保存|提交|跑一下|执行`)
		if write.MatchString(q) {
			return true
		}
	}

	patterns := []string{
		`^(创建|新建|新增|添加|删除|移除|修改|更新|写入|插入|导入|同步|发布|部署|启动|停止|打标|保存)`,
		`create table`, `insert into`, `drop table`, `alter table`, `delete from`,
	}
	for _, p := range patterns {
		if regexp.MustCompile(p).MatchString(lower) {
			return true
		}
	}
	return false
}

// askRefusalMessage is the exact guidance the official product shows.
const askRefusalMessage = "这是修改类请求，Ask 模式只支持读取与分析，不会创建或修改任何资产。请切换到 Agent 模式后再试。"

// filterAskArtifacts drops any artefact whose type is not in the read-only
// whitelist. Used at agentFinish time so Ask mode can never persist or present
// a write artefact even if a tool slipped through.
func filterAskArtifacts(artifacts []any) []any {
	if len(artifacts) == 0 {
		return artifacts
	}
	out := []any{}
	for _, raw := range artifacts {
		a := obj(raw)
		if askReadOnlyArtifactTypes[str(a["type"])] {
			out = append(out, raw)
		}
	}
	return out
}

// isAskMode reports whether a conversation is in Ask (read-only) mode.
func isAskMode(conv M) bool { return str(conv["chat_mode"]) == "ask" }

// toAnySlice converts a []M to []any for shared helpers.
func toAnySlice(items []M) []any {
	out := make([]any, 0, len(items))
	for _, m := range items {
		out = append(out, m)
	}
	return out
}
