#!/bin/bash

# Generate Swagger documentation

echo "🔧 Generating Swagger documentation..."

# Check if swag is installed
if ! command -v $(go env GOPATH)/bin/swag &> /dev/null; then
    echo "⚠️  Installing swag CLI tool..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate documentation
$(go env GOPATH)/bin/swag init -g cmd/server/main.go

echo "✅ Swagger documentation generated successfully!"
echo "📄 Files generated:"
echo "  - docs/docs.go"
echo "  - docs/swagger.json" 
echo "  - docs/swagger.yaml"
echo ""
echo "🌐 Start the server and visit: http://localhost:8080/swagger/index.html"
