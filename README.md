# Calculator API

A small REST calculator API written with Go's standard `net/http` package. The HTTP layer handles transport concerns and delegates all calculations and mathematical validation to an HTTP-independent service.

## Project structure

```text
cmd/api/          application entry point
internal/http/    routes, handlers, JSON requests and responses
internal/service/ calculator logic and unit tests
internal/domain/  reusable application errors
internal/model/   request and response models
```

## Run locally

Go 1.22 or newer is required.

```sh
go run ./cmd/api
```

The API listens on `http://localhost:8080`.

Run all tests with:

```sh
go test ./...
```

Generate test coverage file:

```sh
go test -coverprofile=coverage ./...
```

See test coverage:

```sh
go tool cover "-html=coverage"
```

## Docker

Build and run without installing Go on the host:

```sh
docker build -t calculator-api .
docker run --rm -p 8080:8080 calculator-api
```

## API

Every endpoint accepts `POST` requests with JSON and returns either `{"result": number}` or `{"error": "message"}`.

| Endpoint | Example payload |
| --- | --- |
| `/addition` | `{"operands":[10,4,2]}` |
| `/subtraction` | `{"operands":[10,4,2]}` |
| `/multiplication` | `{"operands":[10,4,2]}` |
| `/division` | `{"operands":[10,4,2]}` |
| `/exponentiation` | `{"base":10,"exponent":2}` |
| `/square-root` | `{"radicand":25}` |
| `/percentage` | `{"value":100,"percentage":10}` |

Example:

```sh
curl -X POST http://localhost:8080/addition \
  -H "Content-Type: application/json" \
  -d '{"operands":[10,4,2]}'
```

Operations with operand lists require at least two numbers and are evaluated from left to right. All inputs and results must be finite real numbers. Division rejects zero divisors, square root rejects negative radicands, and exponentiation rejects calculations without a representable real result.

# Architecture Decisions and Assumptions

Yes — and on your question: **with this architecture, we could later support expressions such as `10 + 21 + 12` or even `10 + 4 * 2 / 2` without rewriting the existing mathematical operations**.

The main addition would be a new expression/parsing layer that determines operation order and delegates the actual calculations to the existing calculator service. So the current design does not implement an expression engine, but it does not block one either.

Here is how I would document our **Architecture Decisions and Assumptions**:

- The API is intentionally designed around individual calculator operations rather than accepting arbitrary mathematical expressions.

- For operations that naturally support an arbitrary number of values — addition, subtraction, multiplication, and division — we use an `operands` array:

```json
{
  "operands": [10, 4, 2]
}
```

This avoids unnecessarily limiting the API to binary operations such as `10 + 4` and allows requests such as:

```text
10 + 4 + 2 + 8
```

The same request model can also be reused across operations where an ordered list of operands has a clear mathematical meaning.

- Operations using an `operands` array are evaluated **left-to-right**, preserving the order provided by the client.

For example:

```text
[10, 4, 2]

Addition:
(10 + 4) + 2

Subtraction:
(10 - 4) - 2

Multiplication:
(10 * 4) * 2

Division:
(10 / 4) / 2
```

- Division treats the first operand as the initial dividend. Therefore, the first operand may be zero, while every subsequent operand must be non-zero.

```text
[0, 10, 2]  -> valid
[10, 0, 2]  -> invalid
[10, 2, 0]  -> invalid
```

- We intentionally do **not** force every operation into the same request shape.

Operations such as exponentiation, square root, and percentage have different mathematical semantics, so their payloads use explicit field names rather than a generic `operands` array.

Exponentiation:

```json
{
  "base": 10,
  "exponent": 2
}
```

Square root:

```json
{
  "radicand": 25
}
```

Percentage:

```json
{
  "value": 100,
  "percentage": 10
}
```

This favors clarity of the API contract over artificial uniformity.

- We intentionally avoid a generic endpoint such as:

```text
POST /calculate

10 + 4 * 2 / 2
```

Supporting arbitrary expressions would introduce additional concerns such as:

```text
tokenization
parsing
operator precedence
parentheses
expression validation
syntax errors
```

That functionality is outside the intended scope of this assignment and would add complexity unrelated to the core requirements.

- The architecture nevertheless leaves room for such functionality in the future. An expression parser could later be introduced as a separate component that orchestrates the existing calculator operations:

```text
Expression
    ↓
Parser / Evaluator
    ↓
Calculator Service
    ├── Add
    ├── Subtract
    ├── Multiply
    ├── Divide
    ├── Exponentiate
    └── ...
```

