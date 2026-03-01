package db

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// MediaDimensionsHistoryEntry represents a media dimension history record.
type MediaDimensionsHistoryEntry struct {
	MdID        string              `json:"md_id"`
	Label       NullString          `json:"label"`
	Width       types.NullableInt64 `json:"width"`
	Height      types.NullableInt64 `json:"height"`
	AspectRatio NullString          `json:"aspect_ratio"`
}

// CreateMediaDimensionFormParams contains form parameters for creating a media dimension.
type CreateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

// UpdateMediaDimensionFormParams contains form parameters for updating a media dimension.
type UpdateMediaDimensionFormParams struct {
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
	MdID        string `json:"md_id"`
}

// MediaDimensionsJSON represents a media dimension as JSON.
type MediaDimensionsJSON struct {
	MdID        string              `json:"md_id"`
	Label       NullString          `json:"label"`
	Width       types.NullableInt64 `json:"width"`
	Height      types.NullableInt64 `json:"height"`
	AspectRatio NullString          `json:"aspect_ratio"`
}

// CreateMediaDimensionParamsJSON contains JSON parameters for creating a media dimension.
type CreateMediaDimensionParamsJSON struct {
	Label       NullString          `json:"label"`
	Width       types.NullableInt64 `json:"width"`
	Height      types.NullableInt64 `json:"height"`
	AspectRatio NullString          `json:"aspect_ratio"`
}

// UpdateMediaDimensionParamsJSON contains JSON parameters for updating a media dimension.
type UpdateMediaDimensionParamsJSON struct {
	Label       NullString          `json:"label"`
	Width       types.NullableInt64 `json:"width"`
	Height      types.NullableInt64 `json:"height"`
	AspectRatio NullString          `json:"aspect_ratio"`
	MdID        string              `json:"md_id"`
}

// MapCreateMediaDimensionParams converts form parameters to database parameters.
func MapCreateMediaDimensionParams(a CreateMediaDimensionFormParams) CreateMediaDimensionParams {
	w := StringToNullInt64(a.Width)
	h := StringToNullInt64(a.Height)
	return CreateMediaDimensionParams{
		Label:       NewNullString(a.Label),
		Width:       types.NullableInt64{Int64: w.Int64, Valid: w.Valid},
		Height:      types.NullableInt64{Int64: h.Int64, Valid: h.Valid},
		AspectRatio: NewNullString(a.AspectRatio),
	}
}

// MapUpdateMediaDimensionParams converts form parameters to database parameters.
func MapUpdateMediaDimensionParams(a UpdateMediaDimensionFormParams) UpdateMediaDimensionParams {
	w := StringToNullInt64(a.Width)
	h := StringToNullInt64(a.Height)
	return UpdateMediaDimensionParams{
		Label:       NewNullString(a.Label),
		Width:       types.NullableInt64{Int64: w.Int64, Valid: w.Valid},
		Height:      types.NullableInt64{Int64: h.Int64, Valid: h.Valid},
		AspectRatio: NewNullString(a.AspectRatio),
		MdID:        a.MdID,
	}
}

// MapCreateMediaDimensionJSONParams converts JSON parameters to database parameters.
func MapCreateMediaDimensionJSONParams(a CreateMediaDimensionParamsJSON) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}
}

// MapUpdateMediaDimensionJSONParams converts JSON parameters to database parameters.
func MapUpdateMediaDimensionJSONParams(a UpdateMediaDimensionParamsJSON) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}
