package datastore

import (
	"log"
	"time"
)

func (ds *Datastore) UserSetStripeId(email, stripeCustomerId string) {
	query := `
		UPDATE users
		SET stripe_customer_id = ?
		WHERE email = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(stripeCustomerId, email)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}

func (ds *Datastore) UserSetStripeSession(id int, stripeSessionId string) {
	query := `
		UPDATE users
		SET stripe_session_id = ?
		WHERE id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(stripeSessionId, id)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}

func (ds *Datastore) RenewSubscription(stripeCustomerId string) {
	query := `
		UPDATE users
		SET period_end = ?
		WHERE stripe_customer_id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	periodEnd := time.Now().AddDate(0, 1, 2)

	_, err := stmt.Exec(periodEnd, stripeCustomerId)
	if err != nil {
		log.Printf("userUpgrade: %v\n", err)
	}
}
