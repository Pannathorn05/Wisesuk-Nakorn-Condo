import { getAmenityIcon } from "../icons/amenityIconMap";
import "./BranchDetailCards.css";

export function BranchAmenitiesCard({ amenities }) {
  const list = amenities || [];

  return (
    <div className="branch-detail-card branch-detail-card--amenities">
      <h3>สิ่งอำนวยความสะดวก</h3>
      {list.length === 0 ? (
        <p className="branch-detail-card__empty">ยังไม่มีข้อมูลสิ่งอำนวยความสะดวกสำหรับสาขานี้</p>
      ) : (
        <ul className="branch-amenities-list">
          {list.map((amenity) => {
            const Icon = getAmenityIcon(amenity.icon);
            return (
              <li key={amenity.id}>
                <Icon width={20} height={20} />
                <span>{amenity.name}</span>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
