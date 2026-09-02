package room

import "time"

type CreateRequest struct {
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	Floor       *int     `json:"floor"`
	Capacity    int      `json:"capacity"`
	Equipment   []string `json:"equipment"`
	Description *string  `json:"description"`
}

type RoomResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	Floor       *int     `json:"floor,omitempty"`
	Capacity    int      `json:"capacity"`
	Equipment   []string `json:"equipment"`
	Description *string  `json:"description,omitempty"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func NewRoomResponse(room Room) RoomResponse {
	return RoomResponse{
		ID:          room.ID.String(),
		Name:        room.Name,
		Location:    room.Location,
		Floor:       room.Floor,
		Capacity:    room.Capacity,
		Equipment:   room.Equipment,
		Description: room.Description,
		IsActive:    room.IsActive,
		CreatedAt:   room.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   room.UpdatedAt.Format(time.RFC3339),
	}
}
