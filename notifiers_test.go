package chronon

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NotifiersSuite struct {
	ChrononSuite
}

func TestNotifiers(t *testing.T) {
	suite.Run(t, new(NotifiersSuite))
}
