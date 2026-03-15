package router

import "github.com/hegner123/modulacms/internal/db"

// MediaResponse wraps db.Media with computed fields for API responses.
type MediaResponse struct {
	db.Media
	DownloadURL string `json:"download_url"`
}

// toMediaResponse adds the download_url field to a media record.
func toMediaResponse(m db.Media) MediaResponse {
	return MediaResponse{
		Media:       m,
		DownloadURL: "/api/v1/media/" + string(m.MediaID) + "/download",
	}
}

// toMediaListResponse wraps a slice of media records.
func toMediaListResponse(items []db.Media) []MediaResponse {
	resp := make([]MediaResponse, len(items))
	for i, m := range items {
		resp[i] = toMediaResponse(m)
	}
	return resp
}
