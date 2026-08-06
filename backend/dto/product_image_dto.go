package dto

type ProductImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
	SortOrder int    `json:"sortOrder"`
	Width     *int   `json:"width,omitempty"`
	Height    *int   `json:"height,omitempty"`
	Format    string `json:"format,omitempty"`
}

type ReorderProductImagesRequest struct {
	ImageIDs []uint `json:"imageIDs" binding:"required,min=1"`
}
