/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package casbin

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestEnforcerCache(t *testing.T) {

	t.SkipNow()

	lock := make(map[string]*CacheData)
	enforcerCacheExpirationInSec := "432000"
	cacheExpiration, _ := strconv.Atoi(enforcerCacheExpirationInSec)
	cache123 := cache.New(time.Second*time.Duration(cacheExpiration), 5*time.Minute)

	t.Run("requesterAndWriter", func(t *testing.T) {
		for i := 0; i < 100_000; i++ {
			emailId := util.GetRandomStringOfGivenLength(rand.Intn(1000)) + "@yopmail.com"
			getAndSet(lock, emailId, cache123)
			result, expiration, b := cache123.GetWithExpiration(emailId)
			fmt.Println("result", result, "expiration", expiration, "found", b)
		}
	})
	t.Run("CacheInvalidate", func(t *testing.T) {
		invalidateCache_123(lock, cache123)
	})

	t.Run("CacheDump", func(t *testing.T) {
		for i := 0; i < 100_000; i++ {
			emailId := util.GetRandomStringOfGivenLength(rand.Intn(50)) + "@yopmail.com"
			getAndSet(lock, emailId, cache123)
			cache123.GetWithExpiration(emailId)
			//result, expiration, b := cache123.GetWithExpiration(emailId)
			//fmt.Println("result", result, "expiration", expiration, "found", b)
		}
		//invalidateCache_123(lock, cache123)

		fmt.Println("dump: ", GetCacheDump(cache123))
	})

}

func GetCacheDump(cache *cache.Cache) string {
	items := cache.Items()
	cacheData, err := json.Marshal(items)
	if err != nil {
		fmt.Println("error occurred while taking cache dump", "reason", err)
		return ""
	}
	return string(cacheData)
}

func invalidateCache_123(lock map[string]*CacheData, cache *cache.Cache) {
	for emailId := range lock {
		cache.Delete(emailId)
		cacheLock123 := getEnforcerCacheLock_123(lock, emailId)
		cacheLock123.lock.Lock()
		cacheLock123.cacheCleaningFlag = true
		cacheLock123.lock.Unlock()
	}
}

func getAndSet(lock map[string]*CacheData, emailId string, cache *cache.Cache) bool {
	cacheLock := getEnforcerCacheLock_123(lock, emailId)
	cacheLock.lock.RLock()
	atomic.AddInt64(&cacheLock.enforceReqCounter, 1)
	_, found := cache.Get(emailId)
	cacheLock.lock.RUnlock()
	if found {
		// do nothing
		cacheLock.lock.Lock()
		defer cacheLock.lock.Unlock()
		returnVal := atomic.AddInt64(&cacheLock.enforceReqCounter, -1)
		if cacheLock.cacheCleaningFlag {
			if returnVal == 0 {
				cacheLock.cacheCleaningFlag = false
			}
		}
		return true
	}

	resultVal := enforce(emailId)
	cacheLock.lock.Lock()
	if !cacheLock.cacheCleaningFlag {
		cache.Set(emailId, resultVal, 0)
	}
	returnVal := atomic.AddInt64(&cacheLock.enforceReqCounter, -1)
	if cacheLock.cacheCleaningFlag {
		if returnVal == 0 {
			cacheLock.cacheCleaningFlag = false
		}
	}
	cacheLock.lock.Unlock()
	return false
}

func getEnforcerCacheLock_123(lock map[string]*CacheData, emailId string) *CacheData {
	enforcerCacheMutex, found := lock[getLockKey(emailId)]
	if !found {
		enforcerCacheMutex =
			&CacheData{
				lock:              &sync.RWMutex{},
				enforceReqCounter: int64(0),
				cacheCleaningFlag: false,
			}
		lock[getLockKey(emailId)] = enforcerCacheMutex
	}
	return enforcerCacheMutex
}

func enforce(randomeKey string) bool {
	return len(randomeKey)%2 == 0
}
