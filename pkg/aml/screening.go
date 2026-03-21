// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package aml provides AML/sanctions screening and transaction monitoring.
// It checks applicants against OFAC SDN, EU/UK sanctions, PEP databases,
// and performs real-time transaction monitoring with configurable rules.
package aml

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"
)

// MatchType describes how a screening match was found.
type MatchType string

const (
	MatchExact   MatchType = "exact"
	MatchFuzzy   MatchType = "fuzzy"
	MatchPartial MatchType = "partial"
)

// RiskLevel describes the assessed risk from a screening.
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// ListSource identifies the sanctions/watchlist source.
type ListSource string

const (
	ListOFAC        ListSource = "ofac_sdn"
	ListEU          ListSource = "eu_sanctions"
	ListUK          ListSource = "uk_hmt"
	ListPEP         ListSource = "pep"
	ListAdverseMedia ListSource = "adverse_media"
)

// ScreeningRequest contains the subject to screen.
type ScreeningRequest struct {
	ID          string   `json:"id,omitempty"` // caller-provided reference
	GivenName   string   `json:"given_name"`
	FamilyName  string   `json:"family_name"`
	DateOfBirth string   `json:"date_of_birth,omitempty"`
	Country     string   `json:"country,omitempty"`
	Nationality string   `json:"nationality,omitempty"`
	TaxID       string   `json:"tax_id,omitempty"`
	Email       string   `json:"email,omitempty"`
	Lists       []ListSource `json:"lists,omitempty"` // which lists to check; nil = all
}

// ScreeningResult is the outcome of a screening check.
type ScreeningResult struct {
	ID          string           `json:"id"`
	RequestID   string           `json:"request_id,omitempty"`
	Risk        RiskLevel        `json:"risk"`
	Matches     []ScreeningMatch `json:"matches"`
	ScreenedAt  time.Time        `json:"screened_at"`
	TotalChecks int              `json:"total_checks"`
	Clear       bool             `json:"clear"` // true if no actionable matches
}

// ScreeningMatch is a single hit against a watchlist.
type ScreeningMatch struct {
	List        ListSource `json:"list"`
	MatchType   MatchType  `json:"match_type"`
	Score       float64    `json:"score"` // 0.0-1.0 confidence
	MatchedName string     `json:"matched_name"`
	ListID      string     `json:"list_id,omitempty"` // entry ID on the sanctions list
	Details     string     `json:"details,omitempty"`
	Country     string     `json:"country,omitempty"`
}

// SanctionEntry represents a single entry on a sanctions/watchlist.
type SanctionEntry struct {
	ID       string     `json:"id"`
	List     ListSource `json:"list"`
	Name     string     `json:"name"`
	Aliases  []string   `json:"aliases,omitempty"`
	Country  string     `json:"country,omitempty"`
	DOB      string     `json:"dob,omitempty"`
	Details  string     `json:"details,omitempty"`
	Category string     `json:"category,omitempty"` // for PEP: head_of_state, minister, etc.
}

// ScreeningService performs AML/sanctions screening.
type ScreeningService struct {
	mu      sync.RWMutex
	entries map[ListSource][]SanctionEntry
	results map[string]*ScreeningResult // resultID -> result
	config  ScreeningConfig
}

// ScreeningConfig controls screening behavior.
type ScreeningConfig struct {
	FuzzyThreshold float64 // minimum score for fuzzy matches (0.0-1.0)
	EnableFuzzy    bool    // enable fuzzy name matching
}

// DefaultScreeningConfig returns sensible defaults.
func DefaultScreeningConfig() ScreeningConfig {
	return ScreeningConfig{
		FuzzyThreshold: 0.80,
		EnableFuzzy:    true,
	}
}

// NewScreeningService creates an AML screening service.
func NewScreeningService(cfg ScreeningConfig) *ScreeningService {
	return &ScreeningService{
		entries: make(map[ListSource][]SanctionEntry),
		results: make(map[string]*ScreeningResult),
		config:  cfg,
	}
}

// AddEntry adds a sanctions/watchlist entry for screening against.
func (s *ScreeningService) AddEntry(entry SanctionEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[entry.List] = append(s.entries[entry.List], entry)
}

// Screen checks a subject against configured watchlists.
func (s *ScreeningService) Screen(ctx context.Context, req *ScreeningRequest) (*ScreeningResult, error) {
	if req.GivenName == "" && req.FamilyName == "" {
		return nil, fmt.Errorf("screening requires at least given_name or family_name")
	}

	resultID := newScreeningID()
	result := &ScreeningResult{
		ID:         resultID,
		RequestID:  req.ID,
		ScreenedAt: time.Now(),
		Matches:    []ScreeningMatch{},
		Clear:      true,
	}

	fullName := normalize(req.GivenName + " " + req.FamilyName)
	listsToCheck := req.Lists
	if len(listsToCheck) == 0 {
		listsToCheck = []ListSource{ListOFAC, ListEU, ListUK, ListPEP, ListAdverseMedia}
	}

	// Read entries under read lock
	s.mu.RLock()
	totalChecks := 0
	for _, listSrc := range listsToCheck {
		entries := s.entries[listSrc]
		for _, entry := range entries {
			totalChecks++
			matches := s.checkEntry(fullName, req, entry)
			result.Matches = append(result.Matches, matches...)
		}
	}
	s.mu.RUnlock()

	result.TotalChecks = totalChecks

	if len(result.Matches) > 0 {
		result.Clear = false
		result.Risk = s.assessRisk(result.Matches)
	} else {
		result.Risk = RiskLow
	}

	// Store result under write lock
	s.mu.Lock()
	s.results[resultID] = result
	s.mu.Unlock()

	return result, nil
}

