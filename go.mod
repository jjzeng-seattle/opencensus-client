module opencensus_exporter/main

go 1.14

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0
	github.com/rogpeppe/gohack v1.0.2 // indirect
	go.opencensus.io v0.22.3
)

replace contrib.go.opencensus.io/exporter/ocagent => ./ocagent
