package accter

import (
	"sync"
	"time"
)

type RetransmissionHandler interface {
	IsRetransmission(key string) bool
	AddToCache(key string) error
	RemoveFromCache(key string)
	CleanCycle() error
	SetCleanCycleSeconds(sec int)
	SetObjectLifetimeSeconds(sec int)
}

type Retransmissions struct {
	sync.RWMutex
	retransmissions      map[string]time.Time
	CleanCycleSeconds    int
	EntryLifetimeSeconds int
}

func CreateLocalRetransmissionHandler() RetransmissionHandler {
	var r = &Retransmissions{
		retransmissions:      make(map[string]time.Time),
		EntryLifetimeSeconds: 300,
		CleanCycleSeconds:    10,
	}
	r.CleanCycle()
	return r
}

func (r *Retransmissions) SetCleanCycleSeconds(sec int) {
	r.Lock()
	defer r.Unlock()
	r.CleanCycleSeconds = sec
}

func (r *Retransmissions) SetObjectLifetimeSeconds(sec int) {
	r.Lock()
	defer r.Unlock()
	r.EntryLifetimeSeconds = sec
}

func (r *Retransmissions) IsRetransmission(key string) bool {
	r.RLock()
	defer r.RUnlock()
	if _, ok := r.retransmissions[key]; ok {
		return true
	}
	return false
}

func (r *Retransmissions) AddToCache(key string) error {
	r.Lock()
	defer r.Unlock()
	r.retransmissions[key] = time.Now()
	return nil
}

func (r *Retransmissions) RemoveFromCache(key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.retransmissions, key)
}

func (r *Retransmissions) CleanCycle() error {
	go func() {
		for {
			time.Sleep(time.Duration(r.CleanCycleSeconds) * time.Second)
			r.RLock()
			rts := make(map[string]time.Time)
			for k, v := range r.retransmissions {
				rts[k] = v
			}
			r.RUnlock()
			for k := range rts {
				if time.Since(rts[k]) > time.Duration(r.EntryLifetimeSeconds)*time.Second {
					r.RemoveFromCache(k)
				}
			}
		}
	}()
	return nil
}
