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
	Phone              string
	Email              string
	TaxID              string
	RegistrationNumber string
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
		name = "AM Keramika"
	}
	return CompanyConfig{
		Name:               name,
		Address:            envTrim("COMPANY_ADDRESS"),
		City:               envTrim("COMPANY_CITY"),
		Phone:              envTrim("COMPANY_PHONE"),
		Email:              envTrim("COMPANY_EMAIL"),
		TaxID:              envTrim("COMPANY_TAX_ID"),
		RegistrationNumber: envTrim("COMPANY_REGISTRATION_NUMBER"),
		BankAccount:        envTrim("COMPANY_BANK_ACCOUNT"),
		Website:            envTrim("COMPANY_WEBSITE"),
	}
}
