#!/bin/bash

echo "🚀 Starting test environment (DB + API only)..."

# Останавливаем и очищаем предыдущие контейнеры
docker compose -f docker-compose.yml -f docker-compose.test.yml down --rmi local --remove-orphans

# Запускаем только postgres_test и api_test
docker compose -f docker-compose.yml -f docker-compose.test.yml up --build postgres_test api_test -d

echo "⏳ Waiting for services to be ready..."
sleep 10

echo "✅ Test environment is ready!"
echo "📊 API: http://localhost:8081"
echo "🗄️  DB: localhost:5433"
echo ""
echo "🧪 Run tests with: k6 run test.js"
echo "📈 Report will be saved as: k6_report.html"