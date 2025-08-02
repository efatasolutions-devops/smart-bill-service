package models

// SplitbillResponse represents the response structure for splitbill API
type SplitbillResponse struct {
	Items            []Item           `json:"items"`
	StoreInformation StoreInformation `json:"store_information"`
	Totals           Totals           `json:"totals"`
	TransactionInfo  TransactionInfo  `json:"transaction_information"`
}

// Item represents an individual item in the receipt
type Item struct {
	Name     string `json:"name" example:"Nasi Goreng"`
	Price    string `json:"price" example:"25000.00"`
	Quantity string `json:"quantity" example:"2"`
	Total    string `json:"total" example:"50000.00"`
}

// StoreInformation represents store details from the receipt
type StoreInformation struct {
	Address     string `json:"address" example:"Jl. Sudirman No. 123, Jakarta"`
	Email       string `json:"email" example:"info@restaurant.com"`
	NPWP        string `json:"npwp" example:"12.345.678.9-012.345"`
	PhoneNumber string `json:"phone_number" example:"+62812345678"`
	StoreName   string `json:"store_name" example:"Restaurant ABC"`
}

// Totals represents the total calculation from the receipt
type Totals struct {
	Change   string `json:"change" example:"5000.00"`
	Discount string `json:"discount" example:"0.00"`
	Payment  string `json:"payment" example:"105000.00"`
	Subtotal string `json:"subtotal" example:"95000.00"`
	Tax      Tax    `json:"tax"`
	Total    string `json:"total" example:"100000.00"`
}

// Tax represents tax information from the receipt
type Tax struct {
	Amount        string `json:"amount" example:"5000.00"`
	ServiceCharge string `json:"service_charge" example:"0.00"`
	DPP           string `json:"dpp" example:"95000.00"`
	Name          string `json:"name" example:"PPN"`
	TotalTax      string `json:"total_tax" example:"5000.00"`
}

// TransactionInfo represents transaction details from the receipt
type TransactionInfo struct {
	Date          string `json:"date" example:"02/08/2025"`
	Time          string `json:"time" example:"19:30"`
	TransactionID string `json:"transaction_id" example:"TXN123456789"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Data   string `json:"data"`
	Status string `json:"status" example:"Error uploading image"`
}

// SuccessResponse represents a success response wrapper
type SuccessResponse struct {
	Data SplitbillResponse `json:"data"`
}
