package event_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/stretchr/testify/require"
)

func TestSubscribeThenEmitSomeThenUnsubscribe(t *testing.T) {
	e := event.New()
	wg := &sync.WaitGroup{}

	for id := 0; id < 3; id++ {
		for ws := 0; ws < 3; ws++ {
			wg.Add(1)
			go func(id string, ws string) {
				c, err := e.Subscribe(id, event.Voter, ws)
				require.NoError(t, err)

				go func(c chan *event.Payload) { for { <-c } }(c)

				eWG := &sync.WaitGroup{}
				for i := 0; i < 3; i++ {
					eWG.Add(1)
					go func() {
						e.Emit(id, event.Voter, event.Enabled, nil)
						eWG.Done()
					}()
				}
				eWG.Wait()

				err = e.Unsubscribe(id, ws)
				require.NoError(t, err)

				wg.Done()
			}(idFor(id), wsFor(ws))
		}
	}

	wg.Wait()
}

func idFor(i int) string {
	return fmt.Sprintf("id%d", i)
}

func wsFor(i int) string {
	return fmt.Sprintf("ws%d", i)
}