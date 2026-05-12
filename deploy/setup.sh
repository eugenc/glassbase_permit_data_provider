#!/bin/bash
set -e

curl -fsSL https://get.docker.com | sh
usermod -aG docker "${USER}" || true

apt-get update -y && apt-get install -y docker-compose-plugin

git clone https://github.com/echayko/glassbase_permit_data_provider.git /opt/glassbase || true
cd /opt/glassbase

cp .env.example .env
echo "Edit /opt/glassbase/.env with your ANTHROPIC_API_KEY and DATABASE_URL"

docker compose up -d

cat > /etc/logrotate.d/glassbase << 'EOF'
/var/log/glassbase/*.log {
    daily
    rotate 14
    compress
    missingok
    notifempty
}
EOF
