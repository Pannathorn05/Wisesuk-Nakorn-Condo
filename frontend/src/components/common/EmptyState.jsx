// สถานะไม่มีข้อมูล — แยกจาก error เสมอ (docs/c.md ข้อ 24)
export function EmptyState({ message = "ยังไม่มีข้อมูลในส่วนนี้", children }) {
  return (
    <div className="state-box">
      <p>{message}</p>
      {children}
    </div>
  );
}
