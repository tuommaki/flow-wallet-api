package transactions

import (
	"github.com/onflow/flow-go-sdk"
)

func NewFlowTransaction(code string, args []Argument) (*flow.Transaction, error) {
	tx := flow.NewTransaction()
	tx.SetScript([]byte(code))

	// Add arguments
	for _, a := range args {
		c, err := ArgAsCadence(a)
		if err != nil {
			return nil, err
		}
		if err := tx.AddArgument(c); err != nil {
			return nil, err
		}
	}

	return tx, nil
}
