package snapshotv1

import "context"

// Store defines the interface for storing and loading snapshots of the order book.
//
//go:generate mockgen -source interface.go -destination=mock/interface_mock.go -package=snapshotv1_mock
type Store interface {
	Store(ctx context.Context, snapshot *Snapshot) error
	LoadStore(ctx context.Context) (*Snapshot, error)
}
