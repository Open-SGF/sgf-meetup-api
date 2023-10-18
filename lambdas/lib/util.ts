export const isDev = process.env.NODE_ENV === 'development';

/**
 * Parse a "YYYYMMDD" string into a date
 * 
 * Example: `parseDateString("20231015") => new Date(2023, 9, 15)`
 */
export function parseDateString(dateString: string): Date {
  const year = parseInt(dateString.slice(0, 4));
  const month = parseInt(dateString.slice(4, 6)) - 1;
  const day = parseInt(dateString.slice(6, 8));

  const date = new Date(year, month, day);

  return date;
}
