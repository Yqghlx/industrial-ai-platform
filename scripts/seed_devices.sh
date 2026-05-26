#!/bin/bash
# Seed test devices with UUIDs
# SECURITY: Read credentials from environment variables

# Check required environment variables
if [ -z "$ADMIN_PASSWORD" ]; then
  echo "❌ ADMIN_PASSWORD environment variable must be set"
  exit 1
fi

# Login and get token
AUTH_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"${ADMIN_PASSWORD}\"}")

TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Login failed"
  exit 1
fi

echo "Token obtained"

# Function to generate simple UUID-like ID
generate_id() {
  echo "$(date +%s%N | md5sum | head -c 8)-$(date +%N | md5sum | head -c 4)"
}

# Create devices with explicit IDs
curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"cnc-001","name":"CNC加工中心-01","type":"CNC","status":"online","location":"车间A-01"}' && echo " CNC-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"cnc-002","name":"CNC加工中心-02","type":"CNC","status":"maintenance","location":"车间A-02"}' && echo " CNC-002"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"inj-001","name":"注塑机-A1","type":"InjectionMolder","status":"online","location":"车间B-01"}' && echo " INJ-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"rob-001","name":"装配机器人-R1","type":"AssemblyRobot","status":"online","location":"车间C-01"}' && echo " ROB-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"rob-002","name":"装配机器人-R2","type":"AssemblyRobot","status":"error","location":"车间C-02"}' && echo " ROB-002"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"cnv-001","name":"输送带-L1","type":"Conveyor","status":"online","location":"车间D-01"}' && echo " CNV-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"sen-001","name":"温度传感器-S1","type":"Sensor","status":"online","location":"车间A-01"}' && echo " SEN-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"sen-002","name":"压力传感器-S2","type":"sensor","status":"online","location":"车间B-01"}' && echo " SEN-002"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"gau-001","name":"流量仪表-G1","type":"gauge","status":"online","location":"车间A-03"}' && echo " GAU-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"plc-001","name":"PLC控制器-P1","type":"PLC","status":"online","location":"控制室-01"}' && echo " PLC-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"mot-001","name":"伺服电机-M1","type":"motor","status":"online","location":"车间A-01"}' && echo " MOT-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"pmp-001","name":"液压泵-P2","type":"pump","status":"offline","location":"车间B-02"}' && echo " PMP-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"vlv-001","name":"电磁阀-V1","type":"valve","status":"online","location":"车间C-01"}' && echo " VLV-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"hea-001","name":"加热器-H1","type":"heater","status":"online","location":"车间D-02"}' && echo " HEA-001"

curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"clr-001","name":"冷却器-C1","type":"cooler","status":"online","location":"车间D-03"}' && echo " CLR-001"

echo ""
echo "Done. Checking device count..."
curl -s http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Total: {d.get(\"total\", \"N/A\")} devices')"