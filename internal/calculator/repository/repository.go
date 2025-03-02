package repository

import (
	"github.com/cockroachdb/pebble"
	"github.com/rs/xid"
)

type Repository struct {
	db *pebble.DB
}

func New(db *pebble.DB) *Repository {
	_ = xid.New()
	return &Repository{db: db}
}
