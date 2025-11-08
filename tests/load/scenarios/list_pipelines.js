/**
 * C8S Load Test: Pipeline Listing
 *
 * Purpose: Test API performance under read-heavy workload
 * Workload: List 1000+ pipeline runs with concurrent users
 * Concurrency: 100 users (configurable)
 * Duration: 5 minutes (configurable)
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';

// Configuration
const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'test_token_user';

// Test options
export const options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp up to 20 users
    { duration: '2m30s', target: 100 }, // Stay at 100 users
    { duration: '30s', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.01'],
  },
};

/**
 * List pipeline runs with various filters
 */
export default function () {
  const headers = {
    'Authorization': `Bearer ${AUTH_TOKEN}`,
    'Content-Type': 'application/json',
  };

  group('List Pipelines - No Filter', function () {
    const params = {
      headers: headers,
    };
    let response = http.get(`${BASE_URL}/api/projects/default/runs`, params);
    check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
      'has runs in response': (r) => r.body.includes('runs') || r.body.includes('status'),
    });
  });

  sleep(1);

  group('List Pipelines - By Status', function () {
    const params = {
      headers: headers,
    };
    let response = http.get(`${BASE_URL}/api/projects/default/runs?status=success`, params);
    check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
      'filtered results': (r) => r.body.includes('status'),
    });
  });

  sleep(1);

  group('List Pipelines - By Branch', function () {
    const params = {
      headers: headers,
    };
    let response = http.get(`${BASE_URL}/api/projects/default/runs?branch=main`, params);
    check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
    });
  });

  sleep(1);

  group('List Pipelines - Paginated', function () {
    const params = {
      headers: headers,
    };
    let response = http.get(`${BASE_URL}/api/projects/default/runs?page=1&per_page=20`, params);
    check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
      'pagination metadata': (r) => r.body.includes('page') || r.body.includes('total'),
    });
  });

  sleep(2);
}
