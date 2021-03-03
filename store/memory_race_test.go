package store_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
	"github.com/stretchr/testify/require"
)

func TestParallelCreatesLoadsAndUpdates(t *testing.T) {
	s := store.NewInMemory()
	wg := &sync.WaitGroup{}

	// create some
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id string) {
			s.Save(id, domain.NewSession())
			wg.Done()
		}(idFor(i))
	}
	wg.Wait()

	// read-modify-write all in parallel
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id string) {
				loaded, err := s.Load(id)
				require.NoError(t, err)

				data := *loaded.Data
				data.Open = !data.Open

				err = s.Save(id, &data, loaded.Version)
				require.NoError(t, err)
				wg.Done()
			}(idFor(j))
		}
	}
	wg.Wait()
}

func idFor(i int) string {
	return fmt.Sprintf("test%d", i)
}
