# GraftNode

**GraftNode** is a core component of the [Plantarium Platform](https://github.com/plantarium-platform), a lightweight and resource-efficient solution inspired by cloud architecture principles, designed for running serverless functions and microservices.

It is a Go-based service that dynamically starts backend service instances only when needed, optimizing resource usage. It integrates with HAProxy for request routing and works with lightweight processes to minimize resource consumption.

## How It Works

### Initial Setup (Cold Start)
- HAProxy is initially configured with a mock backend that doesn’t route to the actual service.
- When the first request to `/hello` is received:
    - GraftNode starts the Java service instance defined in `config-example.yaml`.
    - HAProxy's configuration is updated to route requests to the newly started service.
- The first request might experience a delay due to the service startup time.

![Test Script Logs example](Screenshot%20from%202024-10-30%2013-23-03.png)

![GraftNode Logs example](Screenshot%20from%202024-10-30%2013-29-51.png)

### After Warm-Up
- Once the service instance is running, subsequent requests are handled immediately, ensuring optimal performance.
- HAProxy now routes directly to the active service.

## Project Structure

Below is the simplified file structure of the GraftNode project:

```plaintext
+---examples
|   \---java-service-example    # Example Java service used for testing and demonstrations
|
+---graftnode-go
|   |   go.mod
|   |   go.sum
|   |   main.go
|   |
|   \---services
|       |   config.go           # Handles service configuration parsing and management
|       |   executor.go         # Manages execution of service instances
|       |   haproxy.go          # Interfaces with HAProxy for dynamic backend updates
|       |
|       \---haproxy
|               haproxy_client.go       # HAProxy API client implementation
|               haproxy_config_manager.go # Manages HAProxy configuration through API
|               middleware.go           # Middleware for handling API requests
|
\---resources
    |   config-example.yaml     # Example configuration file for GraftNode
    |   java-service-example.jar # Example Java service binary
    |   test.sh                 # Script for testing functionality
    |
    \---haproxy
            haproxy-run.sh      # Script to run HAProxy in Docker
            haproxy.cfg         # HAProxy configuration file
```

## How to Run

1. **Step 1: Run HAProxy**
  - Use the provided script to run HAProxy in a Docker container. From the root directory, execute:
    ```bash
    ./resources/haproxy/haproxy-run.sh
    ```
  - This command will start HAProxy with the necessary configuration for request routing.

2. **Step 2: Start GraftNode**
  - Start the GraftNode server using the example configuration file:
    ```bash
    go run graftnode-go/main.go --config resources/config-example.yaml
    ```
  - The server will listen on port 8000 and bind the `/hello` backend to HAProxy.
3. **Test the Setup**:
    - Use the `resources/test.sh` script to simulate requests and initiate the activities described:
        - The script will send a `curl` request to `http://localhost/hello`, triggering GraftNode to start the service instance.
        - Once the service is running, HAProxy will replace the initial configuration and forward future requests directly to the newly launched service.
    - Run the script:
      ```bash
      ./resources/test.sh
      ```

## Contact

If you have questions or want to contribute, feel free to reach out on [GitHub](https://github.com/glorko) or [Telegram](https://t.me/glorfindeil).
