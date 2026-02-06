# Performance Logging

## Overview
The socket server includes detailed performance timing logs that can be enabled to monitor the performance of various operations. These logs help identify bottlenecks and optimize the server's performance.

## Enabling Performance Logs

### Command Line Flag
```bash
./bin/socket-server --performance-logs --port 8003 --server-token YOUR_TOKEN --jwt-secret YOUR_SECRET
```

### Environment Variable
```bash
export PERFORMANCE_LOGS=true
./bin/socket-server
```

### Default Behavior
By default, performance logs are **disabled** to reduce log noise in production environments.

## What Gets Logged

When performance logging is enabled, the following metrics are tracked:

### HTTP API Operations
- **JSON decode time** - Time to parse incoming JSON payloads
- **Message creation time** - Time to construct message objects
- **Broadcast type detection** - Time to determine broadcast type
- **Broadcast operation time** - Time to execute the broadcast
- **Response generation time** - Time to generate HTTP response
- **Total request time** - End-to-end request handling time

### WebSocket Broadcasting
- **Channel lookup time** - Time to find channel
- **Client collection time** - Time to gather all clients
- **Individual client send time** - Time per client send (warns if > 10ms)
- **Concurrent sending time** - Total time for parallel sends
- **Total broadcast time** - End-to-end broadcast time

### Laravel Integration
- **Laravel ping dispatch time** - Time to dispatch ping events to Laravel

## Performance Log Format

Performance logs are prefixed with emoji indicators:
- ⏱️ - Timing measurement
- ⚠️ - Slow operation warning
- 🏁 - Total/final timing

Example output:
```
⏱️ JSON decode took: 234µs
⏱️ Message creation took: 12µs
⏱️ Broadcast type detection took: 5µs
⏱️ Channel lookup took: 45µs
⏱️ Getting clients took: 23µs
⏱️ Concurrent sending to 5 clients took: 2.3ms (success: 5)
🏁 Total broadcast request took: 2.8ms
```

## Slow Operation Warnings

When performance logs are enabled, the server will warn about operations that exceed thresholds:
- **Client sends > 10ms** - Individual client message delivery taking too long
- **Broadcast timeouts** - When broadcast operations exceed 1 second

## Use Cases

### Development
Enable performance logs during development to:
- Profile application performance
- Identify slow operations
- Optimize broadcast patterns
- Debug latency issues

### Production Debugging
Temporarily enable in production to:
- Investigate performance complaints
- Monitor specific operations
- Gather metrics for optimization
- Troubleshoot slow clients

### Load Testing
Use during load tests to:
- Measure throughput
- Identify bottlenecks
- Validate optimizations
- Compare before/after performance

## Best Practices

1. **Don't enable in production by default** - Performance logs add overhead and verbosity
2. **Use for targeted debugging** - Enable temporarily when investigating issues
3. **Monitor slow client warnings** - Address consistently slow clients
4. **Track broadcast times** - Ensure broadcasts complete within acceptable timeframes
5. **Correlate with application metrics** - Compare with Laravel app performance

## Performance Impact

Enabling performance logs has minimal overhead:
- Uses Go's `time.Since()` for nanosecond precision
- Only logs when flag is enabled (zero overhead when disabled)
- Conditional checks are optimized by Go compiler

Estimated overhead: < 1% when enabled
