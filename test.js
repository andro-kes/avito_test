import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

export const options = {
    stages: [
        { duration: "30s", target: 5 },
        { duration: "1m", target: 10 },
        { duration: "2m", target: 15 },
        { duration: "30s", target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.1'],
        'pr_creation_success_rate': ['rate>0.95'],
    },
};

// Кастомные метрики - теперь с правильным импортом
const createPRDuration = new Trend('create_pr_duration');
const prCreationSuccessRate = new Rate('pr_creation_success_rate');
const totalPRsCreated = new Counter('total_prs_created');
const errorCounter = new Counter('errors');

const BASE_URL = "http://localhost:8081";

const TEST_TEAM = {
    team_name: "performance_test_team",
    members: [
        { user_id: "perf_user_1", username: "Performance User 1", is_active: true },
        { user_id: "perf_user_2", username: "Performance User 2", is_active: true },
        { user_id: "perf_user_3", username: "Performance User 3", is_active: true },
        { user_id: "perf_user_4", username: "Performance User 4", is_active: true },
        { user_id: "perf_user_5", username: "Performance User 5", is_active: true }
    ]
};

export function setup() {
    console.log("🔧 Setting up test data...");
    
    const teamRes = http.post(`${BASE_URL}/team/add/`, JSON.stringify(TEST_TEAM), {
        headers: { 'Content-Type': 'application/json' }
    });
    
    console.log(`Team creation status: ${teamRes.status}`);
    
    if (teamRes.status === 201) {
        console.log("✅ Test team created successfully");
    } else if (teamRes.status === 400) {
        console.log("ℹ️ Test team already exists");
    } else {
        console.error(`❌ Failed to create test team: ${teamRes.status} - ${teamRes.body}`);
    }
    
    sleep(2);
}

export default function () {
    const randomUser = TEST_TEAM.members[Math.floor(Math.random() * TEST_TEAM.members.length)];
    
    const startTime = Date.now();
    const success = createPR(randomUser.user_id);
    const endTime = Date.now();
    
    createPRDuration.add(endTime - startTime);
    prCreationSuccessRate.add(success);
    
    if (success) {
        totalPRsCreated.add(1);
    } else {
        errorCounter.add(1);
    }
    
    sleep(Math.random() * 2 + 1);
}

function createPR(authorId) {
    const timestamp = Date.now();
    const uniqueId = Math.random().toString(36).substr(2, 5);
    
    const payload = {
        pull_request_id: `perf_pr_${timestamp}_${uniqueId}`,
        pull_request_name: `Performance Test PR ${timestamp}`,
        author_id: authorId,
    };
    
    const res = http.post(`${BASE_URL}/pullRequest/create/`, JSON.stringify(payload), {
        headers: { 'Content-Type': 'application/json' },
        timeout: '10s'
    });
    
    const success = check(res, {
        'PR created successfully': (r) => r.status === 201,
        'Response time under 1s': (r) => r.timings.duration < 1000,
    });
    
    if (res.status === 201) {
        console.log(`✅ PR created by ${authorId}`);
        return true;
    } else if (res.status === 409) {
        console.log(`ℹ️ PR already exists`);
        return false;
    } else {
        console.log(`❌ Failed: ${res.status} - ${res.body}`);
        return false;
    }
}

// Безопасная функция для получения значений метрик
function getMetricValue(metrics, path, defaultValue = 0) {
    try {
        const keys = path.split('.');
        let value = metrics;
        for (const key of keys) {
            value = value[key];
            if (value === null || value === undefined) return defaultValue;
        }
        return value;
    } catch (error) {
        return defaultValue;
    }
}

export function handleSummary(data) {
    const timestamp = new Date().toLocaleString();
    
    // Безопасное извлечение метрик
    const totalRequests = getMetricValue(data, 'metrics.http_reqs.values.count', 0);
    const successRate = getMetricValue(data, 'metrics.checks.values.rate', 0) * 100;
    const avgResponseTime = getMetricValue(data, 'metrics.http_req_duration.values.avg', 0);
    const failedRate = getMetricValue(data, 'metrics.http_req_failed.values.rate', 0) * 100;
    const testDuration = data.state ? (data.state.testRunDuration / 1000000000).toFixed(2) : 0;
    
    const durationMetrics = {
        min: getMetricValue(data, 'metrics.http_req_duration.values.min', 0),
        med: getMetricValue(data, 'metrics.http_req_duration.values.med', 0),
        p90: getMetricValue(data, 'metrics.http_req_duration.values.p90', 0),
        p95: getMetricValue(data, 'metrics.http_req_duration.values.p95', 0),
        max: getMetricValue(data, 'metrics.http_req_duration.values.max', 0)
    };
    
    const checks = {
        passes: getMetricValue(data, 'metrics.checks.values.passes', 0),
        fails: getMetricValue(data, 'metrics.checks.values.fails', 0)
    };
    
    const htmlReport = `
<!DOCTYPE html>
<html>
<head>
    <title>K6 Load Test Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .charts {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin: 20px 0;
        }
        .chart-container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #ddd;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        .metric-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            margin: 10px 0;
        }
        .metric-label {
            font-size: 14px;
            opacity: 0.9;
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .summary {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 K6 Load Test Report</h1>
        <div class="summary">
            <p><strong>Generated:</strong> ${timestamp}</p>
            <p><strong>Test Duration:</strong> ${testDuration}s</p>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Total Requests</div>
                <div class="metric-value">${totalRequests}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Success Rate</div>
                <div class="metric-value">${successRate.toFixed(1)}%</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Avg Response Time</div>
                <div class="metric-value">${avgResponseTime.toFixed(2)}ms</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Failed Requests</div>
                <div class="metric-value">${failedRate.toFixed(1)}%</div>
            </div>
        </div>

        <div class="charts">
            <div class="chart-container">
                <h3>Response Time Percentiles</h3>
                <canvas id="percentileChart"></canvas>
            </div>
            <div class="chart-container">
                <h3>Request Success/Failure</h3>
                <canvas id="requestsChart"></canvas>
            </div>
        </div>

        <div class="chart-container">
            <h3>Response Time Distribution</h3>
            <canvas id="distributionChart"></canvas>
        </div>
    </div>

    <script>
        // Percentiles Chart
        const percentiles = ['Min', 'Median', 'p90', 'p95', 'Max'];
        const percentileValues = [
            ${durationMetrics.min},
            ${durationMetrics.med},
            ${durationMetrics.p90},
            ${durationMetrics.p95},
            ${durationMetrics.max}
        ];

        new Chart(document.getElementById('percentileChart'), {
            type: 'bar',
            data: {
                labels: percentiles,
                datasets: [{
                    label: 'Response Time (ms)',
                    data: percentileValues,
                    backgroundColor: 'rgba(54, 162, 235, 0.5)',
                    borderColor: 'rgba(54, 162, 235, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

        // Requests Chart
        const successCount = ${checks.passes};
        const failCount = ${checks.fails};
        
        new Chart(document.getElementById('requestsChart'), {
            type: 'doughnut',
            data: {
                labels: ['Success', 'Failed'],
                datasets: [{
                    data: [successCount, failCount],
                    backgroundColor: ['#4CAF50', '#F44336']
                }]
            },
            options: {
                responsive: true
            }
        });

        // Distribution Chart (simulated based on percentiles)
        new Chart(document.getElementById('distributionChart'), {
            type: 'line',
            data: {
                labels: ['0%', '25%', '50%', '75%', '90%', '95%', '100%'],
                datasets: [{
                    label: 'Response Time Distribution',
                    data: [
                        ${durationMetrics.min},
                        ${durationMetrics.min + (durationMetrics.med - durationMetrics.min) * 0.5},
                        ${durationMetrics.med},
                        ${durationMetrics.med + (durationMetrics.p90 - durationMetrics.med) * 0.5},
                        ${durationMetrics.p90},
                        ${durationMetrics.p95},
                        ${durationMetrics.max}
                    ],
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1,
                    fill: true,
                    backgroundColor: 'rgba(255, 99, 132, 0.1)'
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    </script>
</body>
</html>`;
    
    return {
        "k6_report.html": htmlReport,
        "stdout": "text-summary"
    };
}