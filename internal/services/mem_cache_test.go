package services_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stjudewashere/seonaut/internal/services"
)

// TestMEmCache tests by setting a value with Set and then retreiving it with Get making sure it contains
// the same value. Then Test that deleting the item actaully removes it from the cache.
func TestMemCache(t *testing.T) {
	cacheKey := "test"
	type data struct {
		S string
	}

	dataSet := &data{S: "Test"}
	memCache := services.NewMemCache()
	err := memCache.Set(cacheKey, dataSet)
	if err != nil {
		t.Errorf("Error setting cache: %v", err)
	}

	dataGet := &data{}
	err = memCache.Get(cacheKey, dataGet)
	if err != nil {
		t.Errorf("Error getting cache: %v", err)
	}

	if dataSet.S != dataGet.S {
		t.Errorf("%s and %s dataGet are not equal", dataSet.S, dataGet.S)
	}

	err = memCache.Delete(cacheKey)
	if err != nil {
		t.Errorf("Error deleting cache: %v", err)
	}

	err = memCache.Get(cacheKey, dataGet)
	if err == nil {
		t.Errorf("Error getting cache after delete")
	}
}

// Test Set and Get cache with a Slice.
func TestMemCacheSlice(t *testing.T) {
	cacheKey := "test"
	dataSet := []int{1, 2, 3}
	getDataSet := []int{}

	memCache := services.NewMemCache()
	err := memCache.Set(cacheKey, &dataSet)
	if err != nil {
		t.Errorf("Error setting cache: %v", err)
	}

	err = memCache.Get(cacheKey, &getDataSet)
	if err != nil {
		t.Errorf("Error getting cache: %v", err)
	}

	for i, v := range dataSet {
		if getDataSet[i] != v {
			t.Errorf("Error getting cache data %v is not equal to %v", getDataSet[i], v)
		}
	}
}

// Test Set and Get values concurrently.
func TestGetSetConcurrent(t *testing.T) {
	c := services.NewMemCache()

	numElements := 100

	var wg sync.WaitGroup
	wg.Add(numElements)

	for i := 0; i < numElements; i++ {
		go func(id int) {
			defer wg.Done()

			key := "key" + strconv.Itoa(id)
			value := strconv.Itoa(rand.Intn(100))
			err := c.Set(key, &value)
			if err != nil {
				t.Errorf("Error setting value in cache: %v", err)
				return
			}

			var retrievedValue string
			err = c.Get(key, &retrievedValue)
			if err != nil {
				t.Errorf("Error getting value from cache: %v", err)
				return
			}

			if retrievedValue != value {
				t.Errorf("Mismatched values for key %s: expected '%s', got '%s'", key, value, retrievedValue)
			}

			err = c.Delete(key)
			if err != nil {
				t.Errorf("Error deleting value from cache: %v", err)
				return
			}

			var deletedValue string
			err = c.Get(key, &deletedValue)
			if err == nil {
				t.Errorf("Error getting deleted from cache.")
				return
			}
		}(i)
	}

	wg.Wait()
}
