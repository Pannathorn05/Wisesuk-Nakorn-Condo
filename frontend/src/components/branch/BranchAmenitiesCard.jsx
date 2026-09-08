import { getAmenityIcon } from "../icons/amenityIconMap";
import { IconStar } from "../icons";
import "./BranchDetailCards.css";

export function BranchAmenitiesCard({ amenities }) {
  const list = amenities || [];

  return (
    <div className="branch-detail-card branch-detail-card--amenities">
      <h3>
        <IconStar width={18} height={18} /> สิ่งอำนวยความสะดวก
      </h3>
      {list.length === 0 ? (
        <p className="branch-detail-card__empty">ยังไม่มีข้อมูลสิ่งอำนวยความสะดวกสำหรับสาขานี้</p>
      ) : (
        <ul className="branch-amenities-list">
          {list.map((amenity) => {
            const Icon = getAmenityIcon(amenity.icon);
            return (
              <li key={amenity.id}>
                <span className="branch-amenities-list__icon">
                  <Icon width={22} height={22} />
                </span>
                <span>{amenity.name}</span>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
