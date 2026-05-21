#!/bin/bash

# 开发环境自签名证书生成脚本
# 用途: 本地开发/测试环境 HTTPS 配置

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/../ssl/dev"

echo "=== Industrial AI 开发环境证书生成 ==="
echo ""

# 创建证书目录
mkdir -p "${CERT_DIR}"

echo "1. 生成 CA 证书 (开发环境根证书)"
openssl req -x509 -new -nodes \
    -keyout "${CERT_DIR}/ca.key" \
    -out "${CERT_DIR}/ca.crt" \
    -days 3650 \
    -subj "/C=CN/ST=Shanghai/L=Shanghai/O=IndustrialAI-Dev-CA/CN=Industrial AI Dev CA" \
    -sha256

echo "   ✓ CA 证书已生成"
echo ""

echo "2. 生成服务器证书"
# 创建扩展配置文件
cat > "${CERT_DIR}/server.ext" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
DNS.3 = industrial-ai.local
DNS.4 = *.industrial-ai.local
DNS.5 = backend
DNS.6 = frontend
DNS.7 = *.industrial-ai.example.com
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# 生成服务器私钥
openssl genrsa -out "${CERT_DIR}/server.key" 4096

# 生成 CSR
openssl req -new \
    -key "${CERT_DIR}/server.key" \
    -out "${CERT_DIR}/server.csr" \
    -subj "/C=CN/ST=Shanghai/L=Shanghai/O=IndustrialAI-Dev/CN=localhost" \
    -sha256

# 使用 CA 签名
openssl x509 -req \
    -in "${CERT_DIR}/server.csr" \
    -CA "${CERT_DIR}/ca.crt" \
    -CAkey "${CERT_DIR}/ca.key" \
    -CAcreateserial \
    -out "${CERT_DIR}/server.crt" \
    -days 365 \
    -extfile "${CERT_DIR}/server.ext" \
    -sha256

# 合并证书链
cat "${CERT_DIR}/server.crt" "${CERT_DIR}/ca.crt" > "${CERT_DIR}/fullchain.pem"

# 复制私钥为标准名称
cp "${CERT_DIR}/server.key" "${CERT_DIR}/privkey.pem"

echo "   ✓ 服务器证书已生成"
echo ""

echo "3. 设置文件权限"
chmod 400 "${CERT_DIR}/privkey.pem"
chmod 400 "${CERT_DIR}/server.key"
chmod 400 "${CERT_DIR}/ca.key"
chmod 644 "${CERT_DIR}/fullchain.pem"
chmod 644 "${CERT_DIR}/server.crt"
chmod 644 "${CERT_DIR}/ca.crt"

echo "   ✓ 权限已设置"
echo ""

echo "4. 证书信息"
echo "   证书有效期: 365 天"
echo "   TLS 版本支持: TLS 1.2, TLS 1.3"
echo "   支持域名:"
echo "     - localhost, *.localhost"
echo "     - industrial-ai.local, *.industrial-ai.local"
echo "     - backend, frontend"
echo "     - *.industrial-ai.example.com"
echo "     - IP: 127.0.0.1, ::1"
echo ""

echo "=== 证书文件位置 ==="
echo "   ${CERT_DIR}/fullchain.pem  - 证书链 (Nginx 使用)"
echo "   ${CERT_DIR}/privkey.pem    - 私钥 (Nginx 使用)"
echo "   ${CERT_DIR}/ca.crt         - CA 证书 (客户端信任)"
echo ""

echo "=== 安装 CA 证书到系统 (macOS) ==="
echo "   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ${CERT_DIR}/ca.crt"
echo ""

echo "=== 安装 CA 证书到系统 (Linux) ==="
echo "   sudo cp ${CERT_DIR}/ca.crt /usr/local/share/ca-certificates/industrial-ai-dev-ca.crt"
echo "   sudo update-ca-certificates"
echo ""

echo "=== Docker Compose 使用 ==="
echo "   docker-compose -f docker-compose.yml -f docker-compose.dev-ssl.yml up -d"
echo ""

echo "✓ 开发环境证书生成完成！"