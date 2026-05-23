# Load Balancer — Learning Plan

Build a load balancer from scratch in Go. The goal is not a working implementation — it is to hit real problems, understand why they exist, and learn the concepts behind how production systems solve them.

## The Core Principle

> Build to understand, not to finish.

Each stage below is designed to manufacture a wall. The wall is where the learning happens. Do not skip walls by using AI or reading solutions prematurely.

---

## Stages

### Stage 1 — TCP Proxy (Single Backend)

```
Client → Your program → Single backend
```

Start here. Not load balancing yet. Just forward bytes between two TCP connections.

**Entry point:**
```go
func main() {
    listener, _ := net.Listen("tcp", ":8080")
    for {
        conn, _ := listener.Accept()
        go handleConn(conn)
    }
}
```

Implement `handleConn` yourself. Connect to a backend, forward bytes both ways.

**Wall you will hit:** if you forward client→backend first, you block. Backend→client never runs.

**Concept:** bidirectional concurrent I/O, goroutines for concurrent streams.

**Checklist:**
- [ ] TCP proxy works for single backend
- [ ] Test with `curl` and a simple Go HTTP server as backend
- [ ] Both directions (request and response) flow correctly

---

### Stage 2 — Multiple Backends + Round Robin

Add two backends. Pick one per request. Implement round-robin: request 1 → A, request 2 → B, request 3 → A...

**Wall you will hit:** your counter is a shared variable accessed by multiple goroutines. Wrong results under load.

**Concept:** race conditions, `sync/atomic` or `sync.Mutex`.

**Verify you hit it:**
```bash
go test -race ./...
```

**Checklist:**
- [ ] Two backends registered
- [ ] Round-robin selection implemented
- [ ] Race detector passes cleanly

---

### Stage 3 — Dead Backend

Kill one backend while requests are flowing.

**Wall you will hit:** dial to dead backend returns error. Client gets nothing or proxy crashes.

**Questions to answer yourself before reading any solution:**
- How do you detect a backend is dead?
- Do you check before sending or after it fails?
- If you retry on another backend, what happens if the request already partially sent?

**Concept:** error handling on dial, retry semantics, idempotency.

**Checklist:**
- [ ] Dead backend error is handled gracefully
- [ ] Client receives a proper error response, not a hang

---

### Stage 4 — Active Health Checking

Ping backends on a schedule. Remove dead ones from rotation automatically.

**Walls you will hit:**
- Backend responds to ping but hangs on real requests
- Backend recovers — how do you bring it back?
- Health check itself consumes a connection

**Questions to answer yourself:**
- What is the difference between a backend that is slow vs dead?
- How many failed checks before marking dead? Why not just one?
- How many successful checks before marking alive again?

**Concept:** liveness vs readiness, circuit breaker pattern.

**Checklist:**
- [ ] Health check goroutine runs on a ticker
- [ ] Dead backends removed from round-robin
- [ ] Recovered backends re-added to rotation

---

### Stage 5 — Slow Backend

Artificially slow down one backend:
```go
time.Sleep(10 * time.Second)
```

**Wall you will hit:** requests pile up on the slow backend. Fast backends are idle. Round-robin does not care about backend load.

**Questions to answer yourself:**
- How do you know a backend is slow vs dead?
- How do you measure slowness?
- What is a better algorithm than round-robin for unequal backends?

**Concept:** least connections, weighted round-robin, latency-based routing.

**Checklist:**
- [ ] Active connection count tracked per backend
- [ ] Least-connections selection implemented
- [ ] Slow backend receives fewer requests than fast ones under load

---

### Stage 6 — Graceful Shutdown

Handle `SIGTERM` without dropping in-flight requests.

**Wall you will hit:** in-flight requests get cut off when the process exits.

**Questions to answer yourself:**
- How do you stop accepting new connections while finishing existing ones?
- How do you know when all connections are done?
- What is the maximum time you are willing to wait?

**Concept:** connection draining, `context.Context` cancellation, `sync.WaitGroup`.

**Checklist:**
- [ ] `SIGTERM` stops accepting new connections
- [ ] Existing connections finish before process exits
- [ ] Configurable drain timeout

---

### Stage 7 — HTTP Awareness (Layer 7)

Route `/api/*` to backend group A and `/static/*` to backend group B.

**Wall you will hit:** you need to read the HTTP request to see the path — but reading it consumes bytes the backend also needs.

**Questions to answer yourself:**
- How do you read the request without consuming it?
- What is `bufio.Reader` and why does it exist?
- What if the client sends a request with no `Content-Length` header?

**Concept:** buffered I/O, HTTP parsing, chunked transfer encoding.

**Checklist:**
- [ ] HTTP request parsed to extract path
- [ ] Path-based routing to different backend groups
- [ ] Chunked transfer encoding handled correctly

---

## AI Usage Rules

| Situation | Rule |
|---|---|
| Starting a new stage | No AI. Attempt first. |
| Stuck less than 30 minutes | No AI. Keep thinking. |
| Stuck 30–60 minutes | Ask AI to verify your hypothesis only, not to solve it |
| Have a working solution | Ask AI "what edge cases did I miss?" |
| Reading stdlib source code | AI to verify your understanding after reading yourself |
| Syntax you forgot | AI fine — e.g. "how do I set a deadline on net.Conn?" |
| Race condition found | Fix it yourself first. AI to verify fix after. |
| Design decision | No AI. Write your options and tradeoffs first. |

**Never ask:** "how do I implement X?" — this skips the wall.

**Always ask:** "I think X happens because Y, is that correct?" — this verifies your understanding.

---

## Final Test

When done, answer these without AI or notes:

1. Why does round-robin fail for backends with unequal capacity?
2. What is the difference between L4 and L7 load balancing, and what does it cost?
3. Why do you need two goroutines per connection in a TCP proxy?
4. What happens to a slow client that receives data slower than your backend sends it?
5. How does NGINX's event loop differ from your goroutine-per-connection model?

If you can answer all five, you built something real.

---

## Concepts You Will Have Learned

- Bidirectional TCP stream forwarding
- Race conditions and safe concurrent access
- Health checking: liveness vs readiness
- Circuit breaker pattern
- Load balancing algorithms: round-robin, least connections, weighted
- Connection draining and graceful shutdown
- Buffered I/O and HTTP/1.1 parsing
- L4 vs L7 proxying tradeoffs
