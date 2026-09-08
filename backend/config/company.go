package config

import (
	"os"
	"strings"
)

// CompanyConfig holds optional firm details for invoice PDFs.
type CompanyConfig struct {
	Name               string
	Address            string
	City               string
	PostalCode         string
	Country            string
	Phone              string
	Email              string
	TaxID              string
	RegistrationNumber string
	BankName           string
	BankAccount        string
	Website            string
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// LoadCompanyConfig reads COMPANY_* env values. Empty fields stay empty.
func LoadCompanyConfig() CompanyConfig {
	name := envTrim("COMPANY_NAME")
	if name == "" {
		name = "AM HADŽIĆ KERAMIKA DOO TUTIN"
	}
	return CompanyConfig{
		Name:               name,
		Address:            withDefault("COMPANY_ADDRESS", "Treće sandžačke brigade 1"),
		City:               withDefault("COMPANY_CITY", "Tutin"),
		PostalCode:         withDefault("COMPANY_POSTAL_CODE", "36320"),
		Country:            withDefault("COMPANY_COUNTRY", "Srbija"),
		Phone:              withDefault("COMPANY_PHONE", "063 652 222"),
		Email:              envTrim("COMPANY_EMAIL"),
		TaxID:              withDefault("COMPANY_TAX_ID", "113560128"),
		RegistrationNumber: withDefault("COMPANY_REGISTRATION_NUMBER", "21890162"),
		BankName:           withDefault("COMPANY_BANK_NAME", "Halkbank"),
		BankAccount:        withDefault("COMPANY_BANK_ACCOUNT", "155-0000000082232-82"),
		Website:            envTrim("COMPANY_WEBSITE"),
	}
}

func withDefault(key string, fallback string) string {
	if value := envTrim(key); value != "" {
		return value
	}
	return fallback
}
