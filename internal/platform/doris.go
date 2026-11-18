package platform

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Doris connectivity is configured via environment variables so real credentials
// can be filled in later without code changes:
//
//	DATABUDDY_DORIS_HOST     e.g. 10.0.0.10   (required)
//	DATABUDDY_DORIS_PORT     e.g. 9030         (default 9030, MySQL feather port)
//	DATABUDDY_DORIS_DB       e.g. doris_demo    (default empty -> information_schema)
//	DATABUDDY_DORIS_USER     default root
//	DATABUDDY_DORIS_PASSWORD default (empty)
//	DATABUDDY_DORIS_ENABLED  "1" enables the Doris query backend
//
// Apache Doris exposes its query interface over the MySQL wire protocol
// (default FE mysql port 9030), so a standard mysql driver is used. No data is
// copied into the local DuckDB lake: reads hit Doris directly. Hadoop/HDFS is
// not involved.
//
// When enabled, user-facing read paths (/api/sql/run, catalog preview,
// analysis) route to Doris. The internal machinery (quality, drift, landing,
// sync, seeding) keeps operating on the local DuckDB lake so it is unaffected.

const (
	dorisDriverName = "mysql"
	// dorisReadTimeout guards against Doris hanging on a heavy query.
	dorisReadTimeout = 30 * time.Second
)

// dorisConfig captures the effective Doris connection settings.
func dorisConfig() (M, bool) {
	host := os.Getenv("DATABUDDY_DORIS_HOST")
	if host == "" || env("DATABUDDY_DORIS_ENABLED", "") != "1" {
		return nil, false
	}
	port := env("DATABUDDY_DORIS_PORT", "9030")
	db := env("DATABUDDY_DORIS_DB", "")
	user := env("DATABUDDY_DORIS_USER", "root")
	pass := env("DATABUDDY_DORIS_PASSWORD", "")
	return M{
		"host": host, "port": port, "db": db, "user": user, "password": pass,
	}, true
}

// dorisDSN builds a go-sql-driver/mysql DSN for a Doris FE. Doris is lenient
// about the DSN; we set tls=false, collation to utf8mb4_general_ci and allow
// native password fallback, which covers the common Doris deployments.
func dorisDSN(cfg M) string {
	params := []string{
		"tls=false",
		"allowNativePasswords=true",
		"allowCleartextPasswords=true",
		"parseTime=true",
		"charset=utf8mb4",
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		str(cfg["user"]), str(cfg["password"]),
		str(cfg["host"]), str(cfg["port"]),
		str(cfg["db"]), strings.Join(params, "&"))
}

// dorisPool is a lazily-created, concurrency-safe connection pool to the Doris FE.
type dorisPool struct {
	mu sync.Mutex
	db *sql.DB
}

var theDoris = &dorisPool{}

// get returns a *sql.DB for Doris, opening one on first use. It fails fast
// (clear message) when Doris is not configured or unreachable.
func (d *dorisPool) get() (*sql.DB, error) {
	cfg, ok := dorisConfig()
	if !ok {
		return nil, fmt.Errorf("Doris 未启用：请设置 DATABUDDY_DORIS_HOST 与 DATABUDDY_DORIS_ENABLED=1")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db != nil {
		return d.db, nil
	}
	db, e := sql.Open(dorisDriverName, dorisDSN(cfg))
	if e != nil {
		return nil, fmt.Errorf("Doris 驱动初始化失败: %v", e)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(3 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = db.PingContext(ctx); e != nil {
		db.Close()
		return nil, fmt.Errorf("Doris 连接失败: %v", e)
	}
	d.db = db
	return d.db, nil
}

func (d *dorisPool) close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
}

// dorisEnabled reports whether the Doris backend is switched on and configured.
func dorisEnabled() bool {
	return os.Getenv("DATABUDDY_DORIS_HOST") != "" && env("DATABUDDY_DORIS_ENABLED", "") == "1"
}

// dorisTest verifies connectivity and returns the Doris/MySQL version plus
// whether the configured account can list databases.
func (a *App) dorisTest() M {
	cfg, ok := dorisConfig()
	if !ok {
		return M{"ok": false, "error": "Doris 未配置：缺少 DATABUDDY_DORIS_HOST 或未设置 DATABUDDY_DORIS_ENABLED=1", "host": os.Getenv("DATABUDDY_DORIS_HOST")}
	}
	db, e := theDoris.get()
	if e != nil {
		return M{"ok": false, "error": e.Error(), "host": cfg["host"], "port": cfg["port"]}
	}
	ctx, cancel := context.WithTimeout(context.Background(), dorisReadTimeout)
	defer cancel()
	var version string
	_ = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	dbs := []string{}
	rows, e := db.QueryContext(ctx, "SHOW DATABASES")
	if e == nil {
		defer rows.Close()
		for rows.Next() {
			var n string
			if rows.Scan(&n) == nil {
				dbs = append(dbs, n)
			}
		}
	}
	return M{"ok": true, "version": version, "databases": dbs, "host": cfg["host"], "port": cfg["port"], "database": cfg["db"]}
}

// dorisRoutes exposes the Doris query backend over the HTTP API. These are
// global (non-workspace) read-only routes: the data lives only in the external
// Doris cluster, never copied to the local lake.
func (a *App) dorisRoutes() []route {
	routes := []route{
		{"GET", "/api/doris/test", false, true, func(c *request) any {
			return M{"doris": a.dorisTest()}
		}},
		{"POST", "/api/doris/query", false, true, func(c *request) any {
			q := guard(c.s("sql"))
			limit := int(num(def(c.Body, "limit", 1000)))
			out := a.queryDoris(q, limit)
			out["chart"] = chart(out)
			return out
		}},
	}
	return routes
}

// queryDoris executes a read-only query against Doris and returns the standard
// {columns, rows, row_count} result contract the rest of the platform consumes.
// The SQL is assumed to be a validated single read statement (callers go
// through guard()/qualify() first).
func (a *App) queryDoris(q string, limit int) M {
	if limit <= 0 || limit > 10000 {
		limit = 1000
	}
	db, e := theDoris.get()
	if e != nil {
		fail(400, e.Error())
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), dorisReadTimeout)
	defer cancel()
	r, e := db.QueryContext(ctx, q)
	if e != nil {
		fail(400, "Doris 查询失败: "+e.Error()+"\nSQL: "+q)
	}
	defer r.Close()
	cols, e := r.Columns()
	must(e)
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = c
	}
	result := [][]any{}
	truncated := false
	for r.Next() {
		if len(result) == limit {
			truncated = true
			break
		}
		values := make([]any, len(cols))
		pointers := make([]any, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}
		must(r.Scan(pointers...))
		for i, v := range values {
			switch v := v.(type) {
			case nil:
				values[i] = nil
			case []byte:
				values[i] = string(v)
			case time.Time:
				values[i] = v.Format("2006-01-02 15:04:05")
			default:
				values[i] = v
			}
		}
		result = append(result, values)
	}
	must(r.Err())
	return M{"columns": colNames, "rows": result, "row_count": len(result), "truncated": truncated, "elapsed_ms": time.Since(start).Milliseconds(), "sql": q, "backend": "doris"}
}