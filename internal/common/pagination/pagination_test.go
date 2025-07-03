package pagination

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationInfo_JSONMarshalling(t *testing.T) {
	p := PaginationInfo{
		Page:       2,
		PageSize:   20,
		Total:      100,
		TotalPages: 5,
	}
	data, err := json.Marshal(p)
	assert.NoError(t, err)
	var unmarshalled PaginationInfo
	err = json.Unmarshal(data, &unmarshalled)
	assert.NoError(t, err)
	assert.Equal(t, p, unmarshalled)
}

func TestPaginationInfo_Fields(t *testing.T) {
	p := PaginationInfo{
		Page:       1,
		PageSize:   10,
		Total:      50,
		TotalPages: 5,
	}
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.PageSize)
	assert.Equal(t, int64(50), p.Total)
	assert.Equal(t, 5, p.TotalPages)
}
