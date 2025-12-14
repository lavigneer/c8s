# Load Testing

Load and performance testing for C8S.

## Status

**Current Status**: Planned, not yet implemented

This directory is reserved for future load testing infrastructure.

## Planned Load Testing

### Tools Under Consideration

1. **k6** (Recommended)
   - Modern load testing tool
   - JavaScript-based test scripts
   - Great Kubernetes integration
   - Excellent reporting

2. **Locust**
   - Python-based
   - Distributed load testing
   - Web UI for monitoring

3. **Artillery**
   - Node.js based
   - Good for HTTP/WebSocket testing
   - YAML-based scenarios

### Test Scenarios to Implement

#### 1. Pipeline Creation Load Test
- **Goal**: Test how many pipelines can be created per second
- **Metrics**: API response time, controller lag, resource usage
- **Thresholds**: <100ms p95, <500ms p99

#### 2. Concurrent Pipeline Execution
- **Goal**: Test multiple pipelines running simultaneously
- **Metrics**: Job creation time, pod scheduling time, resource contention
- **Thresholds**: Successfully run 100 concurrent pipelines

#### 3. Log Streaming Stress Test
- **Goal**: Test SSE connections with high log volume
- **Metrics**: Connection stability, message delivery latency, memory usage
- **Thresholds**: 1000 concurrent SSE connections, <100ms message latency

#### 4. Artifact Upload/Download
- **Goal**: Test S3 operations under load
- **Metrics**: Upload/download throughput, S3 operation success rate
- **Thresholds**: 100MB/s throughput, 99.9% success rate

#### 5. Dashboard Load Test
- **Goal**: Test web UI under concurrent user load
- **Metrics**: Page load time, API response time, database query time
- **Thresholds**: <1s page load, <200ms API response

### Planned Directory Structure

```
tests/load/
├── README.md              # This file
├── k6/                    # k6 load test scripts
│   ├── pipeline-creation.js
│   ├── concurrent-pipelines.js
│   ├── log-streaming.js
│   ├── artifact-operations.js
│   └── dashboard-load.js
├── scenarios/             # Test scenario definitions
│   ├── baseline.yaml     # Baseline performance test
│   ├── stress.yaml       # Stress test (find breaking point)
│   ├── spike.yaml        # Spike test (sudden load increase)
│   └── soak.yaml         # Soak test (sustained load)
├── results/              # Test results (gitignored)
│   └── .gitkeep
└── scripts/              # Helper scripts
    ├── run-load-test.sh
    └── analyze-results.sh
```

## Example k6 Test (Future)

```javascript
// tests/load/k6/pipeline-creation.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 10 },   // Stay at 10 users
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<100', 'p(99)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function() {
  const url = 'http://localhost:8000/api/v1/pipelines';
  const payload = JSON.stringify({
    name: `test-pipeline-${Date.now()}`,
    steps: [
      { name: 'build', image: 'golang:1.25', commands: ['go build'] }
    ]
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ${TOKEN}',
    },
  };

  let res = http.post(url, payload, params);

  check(res, {
    'status is 201': (r) => r.status === 201,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });

  sleep(1);
}
```

## Running Load Tests (Future)

Once implemented:

```bash
# Install k6
brew install k6  # macOS
# or: https://k6.io/docs/getting-started/installation/

# Run specific test
k6 run tests/load/k6/pipeline-creation.js

# Run with custom VUs and duration
k6 run --vus 100 --duration 5m tests/load/k6/concurrent-pipelines.js

# Run with results output
k6 run --out json=results/pipeline-creation.json tests/load/k6/pipeline-creation.js

# Run in Kubernetes (distributed load testing)
k6 cloud run tests/load/k6/pipeline-creation.js
```

## Performance Baselines to Establish

| Metric | Target | Method |
|--------|--------|--------|
| Pipeline creation API | <100ms p95 | k6 HTTP load test |
| Pipeline execution start | <5s | Time from create to first pod running |
| Log streaming latency | <100ms | SSE message delivery time |
| Dashboard page load | <1s | k6 browser test |
| Concurrent pipelines | 100+ | Stress test |
| API throughput | 1000 req/s | k6 load test |

## Integration with CI/CD

Future GitHub Actions workflow:

```yaml
name: Load Tests

on:
  schedule:
    - cron: '0 2 * * 0'  # Weekly on Sunday at 2am
  workflow_dispatch:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6
      - name: Run load tests
        run: make load-test
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: load-test-results
          path: tests/load/results/
```

## Monitoring During Load Tests

Use these tools to monitor C8S during load testing:

```bash
# Watch resource usage
kubectl top pods -n c8s-system --watch

# Watch controller logs
kubectl logs -f deployment/c8s-controller -n c8s-system

# Watch API server metrics
curl http://localhost:8000/metrics | grep http_request

# Watch Kubernetes events
kubectl get events -n c8s-system --watch
```

## Contributing

When implementing load tests:

1. Start with baseline tests (low load)
2. Gradually increase load to find limits
3. Document thresholds and expected behavior
4. Include setup/teardown scripts
5. Add results interpretation guide
6. Update this README with actual usage

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [Load Testing Best Practices](https://k6.io/docs/testing-guides/load-testing-best-practices/)
- [Performance Testing Patterns](https://k6.io/docs/test-types/introduction/)

## Next Steps

To implement load testing:

1. Choose load testing tool (recommend k6)
2. Set up infrastructure (cluster, monitoring)
3. Implement baseline tests
4. Establish performance baselines
5. Add to CI/CD pipeline
6. Document results and thresholds
