package models

import "time"

// Availability represents the availability information for a bus on a specific date
type Bus struct {
	ID          string    `json:"id" bson:"_id"`
	Bus_id      string    `json:"bus_id" bson:"bus_id"`
	Date        string    `json:"date" bson:"date"`
	SeatsTotal  int       `json:"seats_total" bson:"seats_total"`
	SeatsBooked int       `json:"seats_booked" bson:"seats_booked"`
	Created_at   time.Time `json:"created_at" bson:"created_at"`
	Updated_at   time.Time `json:"updated_at" bson:"updated_at"`
}
