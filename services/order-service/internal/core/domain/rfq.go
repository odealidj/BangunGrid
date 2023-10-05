package domain

import (
	"time"

	"github.com/google/uuid"
)

type RFQStatus string

const (
	RFQStatusSubmitted RFQStatus = "SUBMITTED"
	RFQStatusInReview  RFQStatus = "IN_REVIEW"
	RFQStatusQuoted    RFQStatus = "QUOTED"
	RFQStatusAccepted  RFQStatus = "ACCEPTED"
	RFQStatusRejected  RFQStatus = "REJECTED"
)

type RFQItem struct {
	ID             uuid.UUID `json:"id"`
	RFQID          uuid.UUID `json:"rfq_id"`
	MaterialName   string    `json:"material_name"`
	TargetUOM      string    `json:"target_uom"`
	EstimatedQty   float64   `json:"estimated_qty"`
	TargetPrice    float64   `json:"target_price,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

type B2BRFQ struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	CustomerID      uuid.UUID `json:"customer_id"`
	ProjectName     string    `json:"project_name"`
	ProjectAddress  string    `json:"project_address"`
	BoQDocumentURL  string    `json:"boq_document_url,omitempty"`
	EstimatedBudget float64   `json:"estimated_budget"`
	Status          RFQStatus `json:"status"`
	Notes           string    `json:"notes,omitempty"`
	QuotedAmount    float64   `json:"quoted_amount,omitempty"`
	QuotationNotes  string    `json:"quotation_notes,omitempty"`
	Items           []RFQItem `json:"items"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewB2BRFQ(
	tenantID uuid.UUID,
	customerID uuid.UUID,
	projectName string,
	projectAddress string,
	boqURL string,
	estimatedBudget float64,
	notes string,
) *B2BRFQ {
	now := time.Now().UTC()
	return &B2BRFQ{
		ID:              uuid.New(),
		TenantID:        tenantID,
		CustomerID:      customerID,
		ProjectName:     projectName,
		ProjectAddress:  projectAddress,
		BoQDocumentURL:  boqURL,
		EstimatedBudget: estimatedBudget,
		Status:          RFQStatusSubmitted,
		Notes:           notes,
		Items:           make([]RFQItem, 0),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (r *B2BRFQ) AddItem(materialName, targetUOM string, estimatedQty, targetPrice float64, notes string) {
	r.Items = append(r.Items, RFQItem{
		ID:           uuid.New(),
		RFQID:        r.ID,
		MaterialName: materialName,
		TargetUOM:    targetUOM,
		EstimatedQty: estimatedQty,
		TargetPrice:  targetPrice,
		Notes:        notes,
	})
	r.UpdatedAt = time.Now().UTC()
}

func (r *B2BRFQ) UpdateQuote(amount float64, notes string) {
	r.Status = RFQStatusQuoted
	r.QuotedAmount = amount
	r.QuotationNotes = notes
	r.UpdatedAt = time.Now().UTC()
}
