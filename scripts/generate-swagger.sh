#!/bin/bash

# Generate Swagger documentation

echo "🔧 Generating Swagger documentation..."

# Check if swag is installed
if ! command -v $(go env GOPATH)/bin/swag &> /dev/null; then
    echo "⚠️  Installing swag CLI tool..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate documentation and move to API directory
$(go env GOPATH)/bin/swag init -g cmd/server/main.go

# Move swagger files to the API directory
mv docs/swagger.json api/v1/
mv docs/swagger.yaml api/v1/

echo "✅ Swagger documentation generated successfully!"
echo "📄 Files generated:"
echo "  - docs/docs.go"
echo "  - api/v1/swagger.json" 
echo "  - api/v1/swagger.yaml"
echo ""
echo "🌐 Start the server and visit: http://localhost:8080/swagger/index.html"
