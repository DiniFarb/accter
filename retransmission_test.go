package accter

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddAndRemove(t *testing.T) {
	rt := CreateLocalRetransmissionHandler()
	var wg sync.WaitGroup
	wg.Add(3)
	go func(wg *sync.WaitGroup) {
		for i := 0; i < 10000; i++ {
			rt.AddToCache(fmt.Sprintf("foo%d", i))
			y := rt.IsRetransmission(fmt.Sprintf("foo%d", i))
			if y {
				rt.RemoveFromCache(fmt.Sprintf("foo%d", i))
			} else {
				t.Errorf("retransmission not found (NOT IN SYNC)")
			}
		}
		defer wg.Done()
	}(&wg)
	go func(wg *sync.WaitGroup) {
		for i := 0; i < 10000; i++ {
			rt.AddToCache(fmt.Sprintf("bar%d", i))
			y := rt.IsRetransmission(fmt.Sprintf("bar%d", i))
			if y {
				rt.RemoveFromCache(fmt.Sprintf("bar%d", i))
			} else {
				t.Errorf("retransmission not found (NOT IN SYNC)")
			}
		}
		defer wg.Done()
	}(&wg)
	go func(wg *sync.WaitGroup) {
		for i := 0; i < 10000; i++ {
			rt.AddToCache(fmt.Sprintf("foobar%d", i))
			y := rt.IsRetransmission(fmt.Sprintf("foobar%d", i))
			if y {
				rt.RemoveFromCache(fmt.Sprintf("foobar%d", i))
			} else {
				t.Errorf("retransmission not found (NOT IN SYNC)")
			}
		}
		defer wg.Done()
	}(&wg)
	wg.Wait()
	want := 0
	got := len(rt.(*Retransmissions).retransmissions)
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestCleanCycle(t *testing.T) {
	rt := CreateLocalRetransmissionHandler()
	rt.SetCleanCycleSeconds(1)
	rt.SetObjectLifetimeSeconds(1)
	for i := 0; i < 10000; i++ {
		rt.AddToCache(fmt.Sprintf("foobar%d", i))
	}
	time.Sleep(5 * time.Second)
	want := 0
	got := len(rt.(*Retransmissions).retransmissions)
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
