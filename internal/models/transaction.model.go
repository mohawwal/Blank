package models

import "gorm.io/gorm"

type TransactionModel struct {
	gorm.Model
	UserID          uint   `gorm:"not null" json:"user_id"`
	Type            string `gorm:"size:20" json:"type"`             // airtime, data
	Network         string `gorm:"size:20" json:"network"`          // mtn, glo, airtel, 9mobile
	Amount          int    `gorm:"not null" json:"amount"`          // Amount in Naira
	PhoneNumber     string `gorm:"size:20" json:"phone_number"`     // Recipient phone number
	VariationCode   string `gorm:"size:50" json:"variation_code"`   // For data bundles
	Status          string `gorm:"size:20" json:"status"`           // pending, confirmed, processing, completed, failed
	VTPassRequestID string `gorm:"size:100" json:"vtpass_request_id"` // Unique ID for VTPass
	VTPassTxnID     string `gorm:"size:100" json:"vtpass_txn_id"`   // VTPass transaction ID
	PaystackRef     string `gorm:"size:100" json:"paystack_ref"`    // Paystack charge reference
}

func (receiver TransactionModel) TableName() string {
	return "transactions"
}
