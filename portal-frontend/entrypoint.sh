#!/bin/sh
set -e

echo "=== Generating runtime configuration ==="

# 환경변수를 JavaScript 파일로 생성
cat <<EOF > /usr/share/nginx/html/env.js
window.ENV = {
  KEYCLOAK_URL: "${KEYCLOAK_URL:-https://keycloak.miribit.cloud}",
  KEYCLOAK_REALM: "${KEYCLOAK_REALM:-sso-demo}",
  KEYCLOAK_CLIENT_ID: "${KEYCLOAK_CLIENT_ID:-frontend}",
  KEYCLOAK_CLIENT_SECRET: "${KEYCLOAK_CLIENT_SECRET:-aSnWDRHlSNITRlME6uYgIkdTRmIxZk7j}",
  GRAFANA_URL: "${GRAFANA_URL:-https://grafana.miribit.cloud}",
  JENKINS_URL: "${JENKINS_URL:-https://jenkins.miribit.cloud}",
  ARGOCD_URL: "${ARGOCD_URL:-https://argocd.miribit.cloud}",
  BACKEND_URL: "${BACKEND_URL:-https://portal.miribit.cloud}",
  PORTAL_URL: "${PORTAL_URL:-https://portal.miribit.cloud}"
};
EOF

echo "Generated env.js with following configuration:"
cat /usr/share/nginx/html/env.js

# Nginx 프록시 설정도 동적으로 생성
echo ""
echo "=== Updating nginx proxy configuration ==="
if [ -n "$BACKEND_URL" ]; then
  echo "BACKEND_URL is set: $BACKEND_URL"
  echo "Replacing default proxy_pass with: $BACKEND_URL"
  sed -i "s|proxy_pass https://portal.miribit.cloud;|proxy_pass $BACKEND_URL;|g" /etc/nginx/nginx.conf
  
  # 변경 확인
  echo "Verifying nginx configuration update:"
  grep "proxy_pass" /etc/nginx/nginx.conf | head -5
else
  echo "WARNING: BACKEND_URL is not set, using default: https://portal.miribit.cloud"
fi

echo ""
echo "=== Configuration complete, starting Nginx ==="

# Nginx 실행
exec "$@"

