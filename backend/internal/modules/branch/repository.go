package branch

import (
	"context"

	"github.com/google/uuid"

	"backend/internal/database"
)

type Repository struct{ db *database.TxManager }

func NewRepository(db *database.TxManager) *Repository { return &Repository{db: db} }

const columns = `
	id, slug, name, tagline, description, address, phones, line_id, email,
	latitude, longitude, map_url, building_count, floor_count,
	daily_price_from, monthly_price_min, monthly_price_max,
	water_rate, electric_rate, deposit, advance_payment, contract_fee,
	cover_image_url, is_active, created_at, updated_at`

func scan(row interface{ Scan(...any) error }) (*Branch, error) {
	var b Branch
	err := row.Scan(
		&b.ID, &b.Slug, &b.Name, &b.Tagline, &b.Description, &b.Address, &b.Phones, &b.LineID, &b.Email,
		&b.Latitude, &b.Longitude, &b.MapURL, &b.BuildingCount, &b.FloorCount,
		&b.DailyPriceFrom, &b.MonthlyPriceMin, &b.MonthlyPriceMax,
		&b.WaterRate, &b.ElectricRate, &b.Deposit, &b.AdvancePayment, &b.ContractFee,
		&b.CoverImageURL, &b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, database.NormalizeErr(err)
	}
	return &b, nil
}

func (r *Repository) List(ctx context.Context, includeInactive bool) ([]Branch, error) {
	q := `SELECT ` + columns + ` FROM branches WHERE ($1 OR is_active) ORDER BY created_at`
	rows, err := r.db.Executor(ctx).Query(ctx, q, includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Branch{}
	for rows.Next() {
		b, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *b)
	}
	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Branch, error) {
	q := `SELECT ` + columns + ` FROM branches WHERE id = $1`
	return scan(r.db.Executor(ctx).QueryRow(ctx, q, id))
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Branch, error) {
	q := `SELECT ` + columns + ` FROM branches WHERE slug = $1`
	return scan(r.db.Executor(ctx).QueryRow(ctx, q, slug))
}