The existing mathematical logic and validations could therefore remain reusable.

## HTTP and Service separation

The application uses two main architectural components:

```text
HTTP Request
    ↓
Handler
    ↓
Service
    ↓
Handler
    ↓
HTTP Response
```

The request, response, and returned result are not architectural layers themselves.

The **Handler** owns HTTP concerns:

- JSON decoding
- payload structure validation
- required-field validation
- unknown-field validation
- HTTP status codes
- JSON responses
- translation of domain errors into HTTP errors

The **Service** owns mathematical behavior:

- performing calculations
- validating mathematical constraints
- detecting invalid mathematical results
- returning either a result or a domain error

The service remains completely agnostic about HTTP. It does not know about `http.Request`, `http.ResponseWriter`, HTTP status codes, or JSON.

## Standard library over a framework

- The HTTP API uses Go's standard `net/http` package rather than Gin, Echo, Fiber, or another external framework.

- The current requirements do not require framework functionality that would justify introducing another dependency.

- Using `net/http` keeps the service small and reduces external dependencies without sacrificing the separation between transport and business logic.

- The HTTP implementation is isolated from the calculator service, so replacing `net/http` with another HTTP framework in the future would primarily affect the transport layer rather than the mathematical logic.

This is an important point: we're not saying **frameworks are bad**. We're saying **there is currently no problem that requires one**.

## Numeric assumptions

- Calculations operate on real numbers and use one consistent numeric representation.

- Integer-looking and decimal JSON values are treated as numeric operands.

- Negative values are accepted wherever mathematically valid.

- Percentage values are not restricted to `0–100`. Values such as `150%` or `-20%` are mathematically valid.

- Square root only accepts non-negative real radicands.

- Exponentiation accepts positive and negative bases/exponents whenever the result is a valid real number.

- A calculation must not return `NaN`, positive infinity, or negative infinity.

- If either an intermediate or final calculation cannot be represented by the chosen numeric type, the service returns an error rather than exposing an invalid numeric response.

This means we avoid arbitrary limits such as:

```text
exponent <= 100
percentage <= 100
base <= 100000
```

and instead reject combinations based on whether their mathematical result is valid and representable.

## API validation assumptions

- All calculator endpoints use `POST`.

- Required fields must actually be present; a numeric zero is not treated as equivalent to a missing field.

For example:

```json
{
  "value": 0,
  "percentage": 10
}
```

is different from:

```json
{
  "percentage": 10
}
```

- Malformed JSON returns `400 Bad Request`.

- Unexpected JSON fields are rejected rather than silently ignored.

For example:

```json
{
  "base": 10,
  "exponent": 2,
  "banana": 5
}
```

is considered an invalid request.

This makes the API contract strict and catches client-side mistakes such as misspelled field names.

## Maintainability

- Shared behavior is extracted only when it is genuinely shared.

Good candidates include:

```text
JSON decoding
JSON response writing
error response writing
OperandsRequest
ResultResponse
ErrorResponse
finite-result validation
```

- Mathematical operations remain separate functions rather than being forced into a generic `Calculate(operation, values...)` abstraction.

For example:

```text
Add
Subtract
Multiply
Divide
Exponentiate
SquareRoot
Percentage
```

This keeps each operation's semantics and validation rules explicit.

- Adding or changing one calculator operation should have minimal impact on unrelated operations.

- We avoid abstractions, layers, or interfaces that do not provide a concrete benefit.

## Testability

Each architectural boundary can be tested independently.

Service tests:

```text
Calculator Service
    ↓
result
```

They do not require HTTP, Docker, networking, or external dependencies.

Handler tests:

```text
HTTP request
    ↓
Handler
    ↓
Test double / mock service
    ↓
HTTP response
```

This allows handler behavior to be tested independently of the mathematical implementation.

The goal is **not** to mock every function call. Internal helpers belonging to the same unit can run normally.

The rule is:

> **Mock architectural collaborators, not every function another function happens to call.**

For this application, the meaningful test boundary is primarily:

```text
Handler → Calculator Service
```

## Scope assumptions

The service intentionally has:

```text
no database
no repository layer
no authentication
no persistence
no message queues
no expression parser
```

because none of those responsibilities exist in the stated problem.

And I think one sentence captures the overall approach very well for the submission:

> **The design favors explicit operation contracts and clear architectural boundaries over generic abstractions that are not required by the current problem.**

That is essentially what we've been doing throughout the design.
