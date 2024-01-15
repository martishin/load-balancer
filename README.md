# Load Balancer 
Simple round robin load balancer implemented with Go

## How to Use
```
Usage of ./load-balancer:
  -backends string
        Load balanced backends, use commas to separate
  -port int
        Port to serve (default 80)
```

## Running Locally
* Run the balancer and 3 example web-servers: `make run`
* The balancer will be listening on port `8080`
