package braintree

import (
	"testing"
	"time"
)

func TestSettlementBatch(t *testing.T) {
	t.Parallel()

	// Create a new transaction
	tx, err := testGateway.Transaction().Create(&Transaction{
		Type:               "sale",
		Amount:             NewDecimal(1000, 2),
		PaymentMethodNonce: FakeNonceTransactableJCB,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("transaction : %s : %s : %s : %s\n", tx.MerchantAccountId, tx.Id, tx.CreditCard.CardType, tx.Status)
	if tx.Status != "authorized" {
		t.Fatal(tx.Status)
	}

	// Submit for settlement
	tx, err = testGateway.Transaction().SubmitForSettlement(tx.Id, tx.Amount)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("transaction : %s : %s : %s : %s\n", tx.MerchantAccountId, tx.Id, tx.CreditCard.CardType, tx.Status)
	if x := tx.Status; x != "submitted_for_settlement" {
		t.Fatal(x)
	}

	// Settle
	tx, err = testGateway.Testing().Settle(tx.Id)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("transaction : %s : %s : %s : %s\n", tx.MerchantAccountId, tx.Id, tx.CreditCard.CardType, tx.Status)
	if x := tx.Status; x != "settled" {
		t.Fatal(x)
	}

	// Generate Settlement Batch Summary which will include new transaction
	date := time.Now().Format("2006-01-02")
	summary, err := testGateway.Settlement().Generate(&Settlement{Date: date})
	if err != nil {
		t.Fatalf("unable to get settlement batch: %s", err)
	}

	var found bool
	for _, r := range summary.Records.Type {
		t.Logf("record      : %s : %22s : %4d : %6s : %8s\n", r.MerchantAccountId, r.CardType, r.Count, r.Kind, r.AmountSettled)
		if r.MerchantAccountId == tx.MerchantAccountId && r.CardType == tx.CreditCard.CardType && r.Count > 0 && r.Kind == "sale" {
			found = true
		}
	}

	if !found {
		t.Fatalf("Transaction %s created but no record in the settlement batch for it's merchant account and card type.", tx.Id)
	}
}