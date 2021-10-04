package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/flow-hydraulics/flow-wallet-api/configs"
	gormstore "github.com/flow-hydraulics/flow-wallet-api/datastore/gorm"
	"github.com/flow-hydraulics/flow-wallet-api/tests/internal/test"
	"github.com/flow-hydraulics/flow-wallet-api/transactions"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
)

func Test_VerifyEmbeddedTransactionFields(t *testing.T) {
	tStart := time.Now()

	testCases := []struct {
		name     string
		tx       transactions.Transaction
		expected transactions.Transaction
	}{
		{
			name: "insert barebone transaction without code",
			tx: transactions.Transaction{
				TransactionId:   "0xf00ba4",
				TransactionType: transactions.General,
				ProposerAddress: "0xbeeff00d",
				CreatedAt:       tStart,
			},
			expected: transactions.Transaction{
				TransactionId:   "0xf00ba4",
				TransactionType: transactions.General,
				Arguments:       []transactions.TransactionArgument{},
				ProposerAddress: "0xbeeff00d",
				CreatedAt:       tStart,
			},
		},
		{
			name: "insert transaction with code and arguments",
			tx: transactions.Transaction{
				TransactionId:   "0xf00ba4",
				TransactionType: transactions.General,
				Code: transactions.TransactionCode{
					TransactionId: "0xf00ba4",
					Code:          "transaction(greeting: String) { prepare(signer: AuthAccount){} execute { log(greeting.concat(\", World!\")) }}",
				},
				Arguments: []transactions.TransactionArgument{
					{
						TransactionId: "0xf00ba4",
						Type:          "String",
						Value:         "Hello",
					},
				},
				ProposerAddress: "0xbeeff00d",
				CreatedAt:       tStart,
			},
			expected: transactions.Transaction{
				TransactionId:   "0xf00ba4",
				TransactionType: transactions.General,
				Code: transactions.TransactionCode{
					TransactionId: "0xf00ba4",
					Code:          "transaction(greeting: String) { prepare(signer: AuthAccount){} execute { log(greeting.concat(\", World!\")) }}",
				},
				Arguments: []transactions.TransactionArgument{
					{
						TransactionId: "0xf00ba4",
						Type:          "String",
						Value:         "Hello",
					},
				},
				ProposerAddress: "0xbeeff00d",
				CreatedAt:       tStart,
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			cfg := test.LoadConfig(t, testConfigPath)
			db := openDB(t, cfg)
			db.AutoMigrate(&transactions.Transaction{}, &transactions.TransactionCode{}, &transactions.TransactionArgument{})
			store := transactions.NewGormStore(db)

			store.InsertTransaction(&tc.tx)

			tx, err := store.Transaction(tc.tx.TransactionId)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tx, tc.expected, cmpopts.IgnoreFields(transactions.Transaction{}, "UpdatedAt")) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.expected, tx))
			}
		})
	}
}

func openDB(t *testing.T, cfg *configs.Config) *gorm.DB {
	t.Helper()

	db, err := gormstore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dbClose := func() { gormstore.Close(db) }
	t.Cleanup(dbClose)

	return db
}
