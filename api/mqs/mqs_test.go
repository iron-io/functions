package mqs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	ironmq "github.com/iron-io/iron_go3/mq"

	"github.com/iron-io/functions/api/models"
)

const (
	testReserveTimeout = 2 * time.Second

	tmpBolt  = "/tmp/func_test_bolt.db"
	tmpRedis = "redis://127.0.0.1:6301/"
	tmpIron  = "ironmq+http://test:test@127.0.0.1:8080/test"
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

	"bolt": func(t *testing.T) {
		url, err := url.Parse("bolt://" + tmpBolt)
		if err != nil {
			t.Fatalf("failed to parse bolt url: `%v`", err)
		}
		var mq models.MessageQueue
		f := &fixture{
			newMQFunc: func(t *testing.T) models.MessageQueue {
				if mq != nil {
					mq.Close()
				}
				os.Remove(tmpBolt)
				var err error
				mq, err = NewBoltMQ(url, testReserveTimeout)
				if err != nil {
					t.Fatalf("failed to create bolt mq: `%v`", err)
				}
				return mq
			},
			shutdown: func() {
				if mq != nil {
					mq.Close()
				}
				os.Remove(tmpBolt)
			},
		}
		f.run(t)
	},

	"redis": func(t *testing.T) {
		url, err := url.Parse(tmpRedis)
		if err != nil {
			t.Fatalf("failed to parse redis url: `%v`", err)
		}
		startRedis(t.Logf, t.Fatalf)
		mq, err := NewRedisMQ(url, testReserveTimeout)
		if err != nil {
			t.Fatalf("failed to create redis mq: `%v`", err)
		}
		f := &fixture{
			newMQFunc: func(t *testing.T) models.MessageQueue {
				if err := mq.clear(); err != nil {
					t.Fatalf("failed to clear queue: %s", err)
				}
				return mq
			},
			shutdown: func() {
				if mq != nil {
					mq.Close()
				}
				tryRun(t.Logf, "stop redis container", exec.Command("docker", "rm", "-f", "iron-redis-test"))
			},
		}
		f.run(t)
	},

	"ironmq": func(t *testing.T) {
		tryRun(t.Logf, "remove old ironmq container", exec.Command("docker", "rm", "-f", "iron-mq-test"))
		mustRun(t.Fatalf, "start ironmq container", exec.Command("docker", "run", "--name", "iron-mq-test", "-p", "8080:8080", "-d", "iron/mq:lite"))

		url, err := url.Parse(tmpIron)
		if err != nil {
			t.Fatalf("failed to parse ironmq url: `%v`", err)
		}

		var mq *IronMQ

		timeout := time.After(20 * time.Second)
		for {
			mq, err = NewIronMQ(url, testReserveTimeout)
			if err == nil {
				break
			}
			fmt.Println("failed to connect to ironmq:", err)
			select {
			case <-timeout:
				t.Fatal("timed out waiting for ironmq")
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}

		f := &fixture{
			newMQFunc: func(t *testing.T) models.MessageQueue {
				if err := mq.clear(); err != nil {
					t.Fatalf("failed to clear queues: %s", err)
				}
				return mq
			},
			shutdown: func() {
				if mq != nil {
					mq.Close()
				}
				tryRun(t.Logf, "stop ironmq container", exec.Command("docker", "rm", "-f", "iron-mq-test"))
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

	// TODO fix fifo-timeout for iron, redis and memory
	t.Run("fifo-timeout", func(t *testing.T) {
		mq := fixture.newMQ(t)

		a := &models.Task{}
		a.ID = "A"
		a.Priority = new(int32)

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
		// Let the reservation timeout
		<-time.After(20 * testReserveTimeout)

		first, err = mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error on first reserve: `%v`", err)
		} else if first == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil first reserve")
		}
		if first.ID != a.ID {
			t.Errorf("expected 'A' first, but got %q", first.ID)
		}
		second, err := mq.Reserve(ctx)
		if err != nil {
			t.Log(logBuf.String())
			t.Fatalf("unexpected error on second reserve: `%v`", err)
		} else if second == nil {
			t.Log(logBuf.String())
			t.Fatal("unexpected nil second reserve")
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

func startRedis(logf, fatalf func(string, ...interface{})) {
	tryRun(logf, "remove old redis container", exec.Command("docker", "rm", "-f", "iron-redis-test"))
	mustRun(fatalf, "start redis container", exec.Command("docker", "run", "--name", "iron-redis-test", "-p", "6301:6379", "-d", "redis"))
	timeout := time.After(20 * time.Second)

	for {
		c, err := redis.DialURL(tmpRedis)
		if err == nil {
			_, err = c.Do("PING")
			c.Close()
			if err == nil {
				break
			}
		}
		fmt.Println("failed to PING redis:", err)
		select {
		case <-timeout:
			fatalf("timed out waiting for redis")
		case <-time.After(500 * time.Millisecond):
			continue
		}
	}
}

func tryRun(logf func(string, ...interface{}), desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		logf("failed to %s: %s", desc, b.String())
	}
}

func mustRun(fatalf func(string, ...interface{}), desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		fatalf("failed to %s: %s", desc, b.String())
	}
}

func (mq *RedisMQ) clear() error {
	conn := mq.pool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	return err
}

func (mq *IronMQ) clear() error {
	for i := 0; i < 3; i++ {
		err := mq.queues[i].Clear()
		if err != nil && !ironmq.ErrQueueNotFound(err) {
			return err
		}
	}
	return nil
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