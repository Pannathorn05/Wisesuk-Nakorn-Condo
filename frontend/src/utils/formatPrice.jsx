const THB_FORMATTER = new Intl.NumberFormat("th-TH", { maximumFractionDigits: 0 });

export function formatPrice(amount) {
  if (amount === null || amount === undefined || Number.isNaN(amount)) return "-";
  return THB_FORMATTER.format(amount);
}
