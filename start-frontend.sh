#!/bin/bash

echo "🚀 Starting pgAnalytics Frontend"
echo "================================"
echo ""

cd frontend

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
  echo "📦 Installing dependencies..."
  npm install
  echo ""
fi

echo "🌐 Starting development server..."
echo ""
echo "The frontend will be available at: http://localhost:3000"
echo ""
echo "Demo Credentials:"
echo "  Username: demo"
echo "  Password: Demo@12345"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

npm run dev
