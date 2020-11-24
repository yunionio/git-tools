package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yunionio/git-tools/pkg/types"
)

func TestGetSemverBranchVersion(t *testing.T) {
	assert := assert.New(t)

	ver, err := GetSemverBranchVersion("release/3.4")
	assert.Equal("3.4", ver)
	assert.Nil(err)

	ver, err = GetSemverBranchVersion("release-3.4")
	assert.Equal("3.4", ver)
	assert.Nil(err)

	ver, err = GetSemverBranchVersion("release/3.4.0")
	assert.Equal("", ver)
	assert.NotNil(err)

	ver, err = GetSemverBranchVersion("feature/add-x")
	assert.Equal("", ver)
	assert.NotNil(err)
}

func TestFilterTagsByPrefix(t *testing.T) {
	tags := []*types.Tag{
		{
			Name: "v3.4.3",
		},
		{
			Name: "v3.4.1",
		},
		{
			Name: "v3.4.0",
		},
	}

	assert := assert.New(t)
	assert.Equal(tags, filterTagsByPrefix("3.4", tags))

	assert.Equal(make([]*types.Tag, 0), filterTagsByPrefix("3.3", tags))
}
