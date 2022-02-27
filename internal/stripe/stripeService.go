package stripe

import (
	"fmt"
)

type StripeStore interface {
	UserSetStripeId(string, string)
	RenewSubscription(string)
	UserSetStripeSession(int, string)
}

type StripeService struct {
	store StripeStore
}

func NewService(s StripeStore) *StripeService {
	return &StripeService{
		store: s,
	}
}

func (s *StripeService) SetSession(userID int, sessionID string) {
	s.store.UserSetStripeSession(userID, sessionID)
}

func (s *StripeService) HandleEvent(e string, object map[string]interface{}) {
	switch e {
	case "customer.created":
		s.store.UserSetStripeId(fmt.Sprint(object["email"]), fmt.Sprint(object["id"]))
	case "payment_intent.succeeded":
		s.store.RenewSubscription(fmt.Sprint(object["customer"]))
	}
}
