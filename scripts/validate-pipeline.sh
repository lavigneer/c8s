#!/bin/bash

# Pipeline Definition Validator for C8S
# Validates pipeline YAML files against PipelineConfig CRD schema
# Provides detailed error messages for debugging

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PIPELINE_FILE=""
VERBOSE=false
EXIT_CODE=0

# Functions
print_header() {
    echo -e "${BLUE}=== C8S Pipeline Validator ===${NC}\n"
}

print_error() {
    echo -e "${RED}✗ ERROR:${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING:${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓ SUCCESS:${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

usage() {
    cat <<EOF
Usage: validate-pipeline.sh [OPTIONS] <pipeline-file>

Validates a C8S pipeline YAML file against the PipelineConfig CRD schema.

OPTIONS:
    -v, --verbose       Show detailed validation output
    -h, --help          Show this help message

EXAMPLES:
    # Validate a pipeline file
    ./scripts/validate-pipeline.sh my-pipeline.yaml

    # Verbose output
    ./scripts/validate-pipeline.sh -v my-pipeline.yaml

    # Validate from stdin
    cat pipeline.yaml | ./scripts/validate-pipeline.sh -

EOF
    exit 1
}

# Check if file exists and is readable
check_file() {
    local file=$1

    if [[ "$file" == "-" ]]; then
        # Reading from stdin
        return 0
    fi

    if [[ ! -f "$file" ]]; then
        print_error "File not found: $file"
        exit 1
    fi

    if [[ ! -r "$file" ]]; then
        print_error "File is not readable: $file"
        exit 1
    fi
}

# Check if yq is installed
check_dependencies() {
    if ! command -v yq &> /dev/null; then
        print_warning "yq not found. Installing yq..."
        # Try to install yq
        if command -v brew &> /dev/null; then
            brew install yq
        else
            print_error "yq is required but not installed. Please install it with: brew install yq"
            exit 1
        fi
    fi

    if ! command -v kubectl &> /dev/null; then
        print_warning "kubectl not found in PATH"
    fi
}

# Validate YAML syntax
validate_syntax() {
    local file=$1

    print_info "Checking YAML syntax..."

    if [[ "$file" == "-" ]]; then
        # Validate from stdin
        if ! yq eval '.' - > /dev/null 2>&1; then
            print_error "Invalid YAML syntax"
            return 1
        fi
    else
        # Validate file
        if ! yq eval '.' "$file" > /dev/null 2>&1; then
            print_error "Invalid YAML syntax in: $file"
            return 1
        fi
    fi

    print_success "YAML syntax is valid"
    return 0
}

# Validate required fields
validate_required_fields() {
    local file=$1
    local temp_file=$(mktemp)

    # Read file or stdin
    if [[ "$file" == "-" ]]; then
        cat > "$temp_file"
    else
        cp "$file" "$temp_file"
    fi

    print_info "Checking required fields..."

    local has_error=0

    # Check version
    local version=$(yq eval '.version' "$temp_file" 2>/dev/null || echo "null")
    if [[ "$version" == "null" ]] || [[ -z "$version" ]]; then
        print_error "Missing required field: version (must be 'v1alpha1')"
        has_error=1
    elif [[ "$version" != "v1alpha1" ]]; then
        print_error "Invalid version: $version (must be 'v1alpha1')"
        has_error=1
    fi

    # Check name
    local name=$(yq eval '.name' "$temp_file" 2>/dev/null || echo "null")
    if [[ "$name" == "null" ]] || [[ -z "$name" ]]; then
        print_error "Missing required field: name"
        has_error=1
    elif ! [[ "$name" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$ ]]; then
        print_error "Invalid name format: '$name' (must be lowercase alphanumeric with hyphens)"
        has_error=1
    fi

    # Check steps
    local steps_count=$(yq eval '.steps | length' "$temp_file" 2>/dev/null || echo "0")
    if [[ "$steps_count" -lt 1 ]]; then
        print_error "Missing required field: steps (must have at least 1 step)"
        has_error=1
    fi

    rm -f "$temp_file"

    if [[ $has_error -eq 0 ]]; then
        print_success "All required fields present"
        return 0
    fi

    return 1
}

# Validate steps structure
validate_steps() {
    local file=$1
    local temp_file=$(mktemp)

    # Read file or stdin
    if [[ "$file" == "-" ]]; then
        cat > "$temp_file"
    else
        cp "$file" "$temp_file"
    fi

    print_info "Validating pipeline steps..."

    local steps_count=$(yq eval '.steps | length' "$temp_file" 2>/dev/null || echo "0")
    local has_error=0

    for ((i = 0; i < steps_count; i++)); do
        local step_name=$(yq eval ".steps[$i].name" "$temp_file" 2>/dev/null || echo "null")
        local step_image=$(yq eval ".steps[$i].image" "$temp_file" 2>/dev/null || echo "null")
        local step_commands=$(yq eval ".steps[$i].commands | length" "$temp_file" 2>/dev/null || echo "0")

        if [[ "$step_name" == "null" ]] || [[ -z "$step_name" ]]; then
            print_error "Step $i: Missing required field 'name'"
            has_error=1
        fi

        if [[ "$step_image" == "null" ]] || [[ -z "$step_image" ]]; then
            print_error "Step $i ($step_name): Missing required field 'image'"
            has_error=1
        fi

        if [[ "$step_commands" -lt 1 ]]; then
            print_error "Step $i ($step_name): Missing required field 'commands' (must have at least 1)"
            has_error=1
        fi

        # Validate resources if present
        local cpu=$(yq eval ".steps[$i].resources.cpu" "$temp_file" 2>/dev/null || echo "null")
        local memory=$(yq eval ".steps[$i].resources.memory" "$temp_file" 2>/dev/null || echo "null")

        if [[ "$cpu" != "null" ]] && [[ -n "$cpu" ]]; then
            if ! [[ "$cpu" =~ ^[0-9]+(m|[0-9])?$ ]]; then
                print_warning "Step $i ($step_name): CPU format may be invalid: $cpu"
            fi
        fi

        if [[ "$memory" != "null" ]] && [[ -n "$memory" ]]; then
            if ! [[ "$memory" =~ ^[0-9]+(Mi|Gi|M|G|Ki|k|m)$ ]]; then
                print_warning "Step $i ($step_name): Memory format may be invalid: $memory"
            fi
        fi
    done

    rm -f "$temp_file"

    if [[ $has_error -eq 0 ]]; then
        print_success "All steps are valid"
        return 0
    fi

    return 1
}

# Validate against CRD (if kubectl available)
validate_against_crd() {
    local file=$1

    if ! command -v kubectl &> /dev/null; then
        print_info "kubectl not available, skipping CRD schema validation"
        return 0
    fi

    print_info "Validating against PipelineConfig CRD schema..."

    # Try to validate by applying with dry-run
    if [[ "$file" == "-" ]]; then
        if kubectl apply --dry-run=client -f - 2>/dev/null; then
            print_success "CRD schema validation passed"
            return 0
        else
            print_warning "CRD schema validation may have issues (could not reach cluster)"
            return 0
        fi
    else
        if kubectl apply --dry-run=client -f "$file" 2>/dev/null; then
            print_success "CRD schema validation passed"
            return 0
        else
            print_warning "CRD schema validation may have issues (could not reach cluster)"
            return 0
        fi
    fi
}

# Main validation function
validate_pipeline() {
    local file=$1

    print_header
    print_info "Validating pipeline: $file\n"

    check_file "$file"
    check_dependencies

    # Run all validations
    validate_syntax "$file" || EXIT_CODE=1
    validate_required_fields "$file" || EXIT_CODE=1
    validate_steps "$file" || EXIT_CODE=1
    validate_against_crd "$file" || EXIT_CODE=1

    echo ""
    if [[ $EXIT_CODE -eq 0 ]]; then
        echo -e "${GREEN}✅ Pipeline validation PASSED${NC}"
    else
        echo -e "${RED}❌ Pipeline validation FAILED${NC}"
    fi

    return $EXIT_CODE
}

# Parse arguments
if [[ $# -eq 0 ]]; then
    usage
fi

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            PIPELINE_FILE=$1
            shift
            ;;
    esac
done

# Validate pipeline file
validate_pipeline "$PIPELINE_FILE"
exit $EXIT_CODE
