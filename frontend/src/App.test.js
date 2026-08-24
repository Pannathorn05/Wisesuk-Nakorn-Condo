import { render, screen, within } from "@testing-library/react";
import App from "./App";

// stub fetch ทุกเทส — เทสหน้า UI ไม่ควรพึ่ง backend จริงที่รันอยู่หรือไม่ก็ได้
beforeEach(() => {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      ok: true,
      status: 200,
      text: () => Promise.resolve(JSON.stringify({ data: [] })),
    })
  );
});

afterEach(() => {
  jest.restoreAllMocks();
});

test("แสดงหน้า Homepage พร้อมหัวข้อหลักเมื่อเปิดเว็บ", async () => {
  render(<App />);
  expect(await screen.findByRole("heading", { level: 1, name: /หอพักวิเศษสุขนคร คอนโด/i })).toBeInTheDocument();
});

test("แสดงเมนูนำทางหลักครบ 4 รายการ", async () => {
  render(<App />);
  const nav = within(screen.getByRole("navigation", { name: "เมนูหลัก" }));
  for (const label of ["หน้าหลัก", "สาขาของเรา", "ค้นหาห้องพัก", "ติดต่อเรา"]) {
    expect(await nav.findByRole("link", { name: label })).toBeInTheDocument();
  }
});
