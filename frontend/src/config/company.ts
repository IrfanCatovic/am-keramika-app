/**
 * Company details for invoice print documents.
 * Values come only from NEXT_PUBLIC_* env — never invent missing fields.
 */

function env(key: string): string {
  const value = process.env[key];
  return typeof value === 'string' ? value.trim() : '';
}

export const COMPANY_LOGO_SRC = '/logo-stampa-racuni/logo-stampa-racuni.svg';

/** Primary public storefront brand mark. */
export const STOREFRONT_LOGO_SRC = '/logo-stampa-racuni/logo-stampa-racuni.svg';
export const STOREFRONT_HERO_SRC = '/logo-stampa-racuni/Amslika.webp';

export const STOREFRONT_SALON_SRC = '/logo-stampa-racuni/slika1.webp';

export const companyConfig = {
  name: env('NEXT_PUBLIC_COMPANY_NAME') || 'AM HADŽIĆ KERAMIKA DOO TUTIN',
  address:
    env('NEXT_PUBLIC_COMPANY_ADDRESS') || 'Treće sandžačke brigade 1',
  city: env('NEXT_PUBLIC_COMPANY_CITY') || 'Tutin',
  postalCode: env('NEXT_PUBLIC_COMPANY_POSTAL_CODE') || '36320',
  country: env('NEXT_PUBLIC_COMPANY_COUNTRY') || 'Srbija',
  phone: env('NEXT_PUBLIC_COMPANY_PHONE') || '063 652 222',
  email: env('NEXT_PUBLIC_COMPANY_EMAIL'),
  taxId: env('NEXT_PUBLIC_COMPANY_TAX_ID') || '113560128',
  registrationNumber:
    env('NEXT_PUBLIC_COMPANY_REGISTRATION_NUMBER') || '21890162',
  bankName: env('NEXT_PUBLIC_COMPANY_BANK_NAME') || 'Halkbank',
  bankAccount:
    env('NEXT_PUBLIC_COMPANY_BANK_ACCOUNT') || '155-0000000082232-82',
  website: env('NEXT_PUBLIC_COMPANY_WEBSITE'),
} as const;

export type CompanyConfig = typeof companyConfig;

export function companyAddressLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.address) lines.push(config.address);
  const locality = [config.postalCode, config.city].filter(Boolean).join(' ');
  const location = [locality, config.country].filter(Boolean).join(', ');
  if (location) lines.push(location);
  return lines;
}

export function companyContactLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.phone) lines.push(`Telefon: ${config.phone}`);
  if (config.email) lines.push(config.email);
  if (config.website) lines.push(config.website);
  return lines;
}

export function companyIdLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.taxId) lines.push(`PIB: ${config.taxId}`);
  if (config.registrationNumber) {
    lines.push(`Matični broj: ${config.registrationNumber}`);
  }
  return lines;
}

/** Tekući račun za štampu računa (bez PIB/MB). */
export function companyBankLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.bankAccount) {
    const bank = config.bankName
      ? `${config.bankAccount} — ${config.bankName}`
      : config.bankAccount;
    lines.push(`Tekući račun: ${bank}`);
  }
  return lines;
}
