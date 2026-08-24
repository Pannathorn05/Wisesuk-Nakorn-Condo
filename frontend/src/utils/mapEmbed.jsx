// เดารูปแบบ map_url อย่างปลอดภัย — ไม่มีข้อมูลตัวอย่างจริงตอนนี้ (ว่างเปล่าทุกสาขา) จึงรองรับทั้ง 2 กรณี
// ห้ามฝัง API key แผนที่ใหม่เอง (docs/c.md ข้อ 22) — ใช้ URL ที่ backend ให้มาตรง ๆ เท่านั้น
export function isEmbeddableMapUrl(url) {
  if (!url) return false;
  return /\/maps\/embed|output=embed/i.test(url);
}
