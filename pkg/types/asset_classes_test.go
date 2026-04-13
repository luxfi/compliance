// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package types

import "testing"

func TestAllAssetClasses(t *testing.T) {
	all := AllAssetClasses()
	if len(all) < 30 {
		t.Fatalf("expected >= 30 asset classes, got %d", len(all))
	}

	// Verify uniqueness
	seen := map[AssetClass]bool{}
	for _, ac := range all {
		if seen[ac] {
			t.Errorf("duplicate asset class: %q", ac)
		}
		seen[ac] = true
		if ac == "" {
			t.Error("empty asset class constant")
		}
	}
}

func TestAssetClassConstants(t *testing.T) {
	// Spot-check a few constants have expected values
	cases := []struct {
		ac   AssetClass
		want string
	}{
		{AssetClass_US_Equity, "us_equity"},
		{AssetClass_US_Crypto, "us_crypto"},
		{AssetClass_EU_AIF, "eu_aif"},
		{AssetClass_EU_Art, "eu_art"},
		{AssetClass_CH_DLTSecurity, "ch_dlt_security"},
		{AssetClass_AE_TokenizedSecurity, "ae_tokenized_security"},
		{AssetClass_IN_FnO, "in_f_and_o"},
		{AssetClass_SG_REITs, "sg_reits"},
		{AssetClass_AU_ETF, "au_etf"},
		{AssetClass_BR_FII, "br_fiis"},
		{AssetClass_CA_Equity, "ca_equity"},
	}
	for _, tc := range cases {
		if string(tc.ac) != tc.want {
			t.Errorf("AssetClass constant = %q, want %q", tc.ac, tc.want)
		}
	}
}
