package sqlite

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestMetadataDAOGetUnknown(t *testing.T) {
	testMetadataDAO(t, func(tx Transaction, dao *MetadataDAO) {
		res, err := dao.Get("unknown")
		assert.Nil(t, err)
		assert.Equal(t, res, "")
	})
}

func TestMetadataDAOGetExisting(t *testing.T) {
	testMetadataDAO(t, func(tx Transaction, dao *MetadataDAO) {
		res, err := dao.Get("a_metadata")
		assert.Nil(t, err)
		assert.Equal(t, res, "value")
	})
}

func TestMetadataDAOSetUnknown(t *testing.T) {
	testMetadataDAO(t, func(tx Transaction, dao *MetadataDAO) {
		res, err := dao.Get("new_metadata")
		assert.Nil(t, err)
		assert.Equal(t, res, "")

		err = dao.Set("new_metadata", "pamplemousse")
		assert.Nil(t, err)

		res, err = dao.Get("new_metadata")
		assert.Nil(t, err)
		assert.Equal(t, res, "pamplemousse")
	})
}

func TestMetadataDAOSetExisting(t *testing.T) {
	testMetadataDAO(t, func(tx Transaction, dao *MetadataDAO) {
		err := dao.Set("a_metadata", "new_value")
		assert.Nil(t, err)

		res, err := dao.Get("a_metadata")
		assert.Nil(t, err)
		assert.Equal(t, res, "new_value")
	})
}

func testMetadataDAO(t *testing.T, callback func(tx Transaction, dao *MetadataDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewMetadataDAO(tx))
	})
}
