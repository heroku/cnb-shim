// +heroku goVersion go1.14

module github.com/heroku/cnb-shim

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/NYTimes/gziphandler v1.1.1
	github.com/apex/log v1.9.0
	github.com/buildpack/libbuildpack v1.11.0
	github.com/google/uuid v1.1.4
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/heroku/rollrus v0.2.0
	github.com/joeshaw/envdecode v0.0.0-20200121155833-099f1fc765bd
	github.com/joho/godotenv v1.3.0
	github.com/mattn/go-shellwords v1.0.10
	github.com/rollbar/rollbar-go v1.2.0
	github.com/sirupsen/logrus v1.4.2
	gopkg.in/yaml.v2 v2.2.2
)
