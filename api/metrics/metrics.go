package metrics

type Metrics interface {
	Write(name, mtype string, value float64)
}
