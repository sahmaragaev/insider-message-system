# Circuit Breaker Implementation

## Overview
The circuit breaker protects against external service failures by automatically stopping requests when services are down and allowing recovery when they come back online.

## Why Use It?
- Prevents cascading failures when webhooks fail
- Stops wasting resources on broken services
- Automatically recovers when services are healthy again

## How It Works

**CLOSED** → Normal operation, requests go through
**OPEN** → Service is down, requests are rejected immediately  
**HALF-OPEN** → Testing if service recovered, limited requests allowed

## Configuration

```bash
CIRCUIT_BREAKER_ENABLED=true
CIRCUIT_BREAKER_FAILURE_RATE=0.5
CIRCUIT_BREAKER_MIN_REQUESTS=10
CIRCUIT_BREAKER_HALF_OPEN_AFTER=30s
```

| Setting | Default | Description |
|---------|---------|-------------|
| `failure_rate` | 0.5 | Open circuit when 50% of requests fail |
| `min_requests` | 10 | Need at least 10 requests before opening |
| `half_open_after` | 30s | Wait 30s before testing recovery |

## Usage

The circuit breaker automatically protects webhook calls:

```go
response, err := webhookClient.SendMessage(ctx, request)
if err == circuitbreaker.ErrCircuitOpen {
    return errors.New("webhook service temporarily unavailable")
}
```

## Monitoring

Check circuit breaker status:
```bash
GET /api/v1/circuit-breaker/status
```

Response includes current state, failure counts, and configuration.

## Error Handling

When the circuit is open, webhook calls return:
```json
{
  "code": "WEBHOOK_CIRCUIT_OPEN",
  "message": "Webhook service is temporarily unavailable"
}
```

Handle these errors gracefully by queuing messages for later or showing user-friendly error messages.

## Best Practices

1. **Set reasonable thresholds** - Don't make it too sensitive
2. **Monitor the status endpoint** - Set up alerts for state changes  
3. **Handle errors gracefully** - Queue failed messages for retry
4. **Test with failing services** - Make sure it works as expected

## Troubleshooting

**Circuit won't open?** Check if it's enabled and you have enough failed requests.
**Stuck open?** Verify the external service is actually working again.
**Too many failures?** Check network connectivity and service health.