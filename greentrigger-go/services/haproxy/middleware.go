package haproxy

import (
	"fmt"
)

// TransactionMiddleware is a middleware that manages transactions for HAProxy operations.
type TransactionMiddleware func(next func(transactionID string) error) func() error

// NewTransactionMiddleware creates a new instance of TransactionMiddleware.
func NewTransactionMiddleware(configManager *HAProxyConfigurationManager) TransactionMiddleware {
	return func(next func(transactionID string) error) func() error {
		return func() error {
			// Step 1: Retrieve the current HAProxy configuration version.
			cfgVer, err := configManager.GetCurrentConfigVersion()
			if err != nil {
				return fmt.Errorf("failed to retrieve configuration version: %v", err)
			}

			// Step 2: Start a transaction for modifying the configuration.
			transactionID, err := configManager.StartTransaction(cfgVer)
			if err != nil {
				return fmt.Errorf("failed to start transaction: %v", err)
			}

			defer func() {
				// Step 5: Commit or roll back the transaction
				if err != nil {
					configManager.RollbackTransaction(transactionID)
				} else {
					configManager.CommitTransaction(transactionID)
				}
			}()

			return next(transactionID)
		}
	}
}
