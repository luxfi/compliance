// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package aml

import (
	"context"
	"testing"
)

func newTestScreeningService() *ScreeningService {
	svc := NewScreeningService(DefaultScreeningConfig())

	// OFAC entries
	svc.AddEntry(SanctionEntry{
		ID:      "ofac-001",
		List:    ListOFAC,
		Name:    "Viktor Bout",
		Aliases: []string{"Victor Bout", "Viktor Butt"},
		Country: "RU",
		Details: "Arms dealer",
	})
	svc.AddEntry(SanctionEntry{
		ID:      "ofac-002",
		List:    ListOFAC,
		Name:    "Osama Bin Laden",
		Country: "SA",
		Details: "Deceased terrorist leader",
	})

	// EU sanctions
	svc.AddEntry(SanctionEntry{
		ID:      "eu-001",
		List:    ListEU,
		Name:    "Alexandr Lukashenko",
		Country: "BY",
		Details: "President of Belarus",
	})

	// UK sanctions
	svc.AddEntry(SanctionEntry{
		ID:      "uk-001",
		List:    ListUK,
		Name:    "Vladimir Putin",
		Country: "RU",
		Details: "President of Russia",
	})

	// PEP
	svc.AddEntry(SanctionEntry{
		ID:       "pep-001",
		List:     ListPEP,
		Name:     "Xi Jinping",
		Country:  "CN",
		Category: "head_of_state",
	})

	// Adverse media
	svc.AddEntry(SanctionEntry{
		ID:      "media-001",
		List:    ListAdverseMedia,
		Name:    "Sam Bankman Fried",
		Country: "US",
		Details: "FTX fraud",
	})

	return svc
}

func TestScreenExactMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Viktor",
		FamilyName: "Bout",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for exact OFAC match")
	}
	if result.Risk != RiskCritical {
		t.Fatalf("expected critical risk for exact OFAC, got %q", result.Risk)
	}
	if len(result.Matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if result.Matches[0].MatchType != MatchExact {
		t.Fatalf("expected exact match, got %q", result.Matches[0].MatchType)
	}
	if result.Matches[0].Score != 1.0 {
		t.Fatalf("expected score 1.0, got %f", result.Matches[0].Score)
	}
}

func TestScreenAliasMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Victor",
		FamilyName: "Bout",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for alias match")
	}
	found := false
	for _, m := range result.Matches {
		if m.MatchType == MatchExact && m.MatchedName == "Victor Bout" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected exact alias match for 'Victor Bout'")
	}
}

func TestScreenFuzzyMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Vikter",
		FamilyName: "Bout",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for fuzzy match")
	}
	found := false
	for _, m := range result.Matches {
		if m.MatchType == MatchFuzzy {
			found = true
		}
	}
	if !found {
		t.Fatal("expected fuzzy match")
	}
}

func TestScreenNoMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "John",
		FamilyName: "Smith",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clear {
		t.Fatal("expected clear for non-matching name")
	}
	if result.Risk != RiskLow {
		t.Fatalf("expected low risk, got %q", result.Risk)
	}
	if len(result.Matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(result.Matches))
	}
}

func TestScreenSpecificLists(t *testing.T) {
	svc := newTestScreeningService()
	// Only check PEP list - Xi Jinping won't appear on OFAC
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Xi",
		FamilyName: "Jinping",
		Lists:      []ListSource{ListPEP},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for PEP match")
	}
	if len(result.Matches) == 0 {
		t.Fatal("expected PEP match")
	}
	if result.Matches[0].List != ListPEP {
		t.Fatalf("expected PEP list, got %q", result.Matches[0].List)
	}
}

func TestScreenPEPNotOnOFAC(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Xi",
		FamilyName: "Jinping",
		Lists:      []ListSource{ListOFAC},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clear {
		t.Fatal("expected clear when checking only OFAC for PEP entry")
	}
}

func TestScreenEUMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Alexandr",
		FamilyName: "Lukashenko",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for EU sanctions match")
	}
}

func TestScreenUKMatch(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Vladimir",
		FamilyName: "Putin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for UK sanctions match")
	}
}

func TestScreenAdverseMedia(t *testing.T) {
	svc := newTestScreeningService()
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Sam",
		FamilyName: "Bankman Fried",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Clear {
		t.Fatal("expected not clear for adverse media match")
	}
}

func TestScreenEmptyName(t *testing.T) {
	svc := newTestScreeningService()
	_, err := svc.Screen(context.Background(), &ScreeningRequest{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestBatchScreen(t *testing.T) {
	svc := newTestScreeningService()
	requests := []*ScreeningRequest{
		{GivenName: "Viktor", FamilyName: "Bout"},
		{GivenName: "John", FamilyName: "Smith"},
		{GivenName: "Xi", FamilyName: "Jinping"},
	}
	results, err := svc.BatchScreen(context.Background(), requests)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Clear {
		t.Fatal("expected Viktor Bout not clear")
	}
	if !results[1].Clear {
		t.Fatal("expected John Smith clear")
	}
	if results[2].Clear {
		t.Fatal("expected Xi Jinping not clear")
	}
}

func TestGetResult(t *testing.T) {
	svc := newTestScreeningService()
	result, _ := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "John",
		FamilyName: "Smith",
	})
	fetched, err := svc.GetResult(result.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ID != result.ID {
		t.Fatalf("expected ID %q, got %q", result.ID, fetched.ID)
	}
}

func TestGetResultNotFound(t *testing.T) {
	svc := newTestScreeningService()
	_, err := svc.GetResult("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent result")
	}
}

func TestRiskAssessment(t *testing.T) {
	svc := newTestScreeningService()

	// Exact OFAC = critical
	result, _ := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName: "Viktor", FamilyName: "Bout",
	})
	if result.Risk != RiskCritical {
		t.Fatalf("expected critical for exact OFAC, got %q", result.Risk)
	}

	// Exact non-OFAC = high
	result, _ = svc.Screen(context.Background(), &ScreeningRequest{
		GivenName: "Vladimir", FamilyName: "Putin",
	})
	if result.Risk != RiskHigh {
		t.Fatalf("expected high for exact UK sanctions, got %q", result.Risk)
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"John Doe", "john doe"},
		{"JOHN  DOE", "john doe"},
		{"John-Doe", "johndoe"},
		{"John  O'Brien ", "john obrien"},
	}
	for _, tc := range cases {
		got := normalize(tc.input)
		if got != tc.expected {
			t.Errorf("normalize(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
	}
	for _, tc := range cases {
		got := levenshtein(tc.a, tc.b)
		if got != tc.expected {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.expected)
		}
	}
}

func TestSimilarity(t *testing.T) {
	if s := similarity("abc", "abc"); s != 1.0 {
		t.Fatalf("expected 1.0 for identical, got %f", s)
	}
	if s := similarity("abc", "abd"); s < 0.5 {
		t.Fatalf("expected > 0.5 for close strings, got %f", s)
	}
	if s := similarity("", ""); s != 1.0 {
		t.Fatalf("expected 1.0 for empty, got %f", s)
	}
}

func TestFuzzyDisabled(t *testing.T) {
	cfg := DefaultScreeningConfig()
	cfg.EnableFuzzy = false
	svc := NewScreeningService(cfg)
	svc.AddEntry(SanctionEntry{
		ID:   "ofac-001",
		List: ListOFAC,
		Name: "Viktor Bout",
	})

	// Fuzzy match should not fire
	result, err := svc.Screen(context.Background(), &ScreeningRequest{
		GivenName:  "Vikter",
		FamilyName: "Bout",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Clear {
		t.Fatal("expected clear when fuzzy disabled and name differs")
	}
}
