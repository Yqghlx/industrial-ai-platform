#!/bin/bash

# 🔐 Industrial AI Platform 密钥生成脚本
# 生成时间: 2026-05-19
# 重要: 不要将此脚本输出提交到 Git！

echo "========================================="
echo "🔐 Industrial AI Platform 密钥生成"
echo "========================================="
echo ""

# 生成密钥
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 24)
DB_PASSWORD=$(openssl rand -base64 24)
ADMIN_PASSWORD="IndAI_$(openssl rand -base64 16 | tr -d '/+=' | cut -c1-20)_2026"

echo "✅ 密钥已生成！"
echo ""
echo "📋 以下是生成的密钥（请妥善保存）："
echo "========================================="
echo ""
echo "JWT_SECRET:"
echo "$JWT_SECRET"
echo ""
echo "ENCRYPTION_KEY:"
echo "$ENCRYPTION_KEY"
echo ""
echo "REDIS_PASSWORD:"
echo "$REDIS_PASSWORD"
echo ""
echo "DB_PASSWORD:"
echo "$DB_PASSWORD"
echo ""
echo "ADMIN_PASSWORD:"
echo "$ADMIN_PASSWORD"
echo ""
echo "DATABASE_URL:"
echo "postgres://postgres:${DB_PASSWORD}@postgres:5432/industrial_ai?sslmode=require"
echo ""
echo "========================================="

# 保存到临时文件
SECRETS_FILE=".secrets.tmp"
echo "# Industrial AI Platform Secrets (DO NOT COMMIT)" > "$SECRETS_FILE"
echo "# Generated: $(date)" >> "$SECRETS_FILE"
echo "" >> "$SECRETS_FILE"
echo "JWT_SECRET=$JWT_SECRET" >> "$SECRETS_FILE"
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> "$SECRETS_FILE"
echo "REDIS_PASSWORD=$REDIS_PASSWORD" >> "$SECRETS_FILE"
echo "DB_PASSWORD=$DB_PASSWORD" >> "$SECRETS_FILE"
echo "ADMIN_PASSWORD=$ADMIN_PASSWORD" >> "$SECRETS_FILE"
echo "" >> "$SECRETS_FILE"
echo "# For GitHub Secrets:" >> "$SECRETS_FILE"
echo "# Copy each value above to GitHub Settings > Secrets > Actions" >> "$SECRETS_FILE"

echo ""
echo "📁 密钥已保存到: $SECRETS_FILE"
echo "⚠️  此文件已在 .gitignore 中，不会提交到 Git"
echo ""

# 询问是否配置 .env.production
echo "是否自动配置 .env.production？(y/n)"
read -r answer

if [ "$answer" = "y" ] || [ "$answer" = "Y" ]; then
    if [ -f ".env.production" ]; then
        # 更新 .env.production
        sed -i.bak "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env.production
        sed -i.bak "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" .env.production
        sed -i.bak "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASSWORD/" .env.production
        sed -i.bak "s/^DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" .env.production
        
        echo "✅ .env.production 已更新"
        echo "📁 备份文件: .env.production.bak"
    else
        echo "❌ .env.production 文件不存在"
        echo "请先创建 .env.production 文件"
    fi
fi

echo ""
echo "========================================="
echo "📌 下一步："
echo "1. 将密钥保存到密码管理器"
echo "2. 配置 GitHub Secrets:"
echo "   https://github.com/Yqghlx/industrial-ai-platform/settings/secrets/actions"
echo ""
echo "   需要添加的 Secrets:"
echo "   - JWT_SECRET"
echo "   - ENCRYPTION_KEY"
echo "   - REDIS_PASSWORD"
echo "   - DB_PASSWORD"
echo "   - DEPLOY_TOKEN (需创建 GitHub Personal Access Token)"
echo ""
echo "3. 创建 DEPLOY_TOKEN:"
echo "   https://github.com/settings/tokens"
echo "   权限: repo, write:packages, read:packages"
echo ""
echo "4. 配置完成后运行部署 workflow"
echo "========================================="

# 显示 Kubernetes 配置命令
echo ""
echo "📦 Kubernetes Secrets 配置命令:"
echo "----------------------------------------"
echo "kubectl create secret generic industrial-ai-secrets \\"
echo "  --namespace=industrial-ai \\"
echo "  --from-literal=jwt-secret=\"$JWT_SECRET\" \\"
echo "  --from-literal=encryption-key=\"$ENCRYPTION_KEY\" \\"
echo "  --from-literal=redis-password=\"$REDIS_PASSWORD\" \\"
echo "  --from-literal=db-password=\"$DB_PASSWORD\""
echo "----------------------------------------"
echo ""

echo "✅ 完成！请按照上述步骤继续配置。"