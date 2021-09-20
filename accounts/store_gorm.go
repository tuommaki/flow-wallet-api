package accounts

import (
	"github.com/flow-hydraulics/flow-wallet-api/datastore"
	"github.com/flow-hydraulics/flow-wallet-api/flow_helpers"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	db.AutoMigrate(&Account{})
	return &GormStore{db}
}

func (s *GormStore) Accounts(o datastore.ListOptions) (aa []Account, err error) {
	err = s.db.
		Order("created_at desc").
		Limit(o.Limit).
		Offset(o.Offset).
		Find(&aa).Error
	return
}

func (s *GormStore) Account(address string) (a Account, err error) {
	err = s.db.First(&a, "address = ?", flow_helpers.HexString(address)).Error
	return
}

func (s *GormStore) InsertAccount(a *Account) error {
	// Ensure unified address formatting
	a.Address = flow_helpers.HexString(a.Address)
	return s.db.Create(a).Error
}

func (s *GormStore) HardDeleteAccount(a *Account) error {
	// Ensure unified address formatting
	a.Address = flow_helpers.HexString(a.Address)
	return s.db.Unscoped().Delete(a).Error
}
