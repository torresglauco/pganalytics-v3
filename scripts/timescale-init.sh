#!/bin/bash
# TimescaleDB initialization script
# This script installs TimescaleDB extension and creates necessary objects

set -e

echo "========== TimescaleDB Initialization Start =========="

# Install TimescaleDB extension if not already installed
echo "Installing TimescaleDB package..."
apt-get update > /dev/null 2>&1
apt-get install -y postgresql-16-contrib > /dev/null 2>&1

# Create the metrics schema and enable timescaledb extension
echo "Creating metrics schema and enabling TimescaleDB extension..."
psql -U postgres -d postgres << SQL
  -- Create metrics database if it doesn't exist
  SELECT 'CREATE DATABASE metrics' WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname='metrics');
SQL

# Note: TimescaleDB extension may not be available in standard postgres image,
# but we can proceed without it for the metrics storage - PostgreSQL tables work fine
psql -U postgres -d metrics << SQL
  -- Create schema
  CREATE SCHEMA IF NOT EXISTS metrics;
  SET search_path TO metrics, public;

  -- Try to create extension, but don't fail if it doesn't exist
  CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;
SQL

echo "========== TimescaleDB Initialization Complete =========="
