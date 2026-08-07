/**
 * Company details for invoice print documents.
 * Values come only from NEXT_PUBLIC_* env — never invent missing fields.
 */

function env(key: string): string {
  const value = process.env[key];
  return typeof value === 'string' ? value.trim() : '';
}

export const COMPANY_LOGO_SRC = '/logo-stampa-racuni/logo-stampa-racuni.svg';

/** Primary public storefront brand mark (marble AM identity). */
export const STOREFRONT_LOGO_SRC =
  '/logo-stampa-racuni/logo-stampa-racuni.jpeg';

export const STOREFRONT_HERO_SRC = '/logo-stampa-racuni/Amslika.jpg';

export const STOREFRONT_SALON_SRC = '/logo-stampa-racuni/slika1.jpg';

export const companyConfig = {
  name: env('NEXT_PUBLIC_COMPANY_NAME') || 'AM Keramika',
  address: env('NEXT_PUBLIC_COMPANY_ADDRESS'),
  city: env('NEXT_PUBLIC_COMPANY_CITY'),
  phone: env('NEXT_PUBLIC_COMPANY_PHONE'),
  email: env('NEXT_PUBLIC_COMPANY_EMAIL'),
  taxId: env('NEXT_PUBLIC_COMPANY_TAX_ID'),
  registrationNumber: env('NEXT_PUBLIC_COMPANY_REGISTRATION_NUMBER'),
  bankAccount: env('NEXT_PUBLIC_COMPANY_BANK_ACCOUNT'),
  website: env('NEXT_PUBLIC_COMPANY_WEBSITE'),
} as const;

export type CompanyConfig = typeof companyConfig;

export function companyAddressLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.address) lines.push(config.address);
  if (config.city) lines.push(config.city);
  return lines;
}

export function companyContactLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.phone) lines.push(`Tel: ${config.phone}`);
  if (config.email) lines.push(config.email);
  if (config.website) lines.push(config.website);
  return lines;
}

export function companyIdLines(
  config: CompanyConfig = companyConfig
): string[] {
  const lines: string[] = [];
  if (config.taxId) lines.push(`PIB: ${config.taxId}`);
  if (config.registrationNumber) lines.push(`MB: ${config.registrationNumber}`);
  if (config.bankAccount) lines.push(`Žiro račun: ${config.bankAccount}`);
  return lines;
}
