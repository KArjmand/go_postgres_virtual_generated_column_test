package invoices

import "math"

// ID represents an invoice identifier
type ID int64

// Invoice represents an invoice entity in the domain
type Invoice struct {
	id          ID
	customerID  int64
	amountCents int64
	taxRate     float64
	totalCents  int64
}

// NewInvoice creates a new Invoice with pre-computed total
func NewInvoice(id ID, customerID, amountCents int64, taxRate float64, totalCents int64) *Invoice {
	return &Invoice{
		id:          id,
		customerID:  customerID,
		amountCents: amountCents,
		taxRate:     taxRate,
		totalCents:  totalCents,
	}
}

// NewInvoiceWithCalculation creates a new Invoice and calculates total in Go
func NewInvoiceWithCalculation(id ID, customerID, amountCents int64, taxRate float64) *Invoice {
	totalCents := int64(math.Round(float64(amountCents) * (1 + taxRate)))
	return &Invoice{
		id:          id,
		customerID:  customerID,
		amountCents: amountCents,
		taxRate:     taxRate,
		totalCents:  totalCents,
	}
}

// ID returns the invoice ID
func (i *Invoice) ID() ID {
	return i.id
}

// CustomerID returns the customer ID
func (i *Invoice) CustomerID() int64 {
	return i.customerID
}

// AmountCents returns the amount in cents
func (i *Invoice) AmountCents() int64 {
	return i.amountCents
}

// TaxRate returns the tax rate
func (i *Invoice) TaxRate() float64 {
	return i.taxRate
}

// TotalCents returns the total amount in cents
func (i *Invoice) TotalCents() int64 {
	return i.totalCents
}
