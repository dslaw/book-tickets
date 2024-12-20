package payment

import "github.com/dslaw/book-tickets/pkg/entities"

type Card struct {
	Name            string
	Address         string
	Number          string
	ExpirationMonth uint8
	ExpirationYear  uint8
	CVC             string
}

// This, and "implemented" methods, are a stub/placeholder.
type PaymentClient struct{}

// SubmitPayment submits a user's payment for a ticket to a third-party service,
// and returns a boolean indicating whether the payment was accepted or not.
func (svc *PaymentClient) SubmitPayment(ticket entities.Ticket, card Card) (bool, error) {
	return true, nil
}
