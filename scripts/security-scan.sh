#!/bin/bash

# Security scanning script for WTF project
# This script runs various security checks including vulnerability scanning,
# dependency analysis, and security linting.

set -e

echo "🔒 Starting security scan for WTF project..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking required tools..."
    
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v golangci-lint &> /dev/null; then
        print_warning "golangci-lint not found, attempting to install..."
        if ! curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2; then
            missing_tools+=("golangci-lint")
        fi
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_error "Please install the missing tools and try again."
        exit 1
    fi
    
    print_success "All required tools are available"
}

# Run Go vulnerability scanning
run_govulncheck() {
    print_status "Running Go vulnerability check..."
    
    # Install govulncheck if not present
    if ! command -v govulncheck &> /dev/null; then
        print_status "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
    
    if govulncheck ./...; then
        print_success "No known vulnerabilities found"
    else
        print_error "Vulnerabilities detected! Please review and update dependencies."
        return 1
    fi
}

# Run dependency analysis
analyze_dependencies() {
    print_status "Analyzing dependencies..."
    
    # Check for outdated dependencies
    print_status "Checking for outdated dependencies..."
    go list -u -m all | grep -v "^github.com/Vedant9500/WTF" | while read -r line; do
        if [[ $line == *"["* ]]; then
            print_warning "Outdated dependency: $line"
        fi
    done
    
    # Verify module integrity
    print_status "Verifying module integrity..."
    if go mod verify; then
        print_success "Module integrity verified"
    else
        print_error "Module integrity check failed"
        return 1
    fi
    
    # Check for unused dependencies
    print_status "Checking for unused dependencies..."
    if command -v go-mod-outdated &> /dev/null; then
        go list -u -m -json all | go-mod-outdated -update -direct
    else
        print_warning "go-mod-outdated not installed, skipping unused dependency check"
    fi
}

# Run security linting
run_security_linting() {
    print_status "Running security linting with golangci-lint..."
    
    # Run golangci-lint with focus on security
    if golangci-lint run --enable=gosec,gocritic,staticcheck --timeout=5m; then
        print_success "Security linting passed"
    else
        print_error "Security linting found issues"
        return 1
    fi
}

# Check for hardcoded secrets
check_secrets() {
    print_status "Checking for hardcoded secrets..."
    
    local secret_patterns=(
        "password\s*=\s*['\"][^'\"]*['\"]"
        "api[_-]?key\s*=\s*['\"][^'\"]*['\"]"
        "secret\s*=\s*['\"][^'\"]*['\"]"
        "token\s*=\s*['\"][^'\"]*['\"]"
        "-----BEGIN.*PRIVATE KEY-----"
        "-----BEGIN.*CERTIFICATE-----"
    )
    
    local found_secrets=false
    
    for pattern in "${secret_patterns[@]}"; do
        if grep -r -i -E "$pattern" --include="*.go" --include="*.yml" --include="*.yaml" --include="*.json" . 2>/dev/null; then
            print_warning "Potential secret found matching pattern: $pattern"
            found_secrets=true
        fi
    done
    
    if [ "$found_secrets" = false ]; then
        print_success "No hardcoded secrets detected"
    else
        print_warning "Potential secrets found - please review manually"
    fi
}

# Check file permissions
check_file_permissions() {
    print_status "Checking file permissions..."
    
    # Check for world-writable files
    local world_writable=$(find . -type f -perm -002 2>/dev/null | grep -v ".git" || true)
    if [ -n "$world_writable" ]; then
        print_warning "World-writable files found:"
        echo "$world_writable"
    else
        print_success "No world-writable files found"
    fi
    
    # Check for executable files that shouldn't be
    local suspicious_executables=$(find . -name "*.go" -executable 2>/dev/null || true)
    if [ -n "$suspicious_executables" ]; then
        print_warning "Go source files with executable permissions:"
        echo "$suspicious_executables"
    fi
}

# Generate security report
generate_report() {
    print_status "Generating security report..."
    
    local report_file="security-report-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "WTF Security Scan Report"
        echo "========================"
        echo "Generated: $(date)"
        echo "Go version: $(go version)"
        echo ""
        
        echo "Dependencies:"
        go list -m all
        echo ""
        
        echo "Vulnerability Check:"
        if command -v govulncheck &> /dev/null; then
            govulncheck ./... 2>&1 || echo "Vulnerabilities found - see above"
        else
            echo "govulncheck not available"
        fi
        echo ""
        
        echo "Security Linting Results:"
        golangci-lint run --enable=gosec --out-format=tab 2>&1 || echo "Security issues found - see above"
        
    } > "$report_file"
    
    print_success "Security report generated: $report_file"
}

# Main execution
main() {
    local exit_code=0
    
    check_dependencies
    
    print_status "Starting comprehensive security scan..."
    
    # Run all security checks
    run_govulncheck || exit_code=1
    analyze_dependencies || exit_code=1
    run_security_linting || exit_code=1
    check_secrets
    check_file_permissions
    
    # Generate report regardless of exit code
    generate_report
    
    if [ $exit_code -eq 0 ]; then
        print_success "🎉 Security scan completed successfully!"
        print_status "No critical security issues found."
    else
        print_error "❌ Security scan completed with issues!"
        print_error "Please review and address the issues found above."
    fi
    
    exit $exit_code
}

# Run main function
main "$@"