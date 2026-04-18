package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"api-proxy/internal/model"
)

// Returns all proxy rules ordered by ID ascending, decoding tag JSON
// and boolean flags into their model representation.
func (d *DB) ListRules() ([]model.Rule, error) {
	rows, err := d.conn.Query(`
		SELECT id, name, src, dest, dest_api_key, skip_cert_verify, api_key, force_api_key, comment, tags
		FROM rules ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Rule, 0)
	for rows.Next() {
		var item model.Rule
		var skip, forceKey int
		var tagsJSON string
		if err := rows.Scan(&item.ID, &item.Name, &item.Src, &item.Dest,
			&item.DestAPIKey, &skip, &item.APIKey, &forceKey, &item.Comment, &tagsJSON); err != nil {
			return nil, err
		}
		item.SkipCertVerify = skip == 1
		item.ForceAPIKey = forceKey == 1
		_ = json.Unmarshal([]byte(tagsJSON), &item.Tags)
		if item.Tags == nil {
			item.Tags = []string{}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Inserts a new rule row and returns the generated auto-increment ID.
func (d *DB) CreateRule(item model.Rule) (int64, error) {
	tagsJSON, _ := json.Marshal(item.Tags)
	res, err := d.conn.Exec(`
		INSERT INTO rules (name, src, dest, dest_api_key, skip_cert_verify, api_key, force_api_key, comment, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Name, item.Src, item.Dest, item.DestAPIKey,
		boolToInt(item.SkipCertVerify), item.APIKey, boolToInt(item.ForceAPIKey), item.Comment, string(tagsJSON))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Overwrites the rule identified by item.ID with the provided values;
// returns sql.ErrNoRows when no matching row exists.
func (d *DB) UpdateRule(item model.Rule) error {
	tagsJSON, _ := json.Marshal(item.Tags)
	res, err := d.conn.Exec(`
		UPDATE rules SET name=?, src=?, dest=?, dest_api_key=?, skip_cert_verify=?,
		api_key=?, force_api_key=?, comment=?, tags=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		item.Name, item.Src, item.Dest, item.DestAPIKey,
		boolToInt(item.SkipCertVerify), item.APIKey, boolToInt(item.ForceAPIKey), item.Comment, string(tagsJSON), item.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Deletes the rule with the given ID; returns sql.ErrNoRows when no
// matching row exists.
func (d *DB) DeleteRule(id int64) error {
	res, err := d.conn.Exec(`DELETE FROM rules WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Finds the rule whose Src path is the longest prefix of the given
// request path; returns sql.ErrNoRows when no rule matches.
func (d *DB) FindRuleByPath(path string) (*model.Rule, error) {
	items, err := d.ListRules()
	if err != nil {
		return nil, err
	}
	var matched *model.Rule
	for i := range items {
		rule := items[i]
		if isPathMatched(path, rule.Src) {
			if matched == nil || len(rule.Src) > len(matched.Src) {
				copied := rule
				matched = &copied
			}
		}
	}
	if matched == nil {
		return nil, sql.ErrNoRows
	}
	return matched, nil
}

// Checks that no other rule (excluding ignoreID) already uses the given
// Src path; returns an error describing the conflict if one is found.
func (d *DB) ValidateUniqueSrc(src string, ignoreID int64) error {
	items, err := d.ListRules()
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Src == src && item.ID != ignoreID {
			return fmt.Errorf("src path already exists: %s", src)
		}
	}
	return nil
}

// Reports whether the request path falls under the rule's src prefix:
// an exact match, a root-src catch-all, or a proper "/"-separated
// prefix segment.
func isPathMatched(path string, src string) bool {
	if path == src {
		return true
	}
	if src == "/" {
		return strings.HasPrefix(path, "/")
	}
	return strings.HasPrefix(path, src+"/")
}
