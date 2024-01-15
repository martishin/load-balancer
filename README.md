# Load Balancer 
Simple round-robin load balancer with health checks implemented with Go, [demo](https://load-balancer.martishin.com/)

## How to Use
```
Usage of ./load-balancer:
  -backends string
        Load balanced backends, use commas to separate
  -port int
        Port to serve (default 80)
```

## Running Locally
* Build the load balancer: `make build`
* Run the balancer and 3 example web-servers: `make run`
* The balancer will be listening on port `8080`

## Testing
* Run tests: `make test`