func (r *Repository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.Executor(ctx).QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM branches WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

type UpdateParams struct {
	Name            string
	Tagline         string
	Description     string
	Address         string
	Phones          []string
	LineID          string
	Email           string
	Latitude        *float64
	Longitude       *float64
	MapURL          string
	BuildingCount   int
	FloorCount      int
	DailyPriceFrom  *float64
	MonthlyPriceMin *float64
	MonthlyPriceMax *float64
	WaterRate       float64
	ElectricRate    float64
	Deposit         float64
	AdvancePayment  float64
	ContractFee     float64
	CoverImageURL   *string
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, p UpdateParams) error {
	const q = `
		UPDATE branches SET
			name = $2, tagline = $3, description = $4, address = $5, phones = $6,
			line_id = $7, email = $8, latitude = $9, longitude = $10, map_url = $11,
			building_count = $12, floor_count = $13,
			daily_price_from = $14, monthly_price_min = $15, monthly_price_max = $16,
			water_rate = $17, electric_rate = $18, deposit = $19,
			advance_payment = $20, contract_fee = $21,
			cover_image_url = COALESCE($22, cover_image_url),
			updated_at = now()
		WHERE id = $1`
	tag, err := r.db.Executor(ctx).Exec(ctx, q, id,
		p.Name, p.Tagline, p.Description, p.Address, p.Phones, p.LineID, p.Email,
		p.Latitude, p.Longitude, p.MapURL, p.BuildingCount, p.FloorCount,
		p.DailyPriceFrom, p.MonthlyPriceMin, p.MonthlyPriceMax,
		p.WaterRate, p.ElectricRate, p.Deposit, p.AdvancePayment, p.ContractFee,
		p.CoverImageURL,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- images

func (r *Repository) ListImages(ctx context.Context, branchID uuid.UUID) ([]BranchImage, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT id, branch_id, image_url, caption, sort_order
		 FROM branch_images WHERE branch_id = $1 ORDER BY sort_order, created_at`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []BranchImage{}
	for rows.Next() {
		var img BranchImage
		if err := rows.Scan(&img.ID, &img.BranchID, &img.ImageURL, &img.Caption, &img.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (r *Repository) AddImage(ctx context.Context, branchID uuid.UUID, url, caption string, sortOrder int) (*BranchImage, error) {
	var img BranchImage
	err := r.db.Executor(ctx).QueryRow(ctx,
		`INSERT INTO branch_images (branch_id, image_url, caption, sort_order)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, branch_id, image_url, caption, sort_order`,
		branchID, url, caption, sortOrder,
	).Scan(&img.ID, &img.BranchID, &img.ImageURL, &img.Caption, &img.SortOrder)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// DeleteImage จำกัดด้วย branch_id เพื่อกันแอดมินสาขาหนึ่งลบรูปของอีกสาขา
func (r *Repository) DeleteImage(ctx context.Context, branchID, imageID uuid.UUID) error {
	tag, err := r.db.Executor(ctx).Exec(ctx,
		`DELETE FROM branch_images WHERE id = $1 AND branch_id = $2`, imageID, branchID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// ---------------------------------------------------------------- amenities

func (r *Repository) ListAllAmenities(ctx context.Context) ([]Amenity, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT id, code, name, icon, sort_order FROM amenities ORDER BY sort_order, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAmenities(rows)
}

func (r *Repository) ListAmenities(ctx context.Context, branchID uuid.UUID) ([]Amenity, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT a.id, a.code, a.name, a.icon, a.sort_order
		 FROM amenities a
		 JOIN branch_amenities ba ON ba.amenity_id = a.id
		 WHERE ba.branch_id = $1
		 ORDER BY a.sort_order, a.name`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectAmenities(rows)
}

// SetAmenities แทนที่รายการสิ่งอำนวยความสะดวกของสาขาทั้งชุด
func (r *Repository) SetAmenities(ctx context.Context, branchID uuid.UUID, amenityIDs []uuid.UUID) error {
	exec := r.db.Executor(ctx)
	if _, err := exec.Exec(ctx, `DELETE FROM branch_amenities WHERE branch_id = $1`, branchID); err != nil {
		return err
	}
	if len(amenityIDs) == 0 {
		return nil
	}
	_, err := exec.Exec(ctx,
		`INSERT INTO branch_amenities (branch_id, amenity_id) SELECT $1, unnest($2::uuid[])`,
		branchID, amenityIDs)
	return err
}

func collectAmenities(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]Amenity, error) {
	out := []Amenity{}
	for rows.Next() {
		var a Amenity
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.Icon, &a.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------- nearby places

func (r *Repository) ListNearby(ctx context.Context, branchID uuid.UUID) ([]NearbyPlace, error) {
	rows, err := r.db.Executor(ctx).Query(ctx,
		`SELECT id, branch_id, category, name, distance, sort_order
		 FROM nearby_places WHERE branch_id = $1 ORDER BY category, sort_order`, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []NearbyPlace{}
	for rows.Next() {
		var p NearbyPlace
		if err := rows.Scan(&p.ID, &p.BranchID, &p.Category, &p.Name, &p.Distance, &p.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type NearbyInput struct {
	Category  string `json:"category"`
	Name      string `json:"name"`
	Distance  string `json:"distance"`
	SortOrder int    `json:"sort_order"`
}

// ReplaceNearby เขียนทับรายการสถานที่ใกล้เคียงทั้งชุด (ตรงกับ UI ที่แก้ทีละกลุ่ม)
func (r *Repository) ReplaceNearby(ctx context.Context, branchID uuid.UUID, items []NearbyInput) error {
	exec := r.db.Executor(ctx)
	if _, err := exec.Exec(ctx, `DELETE FROM nearby_places WHERE branch_id = $1`, branchID); err != nil {
		return err
	}
	for i, it := range items {
		order := it.SortOrder
		if order == 0 {
			order = i
		}
		if _, err := exec.Exec(ctx,
			`INSERT INTO nearby_places (branch_id, category, name, distance, sort_order)
			 VALUES ($1, $2, $3, $4, $5)`,
			branchID, it.Category, it.Name, it.Distance, order); err != nil {
			return err
		}
	}
	return nil
}

// LoadRelations เติม images / amenities / nearby ให้ branch ที่ดึงมาแล้ว
func (r *Repository) LoadRelations(ctx context.Context, b *Branch) error {
	var err error
	if b.Images, err = r.ListImages(ctx, b.ID); err != nil {
		return err
	}
	if b.Amenities, err = r.ListAmenities(ctx, b.ID); err != nil {
		return err
	}
	if b.Nearby, err = r.ListNearby(ctx, b.ID); err != nil {
		return err
	}
	return nil
}
