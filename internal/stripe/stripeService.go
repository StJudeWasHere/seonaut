package stripe

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

func (s *StripeService) SetId(email, customerID string) {
	s.store.UserSetStripeId(email, customerID)
}

func (s *StripeService) Renew(customerID string) {
	s.store.RenewSubscription(customerID)
}

func (s *StripeService) SetSession(userID int, sessionID string) {
	s.store.UserSetStripeSession(userID, sessionID)
}
