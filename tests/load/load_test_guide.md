# C8S Load Testing Guide

**Version**: 1.0
**Date**: 2025-11-02
**Status**: Testing Strategy Defined

## Overview

This guide provides detailed instructions for performing load testing on C8S. Load testing helps identify performance bottlenecks and validate that the system can handle expected production workloads.

## Prerequisites

### Option 1: Using k6 (Recommended)

```bash
# Install k6
# macOS
brew install k6

# Linux
sudo apt-get install k6

# Windows
choco install k6
```

### Option 2: Using Go's `testing/bench` Package

```bash
# Build custom load testing tool
go build -o bin/load-test ./tests/load/cmd/load-test
```

### Option 3: Using Apache JMeter

```bash
# Install JMeter
brew install jmeter

# Or download from: https://jmeter.apache.org/
```

## Environment Setup

### 1. Start C8S Development Server

```bash
# Terminal 1: Start controller
make run-controller

# Terminal 2: Start API server
make run-api-server

# Terminal 3: Start dashboard (optional)
make run-dashboard
```

### 2. Configure Test Parameters

Create `.env.test`:
```bash
API_BASE_URL=http://localhost:8080
AUTH_TOKEN=test_token_user
LOAD_DURATION=300  # 5 minutes
NUM_CONCURRENT=100
RAMP_UP_TIME=30    # 30 seconds
```

## Load Testing Scenarios

### Scenario 1: Pipeline Listing (Heavy Read)

**Purpose**: Test API performance under read-heavy workload
**Workload**: List 1000+ pipeline runs
**Concurrency**: 100 users
**Duration**: 5 minutes

```bash
# Using k6
k6 run tests/load/scenarios/list_pipelines.js \
  --vus 100 \
  --duration 5m \
  --ramp-up 30s
```

**Success Criteria**:
- p95 response time < 500ms
- p99 response time < 1000ms
- Error rate < 1%
- Throughput > 100 req/s

### Scenario 2: Pipeline Creation

**Purpose**: Test write performance
**Workload**: Create new pipeline runs
**Concurrency**: 10 concurrent creators
**Duration**: 5 minutes

```bash
# Using k6
k6 run tests/load/scenarios/create_pipelines.js \
  --vus 10 \
  --duration 5m \
  --ramp-up 30s
```

**Success Criteria**:
- p95 response time < 1000ms
- Success rate > 99%
- No duplicate creation

### Scenario 3: Log Streaming (Real-time)

**Purpose**: Test Server-Sent Events performance
**Workload**: Stream logs to 50 concurrent clients
**Duration**: 10 minutes

```bash
# Using k6 (requires SSE extension)
k6 run tests/load/scenarios/stream_logs.js \
  --vus 50 \
  --duration 10m
```

**Success Criteria**:
- Log delivery latency < 100ms
- Connection stability > 99.5%
- No dropped connections

### Scenario 4: Artifact Download

**Purpose**: Test file transfer performance
**Workload**: Download various artifact sizes
**Concurrency**: 20 concurrent downloads
**Duration**: 5 minutes

```bash
# Using k6
k6 run tests/load/scenarios/download_artifacts.js \
  --vus 20 \
  --duration 5m
```

**Success Criteria**:
- Throughput > 50MB/s
- Connection stability > 99%
- Error rate < 1%

### Scenario 5: Spike Test

**Purpose**: Test system resilience to traffic spikes
**Workload**: Sudden 500 concurrent requests
**Duration**: 2 minutes

```bash
# Using k6
k6 run tests/load/scenarios/spike_test.js \
  --vus 500 \
  --duration 2m
```

**Success Criteria**:
- System recovers within 2 minutes
- No data corruption
- Error rate acceptable (< 10% is OK for spikes)

## Performance Baselines

### Target Metrics

| Metric | Target | Method |
|--------|--------|--------|
| List Pipelines (p95) | < 500ms | `list_pipelines.js` |
| Create Pipeline (p95) | < 1000ms | `create_pipelines.js` |
| Log Stream Latency | < 100ms | `stream_logs.js` |
| Artifact Download Speed | > 50MB/s | `download_artifacts.js` |
| Max Concurrent Users | > 100 | `list_pipelines.js` |
| Recovery Time (spike) | < 2 min | `spike_test.js` |

