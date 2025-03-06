# Edu Dist Calculate API

This is a distributed system for calculating arithmetic expressions. The system consists of two main components:

- **Orchestrator**: A central server that receives expressions, breaks them into tasks, and manages their execution
  order.
- **Agent**: A computing component that fetches tasks from the orchestrator, performs calculations, and returns results.

> 🚧 Оно почти работает. Если возможно, зайди завтра - точно будет работать

## Architecture

The system uses a distributed architecture where:

1. Users submit expressions to the orchestrator via HTTP
2. The orchestrator parses expressions into tasks
3. Agents request tasks from the orchestrator, compute them, and return results
4. Users can check the status and retrieve results of their calculations

## Features

- Support for basic arithmetic operations: addition, subtraction, multiplication, and division
- Support for parenthesized expressions
- Distributed computation with configurable operation times
- Asynchronous processing with status tracking

## Getting Started

### Prerequisites

- Go 1.21 or higher
- hypermodeinc/badger (fs write permissions)

### Running the Application

#### Start the Orchestrator

```bash
go run cmd/orchestrator/main.go
```

#### Start the Agent

```bash
go run cmd/agent/main.go
```

You can start multiple agent instances to increase processing power.

### Environment Variables

#### Orchestrator

- `TIME_ADDITION_MS`: Time to process addition operations (default: 100ms)
- `TIME_SUBTRACTION_MS`: Time to process subtraction operations (default: 100ms)
- `TIME_MULTIPLICATION_MS`: Time to process multiplication operations (default: 100ms)
- `TIME_DIVISION_MS`: Time to process division operations (default: 100ms)

#### Agent

- `COMPUTING_POWER`: Number of goroutines for parallel processing (default: 1)

## API Reference

### Public API

#### Add Expression for Calculation

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2+2*2"
}'
```

Response:

```json
{
  "id": "cv4l4a3j3vq15tlsces0"
}
```

#### List Expressions

```bash
curl --location 'localhost:8080/api/v1/expressions'
```

Response:

```json
{
  "expressions": [
    {
      "id": "cv4l4a3j3vq15tlsces0",
      "status": "EXPRESSION_STATUS_COMPLETED",
      "result": 6
    }
  ]
}
```

#### Get Expression by ID

```bash
curl --location 'localhost:8080/api/v1/expressions/cv4l4a3j3vq15tlsces0'
```

Response:

```json
{
  "expression": {
    "id": "cv4l4a3j3vq15tlsces0",
    "status": "EXPRESSION_STATUS_COMPLETED",
    "result": 6
  }
}
```

### Internal API (Agent-Orchestrator Communication)

#### Get Task for Execution

```bash
curl --location 'localhost:8080/internal/task'
```

Response:

```json
{
  "task": {
    "id": "task-123",
    "arg1": 2,
    "arg2": 2,
    "operation": "TASK_OPERATION_MULTIPLICATION",
    "operation_time": {
      "seconds": 0,
      "nanos": 100000000
    }
  }
}
```

#### Submit Task Result

```bash
curl --location 'localhost:8080/internal/task' \
--header 'Content-Type: application/json' \
--data '{
    "id": "task-123",
    "result": 4
}'
```

## Project Structure

```
/
├── api/                     # Protocol buffer definitions
│   └── calculator/
│       └── v1/             
│           ├── calculator_public.proto   # Public API definitions
│           └── calculator_agent.proto    # Agent-Orchestrator API definitions
├── cmd/                     # Application entry points
│   ├── agent/               
│   │   └── main.go          # Agent main file
│   └── orchestrator/       
│       └── main.go          # Orchestrator main file
├── internal/                # Internal packages
│   ├── agent/              
│   │   └── ...              # Agent implementation
│   ├── calculator/          
│   │   ├── parser/          # Expression parsing
│   │   ├── repository/      # Data storage
│   │   └── service/         # Business logic
│   └── orchestrator/       
│       └── ...              # Orchestrator implementation
├── pkg/                     # Shared packages
│   └── ...
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Testing

Run all tests with:

```bash
go test ./...
```

## Examples

### Example 1: Simple Addition

```bash
# Submit expression
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression": "2+3"}'

# Response
{"id":"abc123"}

# Check status (immediately after submission)
curl --location 'localhost:8080/api/v1/expressions/abc123'

# Response
{"expression":{"id":"abc123","status":"EXPRESSION_STATUS_PENDING"}}

# Check status (after processing)
curl --location 'localhost:8080/api/v1/expressions/abc123'

# Response
{"expression":{"id":"abc123","status":"EXPRESSION_STATUS_COMPLETED","result":5}}
```

### Example 2: Complex Expression

```bash
# Submit expression
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression": "2*(3+4)/2"}'

# Response
{"id":"cv4l4a3j3vq15tlsces0"}

# Check status (after processing)
curl --location 'localhost:8080/api/v1/expressions/cv4l4a3j3vq15tlsces0'

# Response
{"expression":{"id":"def456","status":"EXPRESSION_STATUS_COMPLETED","result":7}}
```

### Example 3: Invalid Expression

```bash
# Submit invalid expression
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{"expression": "2+"}'

# Response
{"error":"Expression is not valid"}
```
