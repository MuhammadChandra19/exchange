#!/bin/bash

set -e

SWAGGER_DIR="openapiv3"
MODULES=("market-data-service" "order-management-service")

echo "Organizing OpenAPI v3 files by modules..."

# Create module directories
for module in "${MODULES[@]}"; do
    mkdir -p "${SWAGGER_DIR}/${module}/v1"
done

# Function to determine module from proto path or content
determine_module_from_content() {
    local file="$1"
    
    # Check filename patterns first
    for module in "${MODULES[@]}"; do
        if [[ "$file" == *"$module"* ]]; then
            echo "$module"
            return
        fi
    done
    
    # Check file content for module references
    if [ -f "$file" ]; then
        for module in "${MODULES[@]}"; do
            if grep -q "modules\.$module\." "$file" 2>/dev/null || \
               grep -q "\"$module\"" "$file" 2>/dev/null; then
                echo "$module"
                return
            fi
        done
    fi
    
    echo "common"
}

# Function to extract service info from OpenAPI file
extract_service_info() {
    local file="$1"
    local info=()
    
    if [ -f "$file" ]; then
        # Extract title
        local title=$(grep -o '"title": *"[^"]*"' "$file" | head -1 | sed 's/"title": *"//;s/".*//')
        # Extract version  
        local version=$(grep -o '"version": *"[^"]*"' "$file" | head -1 | sed 's/"version": *"//;s/".*//')
        
        info+=("$title" "$version")
    fi
    
    echo "${info[@]}"
}

