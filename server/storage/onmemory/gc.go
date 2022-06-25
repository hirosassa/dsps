package onmemory

import (
	"context"
	"time"

	"github.com/m3dev/dsps/server/sync"
)

var gcInterval = 5 * time.Minute
var gcTimeout = 3 * time.Second

func (s *onmemoryStorage) startGC() {
	s.daemonSystem.Start("gc", func(ctx context.Context) (sync.DaemonNextRun, error) {
		ctx, cancel := context.WithTimeout(ctx, gcTimeout)
		defer cancel()

		err := s.GC(ctx)
		return sync.DaemonNextRun{Interval: gcInterval}, err
	})
}

func (s *onmemoryStorage) GC(ctx context.Context) error {
	unlock, err := s.lock.Lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()

	for _, ch := range s.channels {
		if err := ctx.Err(); err != nil {
			return err // Context canceled
		}
		expireBefore := s.systemClock.Now().Add(-ch.Expire().Duration)

		for sid, sbsc := range ch.subscribers {
			if err := ctx.Err(); err != nil {
				return err // Context canceled
			}

			// Remove expired subscriber.
			if sbsc.lastActivity.Before(expireBefore) {
				delete(ch.subscribers, sid)
				continue
			}

			// Remove expired messages from subscriber queue.
			aliveMsgs := make([]*onmemoryMessage, 0, len(sbsc.messages))
			for _, msg := range sbsc.messages {
				if !msg.ExpireAt.Before(expireBefore) {
					aliveMsgs = append(aliveMsgs, msg)
				}
			}
			sbsc.messages = aliveMsgs
		}

		// Remove expired message log.
		for msgLoc, msg := range ch.log {
			if err := ctx.Err(); err != nil {
				return err // Context canceled
			}

			if msg.ExpireAt.Before(expireBefore) {
				delete(ch.log, msgLoc)
			}
		}
	}

	// Delete expired JWT revocation memory
	for jti, exp := range s.revokedJwts {
		if err := ctx.Err(); err != nil {
			return err // Context canceled
		}

		if time.Time(exp).Before(s.systemClock.Now().Time) {
			delete(s.revokedJwts, jti)
		}
	}
	return nil
}
