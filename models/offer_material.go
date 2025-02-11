package models

import "time"

type OfferMaterial struct {
    ID           int       `json:"id"`
    OfferID      int       `json:"offer_id"`
    MaterialID   int       `json:"material_id"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    *time.Time `json:"deleted_at"`
}
