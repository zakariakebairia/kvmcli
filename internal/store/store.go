package store

import (
	"context"
	"errors"
	"time"

	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// TODO: 1. Delete function for store
//       2. Index for store table on database
//       3. Print function for store (kvmcli get sotre)
// 			 4. Connect stores with other instances using labels

// Errors returned by this package.
var (
	// ErrStoreNameEmpty is returned when the store name is not set.
	ErrStoreNameEmpty = errors.New("store: name is empty")
	// ErrStoreExist is returned when attempting to insert a store that already exists.
	ErrStoreExist = errors.New("store: store already exists")
	// ErrNilDBConn is returned when the database connection is missing.
	ErrNilDBConn = errors.New("store: database connection is nil")
)

// Store represents a bound store resource (configuration + manager).
// It implements resources.Resource (if needed, though Store resource interface requirements might differ?
// Resource interface requires Create/Delete/Start. Store had Start placeholder.
type Store struct {
	Spec    Config
	ctx     context.Context
	manager StoreManager
}

// NewStore creates a new Store resource.
func NewStore(spec Config, manager StoreManager, ctx context.Context) *Store {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Store{
		Spec:    spec,
		manager: manager,
		ctx:     ctx,
	}
}

// Create delegates to the manager.
func (s *Store) Create() error {
	return s.manager.Create(s.ctx, s.Spec)
}

// Delete delegates to the manager.
func (s *Store) Delete() error {
	return s.manager.Delete(s.ctx, s.Spec.Name, s.Spec.Namespace)
}

// Start delegates or does nothing (store doesn't really start).
func (s *Store) Start() error {
	// Stores don't need starting in this context usually, but to satisfy interface:
	return nil
}

// func (st *Store) Header() *tabwriter.Writer {
// 	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
// 	// Columns: store name, namespace, backend, artifacts path,
// 	// images path, and how many images are defined
// 	fmt.Fprintln(w, "NAME\tNAMESPACE\tBACKEND\tARTIFACTS_PATH\tIMAGES_PATH\tIMAGE_COUNT")
// 	return w
// }
//
// func (st *Store) PrintInfo(w *tabwriter.Writer) {
// 	imageCount := len(st.Spec.Images)
// 	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
// 		st.Metadata.Name,
// 		st.Metadata.Namespace,
// 		st.Spec.Backend,
// 		st.Spec.Paths.ArtifactsPath,
// 		st.Spec.Paths.ImagesPath,
// 		imageCount,
// 	)
// }
//
// // func (st *Store) GetBaseImagePath(name string) (string, error) {
// // 	return filepath.Join(st.Spec.Paths.ArtifactsPath, st.Spec.Images[name].File)
// // }
//

// NewStoreRecord creates a new store record from the provided store configuration.
func NewStoreRecord(spec Config) *db.Store {
	images := make([]db.Image, len(spec.Images))

	for index, img := range spec.Images {
		images[index] = db.Image{
			Name:      img.Name,
			Display:   img.Display,
			Version:   img.Version,
			OsProfile: img.OSProfile,
			File:      img.File,
			Checksum:  img.Checksum,
			Size:      img.Size,
		}
	}

	return &db.Store{
		Name:          spec.Name,
		Namespace:     spec.Namespace,
		Labels:        spec.Labels,
		Backend:       spec.Backend,
		ArtifactsPath: spec.Paths.Artifacts,
		ImagesPath:    spec.Paths.Images,
		Images:        images,
		CreatedAt:     time.Now(),
	}
}
