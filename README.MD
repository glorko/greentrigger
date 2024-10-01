
# GreenTrigger [In Development]

**GreenTrigger** is a lightweight, on-demand service execution and management system designed to optimize resource usage by only running services when needed. It’s perfect for environments aiming to reduce their carbon footprint by keeping services idle until triggered, thereby saving resources.

## Features
- **On-Demand Service Execution**: GreenTrigger starts services only when necessary, reducing resource consumption.
- **HAProxy Integration**: Seamlessly integrates with HAProxy, dynamically adding service routes and managing them via HAProxy's API.
- **Service Configuration**: Define service execution commands and environment variables in a simple YAML configuration file.

## How It Works
1. **Configuration**: GreenTrigger reads a YAML configuration file that defines the services and their URLs.
2. **Service Execution**: When a service is triggered, GreenTrigger starts the corresponding service (e.g., a Java application).
3. **Dynamic Routing**: The service is added to HAProxy's routing table, making it accessible through predefined endpoints.
4. **Efficient Resource Usage**: Services only run when needed and can be shut down when no longer in use.

## Project Structure
\`\`\`bash
my-go-project/
├── main.go                   # Entry point of the GreenTrigger application
├── resources/
│   └── config-example.yml     # Example configuration file
├── services/
│   ├── config.go              # Configuration service to read YAML config
│   ├── executor.go            # Executes services on demand
│   └── haproxy.go             # Manages HAProxy API interaction
\`\`\`

## Setup and Usage

### 1. Clone the Repository:
\`\`\`bash
git clone https://github.com/your-username/greentrigger.git
cd greentrigger
\`\`\`

### 2. Install Dependencies:
Make sure Go is installed and run the following to install the required dependencies:
\`\`\`bash
go mod tidy
\`\`\`

### 3. Configure Your Services:
Modify the \`resources/config-example.yml\` file to define the services you want to run on demand:
\`\`\`yaml
services:
- name: hello-service
  url: /hello
  command: "java -jar /path/to/your/hello-world.jar"
  env:
  GLOBAL_VAR: "production"
  \`\`\`

### 4. Run GreenTrigger:
Pass the path to your configuration file as a parameter:
\`\`\`bash
go run main.go resources/config-example.yml
\`\`\`

### 5. HAProxy Setup:
Ensure that HAProxy is running and properly configured. GreenTrigger will dynamically add services to HAProxy when they are started.

### 6. License:
This project is open-source, licensed under MIT License.
