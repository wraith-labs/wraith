package libwraith_test

import (
	"sync"
	"testing"
	"time"

	"dev.l1qu1d.net/wraith-labs/wraith/wraith/libwraith"
)

func TestShmInit(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()
}

func TestShmWriteReadReinit(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()

	testshm.Set("foo", "bar")

	testshm.Init()

	if value := testshm.Get("foo"); value != nil {
		t.Errorf("shm reinit failed (re-initialised shm still holds old values)")
	}
}

func TestShmWriteRead(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()

	dataset := map[string]interface{}{
		"hello":  1,
		"world":  "foo",
		"bar":    0.125,
		"struct": struct{}{},
		"bool":   true,
	}

	for key, value := range dataset {
		testshm.Set(key, value)
	}

	for key, value := range dataset {
		if testshm.Get(key) != value {
			t.Errorf("shm readback failed (wrote `%v` to `%s` but read `%v` back)", dataset[key], key, value)
		}
	}
}

func TestShmWriteReadWatchUnwatchAsync(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()

	dataset := map[string]interface{}{
		"hello":  1,
		"world":  "foo",
		"bar":    0.125,
		"struct": struct{}{},
		"bool":   true,
	}

	wg := sync.WaitGroup{}
	wg.Add(len(dataset))

	for key, value := range dataset {
		go func(key string, value interface{}) {
			watcher, watcherId := testshm.Watch(key)
			testshm.Set(key, value)
			select {
			case data := <-watcher:
				if data != value {
					t.Errorf("shm watch failed (wrote `%v` to `%s` but watcher returned `%v` back)", dataset[key], key, data)
				}
			case <-time.After(500 * time.Millisecond):
				t.Errorf("shm watch failed (timed out waiting for value `%v` in cell `%s`)", value, key)
			}
			testshm.Unwatch(key, watcherId)
			testshm.Set(key, value)
			select {
			case <-watcher:
				t.Errorf("shm unwatch failed (unwatched cell `%s` still sent an update to watch channel)", key)
			case <-time.After(500 * time.Millisecond):
			}
			wg.Done()
		}(key, value)
	}

	wg.Wait()
}

func TestShmDump(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()

	dataset := map[string]interface{}{
		"hello":  1,
		"world":  "foo",
		"bar":    0.125,
		"struct": struct{}{},
		"bool":   true,
	}

	for key, value := range dataset {
		testshm.Set(key, value)
	}

	for key, value := range testshm.Dump() {
		if origvalue, ok := dataset[key]; !ok {
			t.Errorf("shm dump failed (dump missing key `%s`)", key)
		} else if origvalue != value {
			t.Errorf("shm dump failed (dump key `%s` has incorrect value `%v` while `%v` expected)", key, value, origvalue)
		}
	}
}

func TestShmPrune(t *testing.T) {
	testshm := libwraith.Shm{}

	testshm.Init()

	dataset := map[string]interface{}{
		"hello":  1,
		"world":  "foo",
		"bar":    0.125,
		"struct": struct{}{},
		"bool":   true,
		"nil":    nil,
		"nil2":   nil,
	}

	for key, value := range dataset {
		testshm.Set(key, value)
	}

	testshm.Prune()

	shmdump := testshm.Dump()

	for key := range shmdump {
		if key == "nil" || key == "nil2" {
			t.Errorf("shm dump failed (nil-valued cell was not removed)")
		}
	}

	if len(shmdump) != len(dataset)-2 {
		t.Errorf("shm dump failed (incorrect number of cells pruned)")
	}
}
