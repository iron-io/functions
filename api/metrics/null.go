// Use this to ignore metrics.
package metrics

type NullMetrics struct {
}

func (m *NullMetrics) Write(name, mtype string, value float64) {

}
