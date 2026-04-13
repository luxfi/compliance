// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package regulatory

// Framework enumerates regulatory frameworks that span one or more jurisdictions.
type Framework string

const (
	Framework_US_SEC_FINRA Framework = "us_sec_finra"
	Framework_UK_FCA       Framework = "uk_fca"
	Framework_IOM          Framework = "iom"
	Framework_CIRO         Framework = "ciro"
	Framework_CVM          Framework = "cvm"
	Framework_SEBI         Framework = "sebi"
	Framework_MAS          Framework = "mas"
	Framework_ASIC         Framework = "asic"
	Framework_FINMA        Framework = "finma"
	Framework_SCA          Framework = "sca"
	Framework_DFSA         Framework = "dfsa"
	Framework_FSRA         Framework = "fsra"
	Framework_VARA         Framework = "vara"
	Framework_MICA         Framework = "mica"
)

// frameworkJurisdictions maps each framework to the jurisdiction codes that operate under it.
var frameworkJurisdictions = map[Framework][]string{
	Framework_US_SEC_FINRA: {"US"},
	Framework_UK_FCA:       {"GB"},
	Framework_IOM:          {"IM"},
	Framework_CIRO:         {"CA"},
	Framework_CVM:          {"BR"},
	Framework_SEBI:         {"IN"},
	Framework_MAS:          {"SG"},
	Framework_ASIC:         {"AU"},
	Framework_FINMA:        {"CH"},
	Framework_SCA:          {"AE"},
	Framework_DFSA:         {"AE-DIFC"},
	Framework_FSRA:         {"AE-ADGM"},
	Framework_VARA:         {"AE-VARA"},
	Framework_MICA:         {"LU", "DE", "FR", "NL", "IE", "IT", "ES"},
}

// JurisdictionsByFramework returns all jurisdiction codes under a framework.
func JurisdictionsByFramework(f Framework) []string {
	codes, ok := frameworkJurisdictions[f]
	if !ok {
		return nil
	}
	out := make([]string, len(codes))
	copy(out, codes)
	return out
}

// AllFrameworks returns all supported frameworks.
func AllFrameworks() []Framework {
	return []Framework{
		Framework_US_SEC_FINRA, Framework_UK_FCA, Framework_IOM,
		Framework_CIRO, Framework_CVM, Framework_SEBI, Framework_MAS,
		Framework_ASIC, Framework_FINMA, Framework_SCA, Framework_DFSA,
		Framework_FSRA, Framework_VARA, Framework_MICA,
	}
}
