package store_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/akarasz/pajthy-backend/store"
)

func TestSuiteWithInMemory(t *testing.T) {
	s := store.NewInMemory()
	suite.Run(t, &Suite{Subject: s})
}
