import { ImageWithFallback } from "../common/ImageWithFallback";
import { getRoomImage } from "../../assets/branchPhotos";
import { getRoomChips } from "../../utils/roomChips";
import { formatPrice } from "../../utils/formatPrice";
import "./RoomResultCard.css";

const STAY_TYPE_LABEL = { daily: "ห้องพักรายวัน", monthly: "ห้องพักรายเดือน" };
const STAY_TYPE_UNIT = { daily: "วัน", monthly: "เดือน" };

export function RoomResultCard({ room, branchesById, onBook }) {
  const chips = getRoomChips(room, branchesById);

  return (
    <article className="room-card">
      <div className="room-card__image">
        <ImageWithFallback src={getRoomImage(room, branchesById)} alt={room.branch_name} />
      </div>

      <div className="room-card__body">
        <h3>{STAY_TYPE_LABEL[room.stay_type] || "ห้องพัก"}</h3>
        <p className="room-card__branch">{room.branch_name}</p>

        {chips.length > 0 && (
          <div className="room-card__chips">
            {chips.map((chip, i) => (
              <span className="room-card__chip" key={`${chip}-${i}`}>
                {chip}
              </span>
            ))}
          </div>
        )}

        <hr className="room-card__divider" />

        <div className="room-card__footer">
          <div>
            <p className="room-card__price">
              ฿{formatPrice(room.price)}
              <span>/{STAY_TYPE_UNIT[room.stay_type] || ""}</span>
            </p>
            <p className="room-card__utility">
              ค่าน้ำ {formatPrice(room.water_rate)} บาท/ยูนิต, ค่าไฟ {formatPrice(room.electric_rate)} บาท/ยูนิต
            </p>
          </div>
          <button type="button" className="btn btn-primary" onClick={() => onBook(room)}>
            จองเลย
          </button>
        </div>
      </div>
    </article>
  );
}
