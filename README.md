[![codecov](https://codecov.io/gh/alidn/Yalp/branch/master/graph/badge.svg)](https://codecov.io/gh/alidn/Yalp)
![github actions](https://github.com/alidn/Yalp/workflows/Go/badge.svg)
# Yalp

Yalp is a simple load balancer written in Go. It supports Round-Robin and Session Persistence algorithms with session persistence.

# Usage
You can change the configurations in the config.yaml file (see [this](https://github.com/alidn/Yalp/blob/master/config.yaml) for an example)

### Docker
`docker build -t balancer .`

`docker run --name balancer -p:9000:9000 balancer`

Note that you have to expose port 9000

### Using Go
`go build main.go`

`./main`
