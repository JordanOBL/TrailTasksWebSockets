# TrailTasks WebSockets

TrailTasks WebSockets is a Go server that exposes a WebSocket endpoint for managing
collaborative "hiking" sessions. Clients connect to `/groupsession` and exchange
JSON messages that follow a simple protocol for creating rooms, joining them,
and controlling the session timer.

## Installation

1. Install [Go](https://golang.org/doc/install) version 1.23 or later.
2. Clone this repository.
3. Download the dependencies:
   ```bash
   go mod download
   ```

## Running the Server

Start the server on port `8080` using:

```bash
go run ./cmd/server
```

The server listens on `ws://localhost:8080/groupsession`.

## WebSocket Protocols

Messages are JSON objects with a `header` and a `message`. The `header` contains
the `protocol` name, `roomId`, and `userId`. Depending on the protocol, the
`message` may contain additional fields. Key protocols include:

- `create`: create a new room and become its host.
- `join`: join an existing room.
- `ready`: toggle the ready state of a hiker.
- `updateConfig`: update session and timer settings.
- `start`: begin the session timer.
- `pause` / `resume`: pause or resume a hiker's progress.
- `skipBreak`: skip the current break period.
- `extraSet` / `extraSession`: extend the session with additional sets or
  sessions.
- `end`: stop the current session.

Server responses mirror these protocols to broadcast updates or send direct
messages back to a client.

## Running Tests

Execute the Go test suite from the repository root:

```bash
go test ./...
```

The tests require the Go toolchain to fetch dependencies.
