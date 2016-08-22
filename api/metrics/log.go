// Use this to log metrics to the logs, structured log format. See https://github.com/iron-io/functions/issues/61 for format information.
package metrics

import "github.com/Sirupsen/logrus"

type LogMetrics struct {
}

func (m *LogMetrics) Write(name, mtype string, value float64) {
	logrus.WithFields(logrus.Fields{
		"metric": name,
		"type":   mtype,
		"value":  value,
	}).Infoln("metric logged")
}
