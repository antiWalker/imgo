module httpbenchmark

go 1.12

replace imgo/benckmark/http/requester v0.0.0 => ./requester

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/spf13/viper v1.3.2
	imgo/benckmark/http/requester v0.0.0
)
