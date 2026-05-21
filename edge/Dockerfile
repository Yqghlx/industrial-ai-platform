# Edge Simulator Dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder

WORKDIR /app

# Copy project files
COPY *.csproj *.json ./

# Restore dependencies
RUN dotnet restore

# Copy source code
COPY . .

# Build
RUN dotnet publish -c Release -o out

# Final stage
FROM mcr.microsoft.com/dotnet/runtime:8.0

WORKDIR /app

# Copy published files
COPY --from=builder /app/out .

# Entry point
ENTRYPOINT ["dotnet", "EdgeSimulator.dll"]