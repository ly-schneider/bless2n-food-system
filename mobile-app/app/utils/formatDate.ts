import { de } from "date-fns/locale";
import { format } from "date-fns/format"
import { parseISO } from "date-fns/parseISO"
import i18n from "i18next"

const localeMap: { [key: string]: typeof de } = { de };

export function formatDate(
  iso: string,
  pattern = "dd.MM.yyyy",
  options?: Parameters<typeof format>[2],
) {
  const locale = localeMap[i18n.language.split("-")[0]] ?? de;
  return format(parseISO(iso), pattern, { ...options, locale });
}
