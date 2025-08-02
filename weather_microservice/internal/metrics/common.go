package metrics

func RegisterAll(metrics ...interface{ Register() }) {
	for _, m := range metrics {
		m.Register()
	}
}
