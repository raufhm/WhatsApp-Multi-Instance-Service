package broadcast

import (
	"sync"

	"github.com/raufhm/whops/domain"
)

// BroadcastBuffer is the default per-subscriber channel capacity. When a
// subscriber falls behind, its channel is closed and the client must reconnect
// (and backfill via the since cursor).
const BroadcastBuffer = 256

// Broadcaster fans out instance log events to per-host subscriber channels.
// It is the push side of the dashboard Monitoring live tail.
type Broadcaster struct {
	mu   sync.Mutex
	subs map[string]map[chan domain.InstanceLogEvent]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subs: make(map[string]map[chan domain.InstanceLogEvent]struct{})}
}

// Subscribe registers a channel for hostID and returns it with an unsubscribe
// func. The caller must select on the channel and call the unsubscribe func
// when the client disconnects. A closed channel signals that the subscriber
// fell behind (buffer overflow) and should reconnect.
func (b *Broadcaster) Subscribe(hostID string, buffer int) (<-chan domain.InstanceLogEvent, func()) {
	if buffer < 1 {
		buffer = BroadcastBuffer
	}
	ch := make(chan domain.InstanceLogEvent, buffer)

	b.mu.Lock()
	if b.subs[hostID] == nil {
		b.subs[hostID] = make(map[chan domain.InstanceLogEvent]struct{})
	}
	b.subs[hostID][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			if set, ok := b.subs[hostID]; ok {
				delete(set, ch)
				if len(set) == 0 {
					delete(b.subs, hostID)
				}
			}
		})
	}
	return ch, unsubscribe
}

// Publish delivers ev to every subscriber of hostID. A full channel (slow
// consumer) is closed and removed so the client reconnects and backfills.
func (b *Broadcaster) Publish(hostID string, ev domain.InstanceLogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs[hostID] {
		select {
		case ch <- ev:
		default:
			close(ch)
			delete(b.subs[hostID], ch)
			if len(b.subs[hostID]) == 0 {
				delete(b.subs, hostID)
			}
		}
	}
}