## Analysis & Reporting

### Metrics to Collect

1. **Response Time**
   - Minimum
   - Maximum
   - Mean
   - Median (p50)
   - p95 (95th percentile)
   - p99 (99th percentile)

2. **Throughput**
   - Requests per second
   - Bytes per second
   - Success rate

3. **Errors**
   - Error count
   - Error types
   - Error rate

4. **Resource Usage**
   - CPU utilization
   - Memory usage
   - Network bandwidth

### Generating Reports

**Using k6**:
```bash
# HTML report
k6 run --out=html=results.html tests/load/scenarios/list_pipelines.js

# JSON report
k6 run --out=json=results.json tests/load/scenarios/list_pipelines.js

# InfluxDB + Grafana (real-time)
k6 run --out=influxdb=http://localhost:8086/myk8sdb \
  tests/load/scenarios/list_pipelines.js
```

**View Results**:
```bash
# HTML in browser
open results.html

# JSON analysis (custom script)
jq '.data.summary' results.json
```

## Troubleshooting

### Issue: "Connection refused"

**Solution**: Ensure API server is running
```bash
make run-api-server
```

### Issue: "High error rate"

**Possible causes**:
- Server overloaded (reduce concurrent users)
- Invalid auth token (update `.env.test`)
- Network issues (check connectivity)

### Issue: "Timeout errors"

**Solutions**:
1. Increase timeout in k6 script
2. Reduce concurrent load
3. Check server resource usage

### Issue: "Memory leaks suspected"

**Monitoring**:
```bash
# Watch memory usage
watch -n 1 'kubectl top pod -n default | grep c8s'

# Check for goroutine leaks
curl localhost:6060/debug/pprof/goroutine
```

## Load Testing Best Practices

1. **Test in Staging**: Never load test production
2. **Baseline First**: Establish baseline before changes
3. **Gradual Ramp-up**: Ramp up load gradually to avoid connection storms
4. **Multiple Scenarios**: Test different workload patterns
5. **Monitor Resources**: Track CPU, memory, network during tests
6. **Document Results**: Record all test results for comparison
7. **Repeat Tests**: Run tests multiple times for consistency
8. **Peak Hours**: Test during expected peak usage times
9. **Failure Handling**: Include failure scenarios (failed requests)
10. **Realistic Data**: Use realistic data sizes and patterns

## Continuous Load Testing

### CI/CD Integration

```yaml
# Example GitHub Actions workflow
name: Load Testing
on: [push, schedule]
jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install k6
        run: sudo apt-get install k6
      - name: Start services
        run: docker-compose up -d
      - name: Run load tests
        run: k6 run tests/load/scenarios/list_pipelines.js
```

### Scheduled Testing

```bash
# Weekly load test (cron)
0 2 * * 0 /usr/local/bin/k6 run /opt/c8s/tests/load/scenarios/list_pipelines.js
```

## Troubleshooting Checklist

- [ ] API server is running
- [ ] Database is accessible
- [ ] Storage is accessible
- [ ] Network connectivity is good
- [ ] Authentication token is valid
- [ ] Sufficient system resources
- [ ] No rate limiting in effect
- [ ] Firewall not blocking requests
- [ ] SSL/TLS certificates valid (if HTTPS)
- [ ] Load test script is syntactically correct

## Next Steps

1. Run baseline load tests
2. Identify performance bottlenecks
3. Optimize identified issues
4. Re-run tests to verify improvements
5. Document final results
6. Archive test results for comparison

## Related Documentation

- [PHASE3_TESTING_VALIDATION_PLAN.md](../../PHASE3_TESTING_VALIDATION_PLAN.md) - Full testing plan
- [PERFORMANCE_BASELINES.md](./PERFORMANCE_BASELINES.md) - Baseline metrics
- [CONFIGURATION.md](../../docs/CONFIGURATION.md) - Configuration tuning
