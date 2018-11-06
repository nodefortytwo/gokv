package consul_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/philippgille/gokv/consul"
	"github.com/philippgille/gokv/test"
)

// TestClient tests if reading and writing to the store works properly.
//
// Note: This test is only executed if the initial connection to Consul works.
func TestClient(t *testing.T) {
	if !checkConsulConnection() {
		t.Skip("No connection to Consul could be established. Probably not running in a proper test environment.")
	}

	options := consul.DefaultOptions
	options.Folder = "test_" + strconv.FormatInt(time.Now().Unix(), 10)
	client, err := consul.NewClient(options)
	if err != nil {
		t.Error(err)
	}

	test.TestStore(client, t)
}

// TestClientConcurrent launches a bunch of goroutines that concurrently work with the Consul client.
func TestClientConcurrent(t *testing.T) {
	if !checkConsulConnection() {
		t.Skip("No connection to Consul could be established. Probably not running in a proper test environment.")
	}

	options := consul.DefaultOptions
	options.Folder = "test_" + strconv.FormatInt(time.Now().Unix(), 10)
	client, err := consul.NewClient(options)
	if err != nil {
		t.Error(err)
	}

	goroutineCount := 1000

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(goroutineCount) // Must be called before any goroutine is started
	for i := 0; i < goroutineCount; i++ {
		go test.InteractWithStore(client, strconv.Itoa(i), t, &waitGroup)
	}
	waitGroup.Wait()

	// Now make sure that all values are in the store
	expected := test.Foo{}
	for i := 0; i < goroutineCount; i++ {
		actualPtr := new(test.Foo)
		found, err := client.Get(strconv.Itoa(i), actualPtr)
		if err != nil {
			t.Errorf("An error occurred during the test: %v", err)
		}
		if !found {
			t.Errorf("No value was found, but should have been")
		}
		actual := *actualPtr
		if actual != expected {
			t.Errorf("Expected: %v, but was: %v", expected, actual)
		}
	}
}

// checkConsulConnection returns true if a connection could be made, false otherwise.
func checkConsulConnection() bool {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return false
	}
	res, err := client.Status().Leader()
	if err != nil || res == "" {
		return false
	}
	return true
}
