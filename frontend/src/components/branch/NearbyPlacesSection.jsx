import { groupNearbyPlaces } from "../../utils/nearbyPlaces";
import "./NearbyPlacesSection.css";

export function NearbyPlacesSection({ nearbyPlaces }) {
  const groups = groupNearbyPlaces(nearbyPlaces);

  return (
    <div className="nearby-section">
      <h3>สถานที่ใกล้เคียง</h3>

      {groups.length === 0 ? (
        <p className="nearby-section__empty">ยังไม่มีข้อมูลสถานที่ใกล้เคียงสำหรับสาขานี้</p>
      ) : (
        <div className="nearby-section__groups">
          {groups.map((group) => (
            <div className="nearby-group" key={group.category}>
              <h4>{group.label}</h4>
              <div className="nearby-group__grid">
                {group.places.map((place) => (
                  <div className="nearby-group__row" key={place.id}>
                    <span>{place.name}</span>
                    <span className="nearby-group__distance">{place.distance}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
