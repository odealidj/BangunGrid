package http

type OrderItemRequest struct {
	SKUID       string  `json:"sku_id"`
	ProductName string  `json:"product_name"`
	UOMCode     string  `json:"uom_code"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	WeightKg    float64 `json:"weight_kg"`
	VolumeM3    float64 `json:"volume_m3"`
}

type EstimateOrderRequest struct {
	Items            []OrderItemRequest `json:"items"`
	IncludeUnloading bool               `json:"include_unloading"`
	NarrowRoadAccess bool               `json:"narrow_road_access"`
}

type CreateOrderRequest struct {
	BranchID          string             `json:"branch_id"`
	CustomerID        string             `json:"customer_id,omitempty"`
	CustomerType      string             `json:"customer_type"` // B2C_RETAIL, B2B_CONTRACTOR
	DeliveryAddress   string             `json:"delivery_address"`
	DeliveryLatitude  float64            `json:"delivery_latitude"`
	DeliveryLongitude float64            `json:"delivery_longitude"`
	IncludeUnloading  bool               `json:"include_unloading"`
	NarrowRoadAccess  bool               `json:"narrow_road_access"`
	ShippingFee       float64            `json:"shipping_fee"`
	Notes             string             `json:"notes,omitempty"`
	Items             []OrderItemRequest `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

type RFQItemRequest struct {
	MaterialName string  `json:"material_name"`
	TargetUOM    string  `json:"target_uom"`
	EstimatedQty float64 `json:"estimated_qty"`
	TargetPrice  float64 `json:"target_price,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

type CreateRFQRequest struct {
	ProjectName     string           `json:"project_name"`
	ProjectAddress  string           `json:"project_address"`
	BoQDocumentURL  string           `json:"boq_document_url,omitempty"`
	EstimatedBudget float64          `json:"estimated_budget"`
	Notes           string           `json:"notes,omitempty"`
	Items           []RFQItemRequest `json:"items"`
}

type UpdateRFQQuoteRequest struct {
	Status         string  `json:"status"` // QUOTED, ACCEPTED, REJECTED
	QuotedAmount   float64 `json:"quoted_amount"`
	QuotationNotes string  `json:"quotation_notes,omitempty"`
}

type CallOffItemRequest struct {
	SKUID       string  `json:"sku_id"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	UOM         string  `json:"uom"`
}

type CreateCallOffRequest struct {
	SequenceNo            int                  `json:"sequence_no"`
	RequestedDeliveryDate string               `json:"requested_delivery_date"`
	Notes                 string               `json:"notes,omitempty"`
	Items                 []CallOffItemRequest `json:"items"`
}

type UpdateCallOffStatusRequest struct {
	Status string `json:"status"` // SCHEDULED, DISPATCHED, COMPLETED, CANCELLED
}
