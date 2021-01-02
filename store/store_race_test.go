package store_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/akarasz/pajthy-backend/domain"
	"github.com/akarasz/pajthy-backend/store"
	"github.com/stretchr/testify/require"
)

func TestCreateThenReadModifyWrite(t *testing.T) {
	s := store.New()
	wg := &sync.WaitGroup{}

	// create some
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id string) {
			s.Create(id, domain.NewSession())
			wg.Done()
		}(idFor(i))
	}
	wg.Wait()

	// read-modify-write all in parallel
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id string) {
				session, err := s.LockAndLoad(id)
				defer s.Unlock(id)
				require.NoError(t, err)

				session.Open = !session.Open

				err = s.Update(id, session)
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