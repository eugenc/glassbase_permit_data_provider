package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type permitsListResponse struct {
	Permits []map[string]interface{} `json:"permits"`
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
}

func scanPermitRows(rows pgx.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	desc := rows.FieldDescriptions()
	var out []map[string]interface{}

	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]interface{})
		for i, fd := range desc {
			row[string(fd.Name)] = vals[i]
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// ListPermits returns paginated permit rows for one county (dynamic columns).
func (d *Deps) ListPermits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		countyID := r.PathValue("id")
		ctx := r.Context()

		ex, err := permitsTableExists(ctx, d.Pool, countyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ex {
			writeJSON(w, http.StatusOK, permitsListResponse{
				Permits: []map[string]interface{}{},
				Total:   0,
				Page:    1,
				PerPage: parseIntDefault(r.URL.Query().Get("per_page"), 50),
			})
			return
		}

		page := parseIntDefault(r.URL.Query().Get("page"), 1)
		perPage := parseIntDefault(r.URL.Query().Get("per_page"), 50)
		if perPage > 100 {
			perPage = 100
		}
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * perPage

		search := strings.TrimSpace(r.URL.Query().Get("search"))
		dateFrom := strings.TrimSpace(r.URL.Query().Get("date_from"))
		dateTo := strings.TrimSpace(r.URL.Query().Get("date_to"))

		tbl := permitsTableIdent(countyID)
		ident := pgx.Identifier{tbl}.Sanitize()

		var conds []string
		var args []interface{}
		argPos := 1

		if search != "" {
			conds = append(conds, fmt.Sprintf(
				`(permit_number ILIKE $%d OR CAST(raw_data AS TEXT) ILIKE $%d)`, argPos, argPos))
			args = append(args, "%"+search+"%")
			argPos++
		}
		if dateFrom != "" {
			if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
				http.Error(w, "invalid date_from", http.StatusBadRequest)
				return
			}
			conds = append(conds, fmt.Sprintf(`scraped_at >= $%d::date`, argPos))
			args = append(args, dateFrom)
			argPos++
		}
		if dateTo != "" {
			if _, err := time.Parse("2006-01-02", dateTo); err != nil {
				http.Error(w, "invalid date_to", http.StatusBadRequest)
				return
			}
			conds = append(conds, fmt.Sprintf(`scraped_at < ($%d::date + INTERVAL '1 day')`, argPos))
			args = append(args, dateTo)
			argPos++
		}

		where := ""
		if len(conds) > 0 {
			where = "WHERE " + strings.Join(conds, " AND ")
		}

		countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, ident, where)
		var total int64
		if err := d.Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		argsWithPaging := append([]interface{}{}, args...)
		argsWithPaging = append(argsWithPaging, perPage, offset)

		limitIdx := argPos
		offsetIdx := argPos + 1
		selectSQL := fmt.Sprintf(
			`SELECT * FROM %s %s ORDER BY scraped_at DESC LIMIT $%d OFFSET $%d`,
			ident, where, limitIdx, offsetIdx)

		rows, err := d.Pool.Query(ctx, selectSQL, argsWithPaging...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		list, err := scanPermitRows(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		permitRows := normalizePermitRows(list)

		writeJSON(w, http.StatusOK, permitsListResponse{
			Permits: permitRows,
			Total:   total,
			Page:    page,
			PerPage: perPage,
		})
	}
}

func normalizePermitRows(rows []map[string]interface{}) []map[string]interface{} {
	for _, row := range rows {
		for k, v := range row {
			switch t := v.(type) {
			case time.Time:
				row[k] = t.UTC().Format(time.RFC3339)
			case []byte:
				if len(t) > 0 && (t[0] == '{' || t[0] == '[') {
					var raw json.RawMessage = t
					var parsed interface{}
					if json.Unmarshal(raw, &parsed) == nil {
						row[k] = parsed
						continue
					}
				}
				row[k] = string(t)
			default:
				row[k] = v
			}
		}
	}
	return rows
}

// ExportPermits streams CSV for all permits in a county table.
func (d *Deps) ExportPermits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		countyID := r.PathValue("id")
		ctx := r.Context()

		ex, err := permitsTableExists(ctx, d.Pool, countyID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ex {
			http.Error(w, "no permits table", http.StatusNotFound)
			return
		}

		tbl := permitsTableIdent(countyID)
		ident := pgx.Identifier{tbl}.Sanitize()
		q := fmt.Sprintf(`SELECT * FROM %s ORDER BY scraped_at DESC`, ident)

		rows, err := d.Pool.Query(ctx, q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		desc := rows.FieldDescriptions()
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s_permits.csv"`, countyID))

		writer := csv.NewWriter(w)
		hdr := make([]string, len(desc))
		for i, fd := range desc {
			hdr[i] = string(fd.Name)
		}
		if err := writer.Write(hdr); err != nil {
			return
		}

		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				return
			}
			rec := make([]string, len(vals))
			for i, v := range vals {
				rec[i] = fmt.Sprintf("%v", v)
			}
			if err := writer.Write(rec); err != nil {
				return
			}
		}
		if err := rows.Err(); err != nil {
			return
		}
		writer.Flush()
	}
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
