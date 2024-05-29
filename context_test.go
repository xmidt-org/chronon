package chronon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextSuite struct {
	ChrononSuite
}

func (suite *ContextSuite) TestGet() {
	suite.Run("Default", func() {
		clock := Get(context.Background())
		suite.Require().NotNil(clock)
		suite.IsType(systemClock{}, clock)
	})

	suite.Run("ClockPresent", func() {
		fc := suite.newFakeClock()
		ctx := context.WithValue(context.Background(), contextKey{}, fc)
		suite.Same(fc, Get(ctx))
	})
}

func (suite *ContextSuite) TestWith() {
	suite.Run("WithNil", func() {
		ctx := With(context.Background(), nil)
		suite.Equal(context.Background(), ctx)
	})

	suite.Run("WithClock", func() {
		fc := suite.newFakeClock()
		ctx := With(context.Background(), fc)
		suite.Same(fc, ctx.Value(contextKey{}))
	})
}

func TestContext(t *testing.T) {
	suite.Run(t, new(ContextSuite))
}
