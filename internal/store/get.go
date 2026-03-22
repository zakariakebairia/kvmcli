package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/zakariakebairia/kvmcli/internal/common"
	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// StoreInfo holds display information for a store.
type StoreInfo struct {
	Name          string
	Namespace     string
	Backend       string
	ArtifactsPath string
	ImagesPath    string
	ImageCount    int
	Age           string
}

// Header returns a tabwriter with column headers for store listing.
func (info *StoreInfo) Header() *tabwriter.Writer {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tNAMESPACE\tBACKEND\tARTIFACTS PATH\tIMAGES PATH\tIMAGES\tAGE")
	return w
}

// PrintInfo writes the store info as a row to the tabwriter.
func (info *StoreInfo) PrintInfo(w *tabwriter.Writer) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
		info.Name,
		info.Namespace,
		info.Backend,
		info.ArtifactsPath,
		info.ImagesPath,
		info.ImageCount,
		info.Age,
	)
}

// NewStoreInfo constructs a StoreInfo from a database record.
func NewStoreInfo(
	ctx context.Context,
	database *sql.DB,
	record db.Store,
) (*StoreInfo, error) {
	// Count images for this store
	imageCount := len(record.Images)

	return &StoreInfo{
		Name:          record.Name,
		Namespace:     record.Namespace,
		Backend:       record.Backend,
		ArtifactsPath: record.ArtifactsPath,
		ImagesPath:    record.ImagesPath,
		ImageCount:    imageCount,
		Age:           common.FormatAge(record.CreatedAt),
	}, nil
}

// GetStores retrieves all stores from the database and returns StoreInfo slices.
func GetStores(
	ctx context.Context,
	database *sql.DB,
) ([]StoreInfo, error) {
	records, err := db.GetStores(ctx, database, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get store records: %w", err)
	}

	stores := make([]StoreInfo, 0, len(records))

	for _, rec := range records {
		storeInfo, err := NewStoreInfo(ctx, database, rec)
		if err != nil {
			continue
		}
		stores = append(stores, *storeInfo)
	}

	return stores, nil
}
