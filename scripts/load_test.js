// GigCo API Load Testing Script using k6
// Install: brew install k6 (macOS) or download from https://k6.io/
// Run: k6 run scripts/load_test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const jobListDuration = new Trend('job_list_duration');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

// Test configuration
export const options = {
    // Stages define load ramping
    stages: [
        { duration: '30s', target: 10 },   // Ramp up to 10 users
        { duration: '1m', target: 10 },    // Stay at 10 users
        { duration: '30s', target: 50 },   // Ramp up to 50 users
        { duration: '2m', target: 50 },    // Stay at 50 users
        { duration: '30s', target: 100 },  // Ramp up to 100 users
        { duration: '2m', target: 100 },   // Stay at 100 users
        { duration: '1m', target: 0 },     // Ramp down
    ],

    // Thresholds define success criteria
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],  // 95% under 500ms, 99% under 1s
        http_req_failed: ['rate<0.01'],                   // Error rate under 1%
        errors: ['rate<0.05'],                            // Custom error rate under 5%
    },
};

// Test data
const testUsers = [
    { email: 'worker1@gigco.dev', password: 'password123' },
    { email: 'consumer1@gigco.dev', password: 'test123' },
];

// Helper to get random test user
function getRandomUser() {
    return testUsers[Math.floor(Math.random() * testUsers.length)];
}

// Setup function - runs once per VU
export function setup() {
    // Verify the server is reachable
    const healthRes = http.get(`${BASE_URL}/health`);
    check(healthRes, {
        'health check passed': (r) => r.status === 200,
    });

    if (healthRes.status !== 200) {
        throw new Error('Server health check failed');
    }

    console.log(`Load test targeting: ${BASE_URL}`);
    return { baseUrl: BASE_URL };
}

// Default function - runs for each VU
export default function (data) {
    let authToken = null;

    // Test group: Authentication
    group('Authentication', function () {
        // Login test
        const user = getRandomUser();
        const loginPayload = JSON.stringify({
            email: user.email,
            password: user.password,
        });

        const loginStart = Date.now();
        const loginRes = http.post(`${API_URL}/auth/login`, loginPayload, {
            headers: { 'Content-Type': 'application/json' },
        });
        loginDuration.add(Date.now() - loginStart);

        const loginSuccess = check(loginRes, {
            'login status is 200': (r) => r.status === 200,
            'login has token': (r) => {
                try {
                    const body = JSON.parse(r.body);
                    return body.token && body.token.length > 0;
                } catch {
                    return false;
                }
            },
        });

        errorRate.add(!loginSuccess);

        if (loginSuccess && loginRes.status === 200) {
            try {
                authToken = JSON.parse(loginRes.body).token;
            } catch (e) {
                console.error('Failed to parse login response');
            }
        }

        sleep(1);
    });

    // Test group: Jobs API
    if (authToken) {
        group('Jobs API', function () {
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`,
            };

            // List all jobs
            const listStart = Date.now();
            const listRes = http.get(`${API_URL}/jobs`, { headers });
            jobListDuration.add(Date.now() - listStart);

            const listSuccess = check(listRes, {
                'job list status is 200': (r) => r.status === 200,
                'job list returns array': (r) => {
                    try {
                        const body = JSON.parse(r.body);
                        return Array.isArray(body) || (body.jobs && Array.isArray(body.jobs));
                    } catch {
                        return false;
                    }
                },
            });

            errorRate.add(!listSuccess);
            sleep(0.5);

            // Get available jobs
            const availableRes = http.get(`${API_URL}/jobs/available`, { headers });
            check(availableRes, {
                'available jobs status is 200': (r) => r.status === 200,
            });

            sleep(0.5);
        });
    }

    // Test group: Public endpoints
    group('Public Endpoints', function () {
        // Health check
        const healthRes = http.get(`${BASE_URL}/health`);
        check(healthRes, {
            'health check returns 200': (r) => r.status === 200,
        });

        sleep(0.5);
    });

    // Random sleep to simulate real user behavior
    sleep(Math.random() * 2 + 1);
}

// Teardown function - runs once after all VUs complete
export function teardown(data) {
    console.log('Load test completed');
}

// Separate scenario for stress testing
export const stressTest = {
    executor: 'ramping-arrival-rate',
    startRate: 1,
    timeUnit: '1s',
    preAllocatedVUs: 50,
    maxVUs: 200,
    stages: [
        { duration: '2m', target: 10 },   // Start with 10 req/s
        { duration: '5m', target: 50 },   // Ramp to 50 req/s
        { duration: '2m', target: 100 },  // Spike to 100 req/s
        { duration: '5m', target: 50 },   // Back to 50 req/s
        { duration: '2m', target: 0 },    // Ramp down
    ],
};

// Separate scenario for soak testing (long duration)
export const soakTest = {
    executor: 'constant-vus',
    vus: 50,
    duration: '30m',
};
