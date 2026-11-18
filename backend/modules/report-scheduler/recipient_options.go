package reportscheduler

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"net/mail"
	"strings"
)

// Return contact labels only; never expose roles, group membership or user-admin fields.
func (r *postgresRepository) ListRecipientOptions(ctx context.Context, kind, search string, limit int, selected []string) ([]RecipientOption, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("invalid recipient option limit")
	}
	if kind == "group" {
		rows, err := r.db.QueryContext(ctx, `SELECT id::text,COALESCE(NULLIF(display_name,''),name),path
   FROM ihp_rbac_groups WHERE enabled=TRUE
   AND ($1='' OR path ILIKE '%'||$1||'%' OR name ILIKE '%'||$1||'%' OR display_name ILIKE '%'||$1||'%' OR id::text=ANY($3::text[]) OR path=ANY($3::text[]))
   ORDER BY (id::text=ANY($3::text[]) OR path=ANY($3::text[])) DESC,path,id LIMIT $2`, search, limit, pq.Array(selected))
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		options := make([]RecipientOption, 0)
		for rows.Next() {
			var id, label, path string
			if err := rows.Scan(&id, &label, &path); err != nil {
				return nil, err
			}
			if path != "" {
				label += " · " + path
			}
			value := id
			for _, saved := range selected {
				if saved == path {
					value = path
					break
				}
			}
			options = append(options, RecipientOption{ID: "group:" + value, Type: "group", Value: value, Label: label})
		}
		return options, rows.Err()
	}
	if kind != "email" && kind != "portal" {
		return nil, fmt.Errorf("invalid recipient option kind")
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id::text,username,COALESCE(first_name,''),COALESCE(last_name,''),COALESCE(email,'')
  FROM users WHERE enabled=TRUE AND username NOT LIKE 'service-account-%'
  AND ($1='' OR username ILIKE '%'||$1||'%' OR concat_ws(' ',first_name,last_name) ILIKE '%'||$1||'%' OR email ILIKE '%'||$1||'%' OR id::text=ANY($4::text[]) OR email=ANY($4::text[]))
  AND ($3='portal' OR COALESCE(email,'')<>'')
  ORDER BY COALESCE((id::text=ANY($4::text[]) OR email=ANY($4::text[])),FALSE) DESC,lower(COALESCE(first_name,'')),lower(username),id LIMIT $2`, search, limit, kind, pq.Array(selected))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	options := make([]RecipientOption, 0)
	for rows.Next() {
		var id, username, first, last, email string
		if err := rows.Scan(&id, &username, &first, &last, &email); err != nil {
			return nil, err
		}
		if kind == "email" {
			if _, err := mail.ParseAddress(email); err != nil {
				continue
			}
		}
		label := strings.TrimSpace(first + " " + last)
		if label == "" {
			label = username
		} else if label != username {
			label += " · " + username
		}
		if email != "" {
			label += " (" + email + ")"
		}
		option := RecipientOption{ID: "user:" + id, Type: "user", Value: id, Label: label}
		if kind == "email" {
			for _, saved := range selected {
				if saved == email {
					option.ID = "email:" + email
					option.Type = "email"
					option.Value = email
					break
				}
			}
		}
		options = append(options, option)
	}
	return options, rows.Err()
}
