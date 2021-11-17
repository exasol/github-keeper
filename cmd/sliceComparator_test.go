package cmd

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type SliceComparatorSuite struct {
	suite.Suite
}

func TestSliceComparatorSuite(t *testing.T) {
	suite.Run(t, new(SliceComparatorSuite))
}

func (suite SliceComparatorSuite) TestEqual() {
	suite.Assert().True(stringSlicesEqualIgnoringOrder([]string{"a", "b"}, []string{"a", "b"}))
}

func (suite SliceComparatorSuite) TestEqualWithDifferentOrder() {
	suite.Assert().True(stringSlicesEqualIgnoringOrder([]string{"a", "b"}, []string{"b", "a"}))
}

func (suite SliceComparatorSuite) TestDifferent() {
	suite.Assert().False(stringSlicesEqualIgnoringOrder([]string{"a", "b"}, []string{"a", "c"}))
}

// The implementation could modify (e.g. sort) the parameters. That would be bad since, it's an unexpected side effect.
func (suite SliceComparatorSuite) TestParametersAreUnchanged() {
	parameter := []string{"a", "c", "b"}
	stringSlicesEqualIgnoringOrder(parameter, []string{"a", "c"})
	suite.Equal(parameter, []string{"a", "c", "b"})
}
