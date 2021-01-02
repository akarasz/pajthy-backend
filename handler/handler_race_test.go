package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/akarasz/pajthy-backend/event"
	"github.com/akarasz/pajthy-backend/handler"
	"github.com/akarasz/pajthy-backend/store"
)

func TestParallelActions(t *testing.T) {
	s := store.New()
	e := event.New()
	r := handler.NewRouter(s, e)

	wg := &sync.WaitGroup{}

	voters := []string{ "Alice", "Bob", "Carol" }

	for c := 0; c < 3; c++ {
		wg.Add(1)
		go func() {
			res := request(t, r, "POST", "/", `["first", "second"]`)
			sessionUrl := res.HeaderMap["Location"][0]

			for _, v := range voters {
				request(t, r, "PUT", sessionUrl + "/join", fmt.Sprintf(`"%s"`, v))
			}

			sWG := &sync.WaitGroup{}
			for i := 0; i < 4; i++ {
				sWG.Add(1)
				go func(sessionUrl string) {
					request(t, r, "PATCH", sessionUrl + "/start", nil)
					for v := 0; v < 2; v++ {
						for _, v := range voters[1:3] {
							request(t, r, "PUT", sessionUrl, fmt.Sprintf(`{"Choice": "first", "Participant": "%s"}`, v))
							request(t, r, "PUT", sessionUrl, fmt.Sprintf(`{"Choice": "second", "Participant": "%s"}`, v))
						}
					}
					request(t, r, "PATCH", sessionUrl + "/stop", nil)
					request(t, r, "PATCH", sessionUrl + "/reset", nil)

					request(t, r, "PATCH", sessionUrl + "/start", nil)
					for v := 0; v < 2; v++ {
						for _, v := range voters {
							request(t, r, "PUT", sessionUrl, fmt.Sprintf(`{"Choice": "first", "Participant": "%s"}`, v))
							request(t, r, "PUT", sessionUrl, fmt.Sprintf(`{"Choice": "second", "Participant": "%s"}`, v))
						}
					}
					request(t, r, "PATCH", sessionUrl + "/reset", nil)

					sWG.Done()
				}(sessionUrl)
				sWG.Wait()
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

func request(t *testing.T, h *mux.Router, method, url string, body interface{}) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(body.(string))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)

	h.ServeHTTP(res, req)
	return res
}