import { isEmbeddableMapUrl, getBranchMapEmbedUrl, getBranchOpenMapUrl } from "../../utils/mapEmbed";
import { IconPin } from "../icons";
import "./BranchMapPanel.css";

export function BranchMapPanel({ branch }) {
  // ลำดับความสำคัญ (map_url จาก backend มาก่อนเสมอถ้ามีจริง → พิกัดจริงที่รู้อยู่แล้ว → เดาจาก
  // ข้อความที่อยู่) อยู่ในฟังก์ชันเดียวที่ mapEmbed.jsx ทั้งหมดแล้ว
  const branchName = branch?.name;
  const effectiveUrl = getBranchMapEmbedUrl(branch);

  if (!effectiveUrl) {
    return (
      <div className="branch-media-box branch-map__empty">
        <IconPin width={48} height={48} />
        <p>ยังไม่มีแผนที่สำหรับสาขานี้</p>
      </div>
    );
  }

  if (isEmbeddableMapUrl(effectiveUrl)) {
    // ปุ่ม "เปิดใน Maps" ที่ Google แปะมาเองมุมซ้ายบนของ iframe (ควบคุมไม่ได้ เป็น UI ของ Google เอง)
    // พาไปแค่หน้าพิกัดดิบ ไม่มีข้อมูลสถานที่ เพราะ q= ด้านบนใช้ lat,lng ไม่ใช่ธุรกิจจริง — ลิงก์
    // นี้แยกไว้ต่างหาก (getBranchOpenMapUrl) พาไปหน้าสถานที่จริงของ Google Business Profile แทน
    const openUrl = getBranchOpenMapUrl(branch);
    return (
      <div className="branch-media-box">
        <iframe src={effectiveUrl} title={`แผนที่ ${branchName}`} loading="lazy" allowFullScreen />
        {openUrl && (
          <a href={openUrl} target="_blank" rel="noopener noreferrer" className="branch-map__open-link">
            <IconPin width={14} height={14} /> เปิดหน้าสถานที่เต็มใน Maps
          </a>
        )}
      </div>
    );
  }

  // ไม่ใช่ embed URL — เปิดเป็นลิงก์ธรรมดาแทน ปลอดภัยกว่าการเดา embed แล้วฝังผิดรูปแบบ
  // (ลิงก์ที่สร้างจากที่อยู่จริงด้านบนก็เข้าเงื่อนไขนี้เสมอ เพราะไม่ใช่ /maps/embed หรือ output=embed)
  return (
    <div className="branch-media-box branch-map__link">
      <IconPin width={48} height={48} />
      <p>ดูตำแหน่งสาขาบนแผนที่</p>
      <a href={effectiveUrl} target="_blank" rel="noopener noreferrer" className="btn btn-primary">
        เปิดใน Google Maps
      </a>
    </div>
  );
}