// BatchScreen screens multiple subjects.
func (s *ScreeningService) BatchScreen(ctx context.Context, requests []*ScreeningRequest) ([]*ScreeningResult, error) {
	results := make([]*ScreeningResult, 0, len(requests))
	for _, req := range requests {
		result, err := s.Screen(ctx, req)
		if err != nil {
			return results, fmt.Errorf("batch screening failed for %s %s: %w", req.GivenName, req.FamilyName, err)
		}
		results = append(results, result)
	}
	return results, nil
}

// GetResult returns a previously computed screening result.
func (s *ScreeningService) GetResult(id string) (*ScreeningResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.results[id]
	if !ok {
		return nil, fmt.Errorf("screening result %s not found", id)
	}
	return r, nil
}

func (s *ScreeningService) checkEntry(fullName string, req *ScreeningRequest, entry SanctionEntry) []ScreeningMatch {
	var matches []ScreeningMatch

	entryName := normalize(entry.Name)

	// Exact match
	if fullName == entryName {
		matches = append(matches, ScreeningMatch{
			List:        entry.List,
			MatchType:   MatchExact,
			Score:       1.0,
			MatchedName: entry.Name,
			ListID:      entry.ID,
			Details:     entry.Details,
			Country:     entry.Country,
		})
		return matches
	}

	// Check aliases for exact match
	for _, alias := range entry.Aliases {
		if fullName == normalize(alias) {
			matches = append(matches, ScreeningMatch{
				List:        entry.List,
				MatchType:   MatchExact,
				Score:       1.0,
				MatchedName: alias,
				ListID:      entry.ID,
				Details:     entry.Details,
				Country:     entry.Country,
			})
			return matches
		}
	}

	// Partial match: check if all parts of the query appear in the entry name
	queryParts := strings.Fields(fullName)
	entryParts := strings.Fields(entryName)
	if len(queryParts) > 0 && len(entryParts) > 0 {
		allFound := true
		for _, qp := range queryParts {
			found := false
			for _, ep := range entryParts {
				if qp == ep {
					found = true
					break
				}
			}
			if !found {
				allFound = false
				break
			}
		}
		if allFound && fullName != entryName {
			matches = append(matches, ScreeningMatch{
				List:        entry.List,
				MatchType:   MatchPartial,
				Score:       0.85,
				MatchedName: entry.Name,
				ListID:      entry.ID,
				Details:     entry.Details,
				Country:     entry.Country,
			})
			return matches
		}
	}

	// Fuzzy match using Levenshtein distance
	if s.config.EnableFuzzy {
		score := similarity(fullName, entryName)
		if score >= s.config.FuzzyThreshold {
			matches = append(matches, ScreeningMatch{
				List:        entry.List,
				MatchType:   MatchFuzzy,
				Score:       score,
				MatchedName: entry.Name,
				ListID:      entry.ID,
				Details:     entry.Details,
				Country:     entry.Country,
			})
		}

		// Check aliases for fuzzy
		for _, alias := range entry.Aliases {
			score := similarity(fullName, normalize(alias))
			if score >= s.config.FuzzyThreshold {
				matches = append(matches, ScreeningMatch{
					List:        entry.List,
					MatchType:   MatchFuzzy,
					Score:       score,
					MatchedName: alias,
					ListID:      entry.ID,
					Details:     entry.Details,
					Country:     entry.Country,
				})
			}
		}
	}

	return matches
}

func (s *ScreeningService) assessRisk(matches []ScreeningMatch) RiskLevel {
	maxScore := 0.0
	hasExact := false
	hasOFAC := false

	for _, m := range matches {
		if m.Score > maxScore {
			maxScore = m.Score
		}
		if m.MatchType == MatchExact {
			hasExact = true
		}
		if m.List == ListOFAC {
			hasOFAC = true
		}
	}

	// Exact OFAC hit is always critical
	if hasExact && hasOFAC {
		return RiskCritical
	}
	if hasExact {
		return RiskHigh
	}
	if maxScore >= 0.90 {
		return RiskHigh
	}
	if maxScore >= 0.80 {
		return RiskMedium
	}
	return RiskLow
}

// normalize lowercases, replaces hyphens with spaces, strips non-alpha chars,
// and collapses whitespace for comparison. This ensures "Al-Qaeda" and "Al Qaeda"
// normalize identically.
func normalize(s string) string {
	// Replace hyphens with spaces before stripping non-letter chars so that
	// hyphenated names like "Al-Qaeda" match "Al Qaeda".
	s = strings.ReplaceAll(s, "-", " ")
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	// Collapse whitespace
	return strings.Join(strings.Fields(b.String()), " ")
}

// similarity computes a normalized similarity score using Levenshtein distance.
func similarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	d := levenshtein(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(d)/float64(maxLen)
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Use two rows instead of full matrix for O(min(la,lb)) space
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(ins, del, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func newScreeningID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return "scr_" + hex.EncodeToString(b)
}
