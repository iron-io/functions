package mqs

import (
	"bytes"
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/treeder/functions/api/models"
)

const (
	testReserveTimeout = 2 * time.Second
)

// A test fixture for generating test message queues.
type fixture struct {
	// Returns a new message queue.
	newMQFunc func(*testing.T) models.MessageQueue
	// Releases any background resources (e.g. docker containers) started by the creation of this fixture.
	shutdown func()
}

var tests = map[string]func(*testing.T){

	"memory": func(t *testing.T) {
		var mq models.MessageQueue
		f := &fixture{
			newMQFunc: func(*testing.T) models.MessageQueue {
				if mq != nil {
					mq.Close()
				}
				mq := NewMemoryMQ(testReserveTimeout)
				return mq
			},
			shutdown: func() {
				if mq != nil {
					mq.Close()
				}
			},
		}
		f.run(t)
	},
}

func TestMQ(t *testing.T) {
	captureLogs()
	for name, test := range tests {
		t.Run(name, test)
	}
}

func (fixture fixture) newMQ(t *testing.T) models.MessageQueue {
	resetLogBuf()
	return fixture.newMQFunc(t)
}

// Runs each sub-test against a new message queue.
func (fixture fixture) run(t *testing.T) {
	defer fixture.shutdown()
	ctx := context.Background()

	t.Run("push-invalid", func(t *testing.T) {
		mq := fixture.newMQ(t)

		// Push nil Task
		if task, err := mq.Push(ctx, nil); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQEmptyTaskID)
		} else if err != models.ErrMQMissingTask {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQMissingTask, err)
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		// Push empty Task
		if task, err := mq.Push(ctx, &models.Task{}); err == nil {
			t.Error("expected error but got none")
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		task := new(models.Task)
		task.ID = "testID"
		task.Priority = new(int32)

		// Push missing ID
		missingID := *task
		missingID.ID = ""
		if task, err := mq.Push(ctx, &missingID); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQEmptyTaskID)
		} else if err != models.ErrMQEmptyTaskID {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQEmptyTaskID, err)
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		// Push missing Priority
		missingPriority := *task
		missingPriority.Priority = nil
		if task, err := mq.Push(ctx, &missingPriority); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQMissingTaskPriority)
		} else if err != models.ErrMQMissingTaskPriority {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQMissingTaskPriority, err)
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		// Push invalid (negative) Priority
		negativePriority := *task
		negativePriority.Priority = new(int32)
		*negativePriority.Priority = -1
		if task, err := mq.Push(ctx, &negativePriority); err == nil {
			t.Error("expected invalid priority error but got none")
		} else if !models.IsErrMQInvalidTaskPriority(err) {
			t.Errorf("expected invalid priority error but got `%v`", err)
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		// Push invalid (>2) Priority
		invalidPriority := *task
		invalidPriority.Priority = new(int32)
		*invalidPriority.Priority = 5
		if task, err := mq.Push(ctx, &invalidPriority); err == nil {
			t.Error("expected invalid priority error but got none")
		} else if !models.IsErrMQInvalidTaskPriority(err) {
			t.Errorf("expected invalid priority error but got `%v`", err)
		} else if task != nil {
			t.Errorf("expected nil task but got: %v", task)
		}

		if t.Failed() {
			t.Log(logBuf.String())
		}
	})

	t.Run("delete-invalid", func(t *testing.T) {
		mq := fixture.newMQ(t)

		// Delete nil Task
		if err := mq.Delete(ctx, nil); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQMissingTask)
		} else if err != models.ErrMQMissingTask {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQMissingTask, err)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		// Delete empty task
		if err := mq.Delete(ctx, &models.Task{}); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQEmptyTaskID)
		} else if err != models.ErrMQEmptyTaskID {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQEmptyTaskID, err)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}

		resetLogBuf()

		//Delete non-existent
		notReal := &models.Task{}
		notReal.ID = "notReal"
		notReal.Priority = new(int32)
		if err := mq.Delete(ctx, notReal); err == nil {
			t.Errorf("expected error `%v` but got none", models.ErrMQTaskNotReserved)
		} else if err != models.ErrMQTaskNotReserved {
			t.Errorf("expected error `%v` but got `%v`", models.ErrMQTaskNotReserved, err)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}
	})

	t.Run("reserve-invalid", func(t *testing.T) {
		mq := fixture.newMQ(t)

		// Reserve when empty
		if task, err := mq.Reserve(ctx); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error:", err)
		} else if task != nil {
			t.Log(logBuf.String())
			t.Fatalf("expected nil task but got: %v", task)
		}
	})

	t.Run("single-timeout", func(t *testing.T) {
		mq := fixture.newMQ(t)

		task := &models.Task{}
		task.ID = "test"
		task.Priority = new(int32)

		if got, err := mq.Push(ctx, task); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error inserting task:", err)
		} else if !reflect.DeepEqual(*task, *got) {
			t.Log(logBuf.String())
			t.Fatalf("pushed %v but returned %v", task, got)
		}

		if got, err := mq.Reserve(ctx); err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error on reserve: `%v`", err)
		} else if !reflect.DeepEqual(*task, *got) {
			t.Log(logBuf.String())
			t.Fatalf("pushed %v but reserved %v", task, got)
		}

		// Let reservation timeout
		<-time.After(20 * testReserveTimeout)

		if got, err := mq.Reserve(ctx); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error on reserve:", err)
		} else if got == nil {
			t.Log(logBuf.String())
			t.Fatal("expected to reserve non-nil")
		} else if !reflect.DeepEqual(*task, *got) {
			t.Log(logBuf.String())
			t.Fatalf("pushed %v but reserved %v", task, got)
		}

		// Let reservation timeout
		<-time.After(20 * testReserveTimeout)
		// Delete too late
		if err := mq.Delete(ctx, task); err != models.ErrMQTaskNotReserved {
			t.Log(logBuf.String())
			t.Fatalf("expected error `%v` but got `%v`", models.ErrMQTaskNotReserved, err)
		}

		if got, err := mq.Reserve(ctx); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error on reserve:", err)
		} else if got == nil {
			t.Log(logBuf.String())
			t.Fatal("expected to reserve non-nil")
		} else if !reflect.DeepEqual(*task, *got) {
			t.Log(logBuf.String())
			t.Fatalf("pushed %v but reserved %v", task, got)
		}
		if err := mq.Delete(ctx, task); err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error `%v`", err)
		}

		if got, err := mq.Reserve(ctx); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error on reserve:", err)
		} else if got != nil {
			t.Log(logBuf.String())
			t.Fatalf("expected nil but reserved %v", got)
		}
	})

	t.Run("fifo", func(t *testing.T) {
		mq := fixture.newMQ(t)

		a := &models.Task{}
		a.ID = "A"
		a.Priority = new(int32)

		b := &models.Task{}
		b.ID = "B"
		b.Priority = new(int32)

		if _, err := mq.Push(ctx, a); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error pushing a:", err)
		}
		if _, err := mq.Push(ctx, b); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error pushing a:", err)
		}

		first, err := mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error on first reserve:", err)
		} else if first == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil first reserve")
		}
		second, err := mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error on second reserve:", err)
		} else if second == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil second reserve")
		}

		if first.ID != a.ID {
			t.Errorf("expected 'A' first, but got %q", first.ID)
		}
		if second.ID != b.ID {
			t.Errorf("expected 'B' second, but got %q", second.ID)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}
	})

	t.Run("delay", func(t *testing.T) {
		mq := fixture.newMQ(t)

		a := &models.Task{}
		a.ID = "A"
		a.Priority = new(int32)
		a.Delay = 2

		b := &models.Task{}
		b.ID = "B"
		b.Priority = new(int32)

		if _, err := mq.Push(ctx, a); err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error pushing a: `%v`", err)
		}
		if _, err := mq.Push(ctx, b); err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error pushing a: `%v`", err)
		}

		first, err := mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error on first reserve: `%v`", err)
		} else if first == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil first reserve")
		}
		if first.ID != b.ID {
			t.Errorf("expected 'b' first, but got %q", first.ID)
		}
		if err := mq.Delete(ctx, first); err != nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected error deleteing task")
		}

		// Wait for delay
		<-time.After(2 * time.Duration(a.Delay) * time.Second)

		second, err := mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error on second reserve: `%v`", err)
		} else if second == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil second reserve")
		}

		if second.ID != a.ID {
			t.Errorf("expected 'a' second, but got %q", second.ID)
		}
		if t.Failed() {
			t.Log(logBuf.String())
		}
	})
}

var logBuf bytes.Buffer

func captureLogs() {
	logrus.SetOutput(&logBuf)
	log.SetOutput(&logBuf)
}

func resetLogBuf() {
	logBuf.Reset()
	logBuf.WriteByte('\n')
}
