// สถานะ error กลาง — ทุกจุดที่ดึง API พังต้องโชว์อันนี้แทนหน้าขาว (docs/c.md ข้อ 24)
export function ErrorState({ message = "เกิดข้อผิดพลาด กรุณาลองใหม่", onRetry }) {
  return (
    <div className="state-box" role="alert">
      <p>{message}</p>
      {onRetry && (
        <button type="button" className="btn btn-outline" onClick={onRetry}>
          ลองใหม่
        </button>
      )}
    </div>
  );
}
