// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package types

// AssetClass identifies a category of financial instrument within a specific
// regulatory regime. Consumers (BD, ATS, broker) import these constants
// instead of using string literals.
type AssetClass string

const (
	// US asset classes
	AssetClass_US_Equity    AssetClass = "us_equity"
	AssetClass_US_FixedIncome AssetClass = "us_fixed_income"
	AssetClass_US_Options   AssetClass = "us_options"
	AssetClass_US_Crypto    AssetClass = "us_crypto"
	AssetClass_US_RegD      AssetClass = "us_reg_d"       // Reg D private placement
	AssetClass_US_RegA      AssetClass = "us_reg_a"       // Reg A+ offering
	AssetClass_US_RegCF     AssetClass = "us_reg_cf"      // Regulation Crowdfunding

	// Canada
	AssetClass_CA_Equity    AssetClass = "ca_equity"
	AssetClass_CA_Crypto    AssetClass = "ca_crypto"

	// Brazil
	AssetClass_BR_Equity    AssetClass = "br_equity"
	AssetClass_BR_FII       AssetClass = "br_fiis"        // Fundos de Investimento Imobiliario
	AssetClass_BR_Crypto    AssetClass = "br_crypto"

	// India
	AssetClass_IN_Equity    AssetClass = "in_equity"
	AssetClass_IN_FnO       AssetClass = "in_f_and_o"     // Futures & Options (NSE/BSE)
	AssetClass_IN_MutualFund AssetClass = "in_mutual_fund"

	// Singapore
	AssetClass_SG_Equity    AssetClass = "sg_equity"
	AssetClass_SG_REITs     AssetClass = "sg_reits"
	AssetClass_SG_DPT       AssetClass = "sg_dpt"         // Digital Payment Tokens (PS Act)

	// Australia
	AssetClass_AU_Equity    AssetClass = "au_equity"
	AssetClass_AU_ETF       AssetClass = "au_etf"
	AssetClass_AU_Crypto    AssetClass = "au_crypto"

	// Switzerland
	AssetClass_CH_Equity       AssetClass = "ch_equity"
	AssetClass_CH_DLTSecurity  AssetClass = "ch_dlt_security"  // Ledger-based securities (CO art. 973d)

	// UAE
	AssetClass_AE_Equity             AssetClass = "ae_equity"
	AssetClass_AE_TokenizedSecurity  AssetClass = "ae_tokenized_security"
	AssetClass_AE_VirtualAsset       AssetClass = "ae_virtual_asset"       // VARA-regulated

	// EU (MiCA jurisdictions)
	AssetClass_EU_Equity        AssetClass = "eu_equity"        // MiFID II instruments
	AssetClass_EU_AIF           AssetClass = "eu_aif"           // Alternative Investment Funds (AIFMD)
	AssetClass_EU_UCITS         AssetClass = "eu_ucits"         // UCITS funds
	AssetClass_EU_Art           AssetClass = "eu_art"           // Asset-Referenced Token (MiCA)
	AssetClass_EU_EMT           AssetClass = "eu_emt"           // E-Money Token (MiCA)
	AssetClass_EU_CryptoAsset   AssetClass = "eu_crypto_asset"  // Other crypto-assets (MiCA Title II)
	AssetClass_EU_Bond          AssetClass = "eu_bond"
	AssetClass_EU_Derivative    AssetClass = "eu_derivative"
	AssetClass_EU_Securitisation AssetClass = "eu_securitisation"

	// UK
	AssetClass_GB_Equity    AssetClass = "gb_equity"
	AssetClass_GB_Crypto    AssetClass = "gb_crypto"

	// Isle of Man
	AssetClass_IM_Equity    AssetClass = "im_equity"
)

// AllAssetClasses returns all defined asset classes.
func AllAssetClasses() []AssetClass {
	return []AssetClass{
		AssetClass_US_Equity, AssetClass_US_FixedIncome, AssetClass_US_Options,
		AssetClass_US_Crypto, AssetClass_US_RegD, AssetClass_US_RegA, AssetClass_US_RegCF,
		AssetClass_CA_Equity, AssetClass_CA_Crypto,
		AssetClass_BR_Equity, AssetClass_BR_FII, AssetClass_BR_Crypto,
		AssetClass_IN_Equity, AssetClass_IN_FnO, AssetClass_IN_MutualFund,
		AssetClass_SG_Equity, AssetClass_SG_REITs, AssetClass_SG_DPT,
		AssetClass_AU_Equity, AssetClass_AU_ETF, AssetClass_AU_Crypto,
		AssetClass_CH_Equity, AssetClass_CH_DLTSecurity,
		AssetClass_AE_Equity, AssetClass_AE_TokenizedSecurity, AssetClass_AE_VirtualAsset,
		AssetClass_EU_Equity, AssetClass_EU_AIF, AssetClass_EU_UCITS, AssetClass_EU_Art,
		AssetClass_EU_EMT, AssetClass_EU_CryptoAsset, AssetClass_EU_Bond,
		AssetClass_EU_Derivative, AssetClass_EU_Securitisation,
		AssetClass_GB_Equity, AssetClass_GB_Crypto,
		AssetClass_IM_Equity,
	}
}
