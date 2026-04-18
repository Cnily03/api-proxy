package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"api-proxy/internal/db"
	"api-proxy/internal/model"
)

// Business-logic layer for proxy rules, wrapping the DB and applying
// validation and normalization around raw CRUD operations.
type RuleService struct {
	db *db.DB
}

// Constructs a RuleService backed by the given database handle.
func NewRuleService(d *db.DB) *RuleService {
	return &RuleService{db: d}
}

// Returns all proxy rules as stored in the database.
func (s *RuleService) List() ([]model.Rule, error) {
	return s.db.ListRules()
}

// Validates and normalizes the incoming rule, enforces Src uniqueness,
// then inserts it and returns the generated ID.
func (s *RuleService) Create(item model.Rule) (int64, error) {
	normalized, err := normalizeAndValidate(item)
	if err != nil {
		return 0, err
	}
	if err := s.db.ValidateUniqueSrc(normalized.Src, 0); err != nil {
		return 0, err
	}
	return s.db.CreateRule(normalized)
}

// Validates and normalizes the rule, enforces Src uniqueness against
// other rules, then persists the update; returns sql.ErrNoRows when
// the rule does not exist.
func (s *RuleService) Update(item model.Rule) error {
	normalized, err := normalizeAndValidate(item)
	if err != nil {
		return err
	}
	normalized.ID = item.ID
	if err := s.db.ValidateUniqueSrc(normalized.Src, normalized.ID); err != nil {
		return err
	}
	return s.db.UpdateRule(normalized)
}

// Removes the rule with the given ID; returns sql.ErrNoRows when no
// such rule exists.
func (s *RuleService) Delete(id int64) error {
	return s.db.DeleteRule(id)
}

// Finds the rule whose Src is the longest prefix of the given request
// path; returns sql.ErrNoRows for empty or non-absolute paths and when
// no rule matches.
func (s *RuleService) FindMatchByPath(path string) (*model.Rule, error) {
	if path == "" || !strings.HasPrefix(path, "/") {
		return nil, sql.ErrNoRows
	}
	return s.db.FindRuleByPath(path)
}

// Trims whitespace, canonicalizes the Src path, validates that Dest
// is an absolute http/https URL, and returns a cleaned copy of the
// rule ready for persistence; returns an error on any validation
// failure.
func normalizeAndValidate(item model.Rule) (model.Rule, error) {
	src := normalizeSrc(item.Src)
	if src == "" || !strings.HasPrefix(src, "/") {
		return model.Rule{}, errors.New("src must start with '/'")
	}

	dest := strings.TrimSpace(item.Dest)
	parsed, err := url.Parse(dest)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return model.Rule{}, fmt.Errorf("dest must be a valid absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return model.Rule{}, fmt.Errorf("dest must use http or https")
	}
	parsed.Fragment = ""

	tags := item.Tags
	if tags == nil {
		tags = []string{}
	}

	return model.Rule{
		Name:           strings.TrimSpace(item.Name),
		Src:            src,
		Dest:           parsed.String(),
		DestAPIKey:     strings.TrimSpace(item.DestAPIKey),
		SkipCertVerify: item.SkipCertVerify,
		APIKey:         strings.TrimSpace(item.APIKey),
		ForceAPIKey:    item.ForceAPIKey,
		Comment:        strings.TrimSpace(item.Comment),
		Tags:           tags,
	}, nil
}

// Canonicalizes a rule Src path by trimming whitespace, ensuring a
// single leading '/', and stripping any trailing '/' except for the
// root path.
func normalizeSrc(src string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}
	if !strings.HasPrefix(src, "/") {
		src = "/" + src
	}
	if src != "/" {
		src = strings.TrimRight(src, "/")
	}
	return src
}
