package models

import (
	"time"
	"github.com/teris-io/shortid"
	
	"gorm.io/gorm"
)

// Order represents the customer's overall order in the DB
type Order struct {
	ID           string      `gorm:"primaryKey;size:14" json:"id"`
	CustomerName string      `gorm:"not null" json:"customerName"`
	Status       string      `gorm:"not null" json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	Items        []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
}

// OrderItem represents a single pizza within an Order
type OrderItem struct {
	ID      string `gorm:"primaryKey;size:14" json:"id"`
	OrderID string `gorm:"index;size:14;not null" json:"orderId"`
	Pizza   string `gorm:"not null" json:"pizza"`
	Size    string `gorm:"not null" json:"size"`
}


func (o *Order) BeforeCreate(tx *gorm.DB) error {
	id, err := shortid.Generate()
	if err != nil {
		return err
	}
	o.ID = id
	return nil
}


func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	id, err := shortid.Generate()
	if err != nil {
		return err
	}
	oi.ID = id
	return nil
}