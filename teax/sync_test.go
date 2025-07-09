package teax_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/donderom/sqwat/teax"
)

var err = errors.New("boom")

type saveTracker struct {
	saved  int
	failed bool
}

func (s *saveTracker) save() error {
	s.saved++
	if s.failed {
		return err
	}

	return nil
}

func TestSaver(t *testing.T) {
	t.Parallel()

	tracker := &saveTracker{}
	saver := teax.NewSaver(tracker.save)
	assert.NoError(t, saver.Save())
	assert.Equal(t, 1, tracker.saved)
}

func TestFailedSaver(t *testing.T) {
	t.Parallel()

	tracker := &saveTracker{failed: true}
	saver := teax.NewSaver(tracker.save)
	assert.Equal(t, err, saver.Save())
	assert.Equal(t, 1, tracker.saved)
}
