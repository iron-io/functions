package common

import (
	"context"

	"github.com/Sirupsen/logrus"
	rcommon "github.com/iron-io/runner/common"
)

// todo: the rest of these are in runner repo, should consolidate. Baybe into a separate "utils" repo or something?

// LoggerWithStack is a helper to add a "stack" trace to the logs. call is typically the current method.
func LoggerWithStack(ctx context.Context, call string) (context.Context, logrus.FieldLogger) {
	l := rcommon.Logger(ctx)
	entry, ok := l.(*logrus.Entry)
	if !ok {
		l.Errorln("The type of the logger wasn't a logrus.Entry, maybe it was the base Logger?")
		return ctx, l
	}
	// grab the stack field and append to it
	v, ok := entry.Data["stack"]
	stack := ""
	if ok && v != nil {
		stack = v.(string)
	}
	stack += "." + call
	l = l.WithField("stack", v)
	ctx = rcommon.WithLogger(ctx, l)
	return ctx, l
}
