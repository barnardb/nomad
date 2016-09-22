package client

import (
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
)

type serverlist struct {
	e  endpoints
	mu sync.RWMutex
}

func newServerList() *serverlist {
	return &serverlist{}
}

// set the server list to a new list. The new list will be shuffled and sorted
// by priority.
func (s *serverlist) set(newe []*endpoint) {
	// Randomize the order
	for i, j := range rand.Perm(len(newe)) {
		newe[i], newe[j] = newe[j], newe[i]
	}

	// Sort by priority
	sort.Sort(endpoints(newe))

	s.mu.Lock()
	defer s.mu.Unlock()
	s.e = newe
	//TODO notify waiters of change?
}

// get a server or nil if no servers available.
func (s *serverlist) get() *endpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.e) > 0 {
		return s.e[0]
	}
	return nil
}

// all returns a copy of the full server list
func (s *serverlist) all() []*endpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*endpoint, len(s.e))
	copy(out, s.e)
	return out
}

// mark a server as failed by rotating it to the end of the list regardless of
// priorities.
func (s *serverlist) mark(e *endpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := 0; i < len(s.e); i++ {
		if s.e[i].equal(e) {
			copy(s.e[i:], s.e[i+1:])
			s.e[len(s.e)-1] = e
			break
		}
	}
}

func (e endpoints) Len() int {
	return len(e)
}

func (e endpoints) Less(i int, j int) bool {
	// Sort only by priority as endpoints should be shuffled and ordered
	// only by priority
	return e[i].priority < e[j].priority
}

func (e endpoints) Swap(i int, j int) {
	e[i], e[j] = e[j], e[i]
}

type endpoints []*endpoint

func (e endpoints) String() string {
	names := make([]string, 0, len(e))
	for _, endpoint := range e {
		names = append(names, endpoint.name)
	}
	return strings.Join(names, ",")
}

type endpoint struct {
	name string
	addr net.Addr

	// 0 being the highest priority
	priority int
}

// equal returns true if the name and addr match between two endpoints.
// Priority is ignored.
func (e *endpoint) equal(o *endpoint) bool {
	return e.name == o.name && e.addr == o.addr
}
