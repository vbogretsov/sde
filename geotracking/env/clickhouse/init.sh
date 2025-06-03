#!/bin/bash
set -e

clickhouse-client -n -u admin --password $CLICKHOUSE_ADMIN_PASSWORD <<-EOF

CREATE DATABASE IF NOT EXISTS app ON CLUSTER cluster;

EOF