# Process generated OpenAPI files
if [ -d "$SWAGGER_DIR" ]; then
    # Find all swagger/openapi files that are not in module directories
    find "$SWAGGER_DIR" -name "*.swagger.json" -o -name "*.json" | \
    grep -v -E "(market-data-service|order-management-service)" | \
    while read -r api_file; do
        if [ -f "$api_file" ]; then
            filename=$(basename "$api_file")
            
            # Determine module for this file
            module=$(determine_module_from_content "$api_file")
            
            # Create module directory if it doesn't exist
            if [ "$module" != "common" ]; then
                mkdir -p "${SWAGGER_DIR}/${module}/v1"
                target_dir="${SWAGGER_DIR}/${module}/v1"
            else
                mkdir -p "${SWAGGER_DIR}/common"
                target_dir="${SWAGGER_DIR}/common"
            fi
            
            echo "Moving $filename to $module module"
            mv "$api_file" "$target_dir/"
            
            # Extract service information for indexing
            service_info=($(extract_service_info "$target_dir/$filename"))
            if [ ${#service_info[@]} -gt 0 ]; then
                echo "  Service: ${service_info[0]} (${service_info[1]})"
            fi
        fi
    done
fi

# Generate module-specific consolidated files and documentation
for module in "${MODULES[@]}" "common"; do
    module_dir="${SWAGGER_DIR}/${module}"
    
    if [ -d "$module_dir" ] && [ "$(ls -A "$module_dir" 2>/dev/null)" ]; then
        echo "Processing $module module..."
        
        # Create module info file
        cat > "$module_dir/module-info.json" << EOF
{
  "module": "$module",
  "version": "v1",
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "description": "OpenAPI v3 specifications for $module",
  "services": []
}
EOF

        # List all API files in this module
        if [ -d "$module_dir/v1" ]; then
            find "$module_dir/v1" -name "*.json" | while read -r file; do
                filename=$(basename "$file")
                service_info=($(extract_service_info "$file"))
                echo "    - $filename (${service_info[0]:-Unknown Service})"
            done
        fi
        
        echo "  Module $module processed successfully"
    fi
done

# Create main documentation index
cat > "${SWAGGER_DIR}/index.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exchange API Documentation</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0; padding: 40px; background: #f8f9fa; 
        }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { border-bottom: 2px solid #e9ecef; padding-bottom: 20px; margin-bottom: 30px; }
        .header h1 { margin: 0; color: #343a40; font-size: 2.5em; }
        .header p { margin: 10px 0 0 0; color: #6c757d; font-size: 1.1em; }
        .module { 
            margin: 30px 0; padding: 25px; border: 1px solid #dee2e6; 
            border-radius: 6px; background: #ffffff; 
        }
        .module h2 { 
            color: #495057; margin: 0 0 15px 0; font-size: 1.5em;
            border-left: 4px solid #007bff; padding-left: 15px;
        }
        .module p { color: #6c757d; margin-bottom: 20px; }
        .api-files { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 15px; }
        .api-file {
            padding: 15px; background: #f8f9fa; border-radius: 4px; border: 1px solid #e9ecef;
            transition: all 0.2s ease; text-decoration: none; color: inherit;
        }
        .api-file:hover { 
            background: #e3f2fd; border-color: #2196f3; transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(33, 150, 243, 0.15);
        }
        .api-file-name { font-weight: 600; color: #1976d2; margin-bottom: 8px; }
        .api-file-desc { font-size: 0.9em; color: #757575; }
        .stats { 
            display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
            gap: 20px; margin: 30px 0; 
        }
        .stat-card {
            padding: 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; border-radius: 6px; text-align: center;
        }
        .stat-number { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .stat-label { opacity: 0.9; }
        .footer { 
            text-align: center; margin-top: 40px; padding-top: 20px; 
            border-top: 1px solid #e9ecef; color: #6c757d; font-size: 0.9em; 
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Exchange API Documentation</h1>
            <p>Comprehensive OpenAPI v3 specifications for all exchange services</p>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number" id="moduleCount">0</div>
                <div class="stat-label">Modules</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="serviceCount">0</div>
                <div class="stat-label">Services</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="apiCount">0</div>
                <div class="stat-label">API Files</div>
            </div>
        </div>
EOF

# Add module sections dynamically
module_count=0
service_count=0
api_count=0

for module in "${MODULES[@]}" "common"; do
    module_dir="${SWAGGER_DIR}/${module}"
    if [ -d "$module_dir" ] && [ "$(ls -A "$module_dir" 2>/dev/null)" ]; then
        module_count=$((module_count + 1))
        
        # Convert module name to display format
        display_name=$(echo "$module" | sed 's/-/ /g' | sed 's/\b\w/\U&/g')
        
        echo "        <div class=\"module\">" >> "${SWAGGER_DIR}/index.html"
        echo "            <h2>$display_name</h2>" >> "${SWAGGER_DIR}/index.html"
        echo "            <p>API documentation for $display_name module</p>" >> "${SWAGGER_DIR}/index.html"
        echo "            <div class=\"api-files\">" >> "${SWAGGER_DIR}/index.html"
        
        # Find API files in this module
        find "$module_dir" -name "*.json" | sort | while read -r file; do
            filename=$(basename "$file")
            relative_path="${file#${SWAGGER_DIR}/}"
            
            # Extract service info for better display
            service_info=($(extract_service_info "$file"))
            service_title="${service_info[0]:-$filename}"
            service_version="${service_info[1]:-v1}"
            
            echo "                <a href=\"$relative_path\" class=\"api-file\" target=\"_blank\">" >> "${SWAGGER_DIR}/index.html"
            echo "                    <div class=\"api-file-name\">$service_title</div>" >> "${SWAGGER_DIR}/index.html"
            echo "                    <div class=\"api-file-desc\">Version: $service_version | File: $filename</div>" >> "${SWAGGER_DIR}/index.html"
            echo "                </a>" >> "${SWAGGER_DIR}/index.html"
            
            api_count=$((api_count + 1))
        done
        
        echo "            </div>" >> "${SWAGGER_DIR}/index.html"
        echo "        </div>" >> "${SWAGGER_DIR}/index.html"
        
        service_count=$((service_count + $(find "$module_dir" -name "*.json" | wc -l)))
    fi
done

cat >> "${SWAGGER_DIR}/index.html" << EOF
        
        <div class="footer">
            <p>Generated on $(date) | Exchange Proto Documentation</p>
            <p>Use tools like <a href="https://editor.swagger.io/" target="_blank">Swagger Editor</a> or <a href="https://github.com/Redocly/redoc" target="_blank">ReDoc</a> to view these API specifications</p>
        </div>
    </div>
    
    <script>
        // Update statistics
        document.getElementById('moduleCount').textContent = '$module_count';
        document.getElementById('serviceCount').textContent = '$service_count';
        document.getElementById('apiCount').textContent = '$api_count';
    </script>
</body>
</html>
EOF

echo ""
echo "✅ OpenAPI v3 organization complete!"
echo "📊 Statistics:"
echo "   - Modules: $module_count"
echo "   - Services: $service_count" 
echo "   - API Files: $api_count"
echo ""
echo "🌐 View documentation: generated/openapiv3/index.html"
echo ""
echo "💡 Tip: Use Swagger Editor (https://editor.swagger.io/) to view individual API specs"