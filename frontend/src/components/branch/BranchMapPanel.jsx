import { isEmbeddableMapUrl } from "../../utils/mapEmbed";
import { IconPin } from "../icons";
import "./BranchMapPanel.css";

export function BranchMapPanel({ mapUrl, branchName }) {
  if (!mapUrl) {
    return (
      <div className="branch-media-box branch-map__empty">
        <IconPin width={48} height={48} />
        <p>ยังไม่มีแผนที่สำหรับสาขานี้</p>
      </div>
    );
  }

  if (isEmbeddableMapUrl(mapUrl)) {
    return (
      <div className="branch-media-box">
        <iframe src={mapUrl} title={`แผนที่ ${branchName}`} loading="lazy" allowFullScreen />
      </div>
    );
  }

  // ไม่ใช่ embed URL — เปิดเป็นลิงก์ธรรมดาแทน ปลอดภัยกว่าการเดา embed แล้วฝังผิดรูปแบบ
  return (
    <div className="branch-media-box branch-map__link">
      <IconPin width={48} height={48} />
      <p>ดูตำแหน่งสาขาบนแผนที่</p>
      <a href={mapUrl} target="_blank" rel="noopener noreferrer" className="btn btn-primary">
        เปิดใน Google Maps
      </a>
    </div>
  );
}
