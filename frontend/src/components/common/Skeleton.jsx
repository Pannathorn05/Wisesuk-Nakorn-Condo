// building block กลางของสถานะ loading — ใช้แทน placeholder รูป/ข้อความระหว่างรอ API
export function Skeleton({ width = "100%", height = "1rem", radius, className = "", style }) {
  return (
    <span
      className={`skeleton ${className}`}
      style={{ display: "block", width, height, borderRadius: radius, ...style }}
      aria-hidden="true"
    />
  );
}
