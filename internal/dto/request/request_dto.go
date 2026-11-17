package request

type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
