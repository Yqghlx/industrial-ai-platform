
# Hermes + industrial-ai-platform 迁移指南

# 虚拟机 → macOS 主机

## 仓库地址 (已准备好!)

| 仓库 | 用途 | 地址 |
|------|------|------|
| **myhermes** | Hermes配置+记忆+Skills | https://gitee.com/yangqinggang-it/myhermes.git |
| **hermes-agent** | Hermes核心代码 | https://github.com/NousResearch/hermes-agent.git |
| **industrial-ai-platform** | 项目代码 | https://gitee.com/yangqinggang-it/industrial-ai-platform.git |

---

## 主机上执行 (仅3步!)

### Step 1: 安装 Hermes 核心

```bash
cd ~
mkdir -p .hermes
cd .hermes
git clone https://github.com/NousResearch/hermes-agent.git
cd hermes-agent
python3 -m venv venv
source venv/bin/activate
pip install -e .
```

### Step 2: 恢复配置 (从 Gitee clone)
```bash
cd ~
git clone https://gitee.com/yangqinggang-it/myhermes.git
cd myhermes
cp .env ~/.hermes/
cp config.yaml ~/.hermes/
cp auth.json ~/.hermes/
cp SOUL.md ~/.hermes/
cp -r skills ~/.hermes/
cp -r memories ~/.hermes/
cp -r cron ~/.hermes/
cd ..
rm -rf myhermes
```

### Step 3: Clone 项目
```bash
mkdir -p ~/Projects
cd ~/Projects
git clone https://gitee.com/yangqinggang-it/industrial-ai-platform.git
cd industrial-ai-platform
git log --oneline -7
```

---

## 验证
```bash
hermes --version
cd ~/Projects/industrial-ai-platform/backend && go build ./...
```

---

## 备注
- myhermes 仓库已包含所有配置、记忆、Skills
- API 密钥已保存，无需手动配置
- 迁移后我会记得之前的对话历史
