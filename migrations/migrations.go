package migrations

import (
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20210922"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211004"
	"github.com/flow-hydraulics/flow-wallet-api/migrations/internal/m20211005"
	"github.com/go-gormigrate/gormigrate/v2"
)

func List() []*gormigrate.Migration {
	ms := []*gormigrate.Migration{
		&gormigrate.Migration{
			ID:       m20210922.ID,
			Migrate:  m20210922.Migrate,
			Rollback: m20210922.Rollback,
		},
		&gormigrate.Migration{
			ID:       m20211004.ID,
			Migrate:  m20211004.Migrate,
			Rollback: m20211004.Rollback,
		},
		&gormigrate.Migration{
			ID:       m20211005.ID,
			Migrate:  m20211005.Migrate,
			Rollback: m20211005.Rollback,
		},
	}
	return ms
}
