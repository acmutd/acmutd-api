#!/bin/bash

# ACM UTD API Setup Script
# Cross-platform setup for macOS and Windows (via Git Bash/WSL)
# Sets up Go dependencies and shared Python virtual environment

set -e

echo "Setting up ACM UTD API development environment..."

# Detect operating system
detect_os() {
    case "$OSTYPE" in
        darwin*)  echo "macos" ;;
        linux*)   echo "linux" ;;
        msys*|mingw*|cygwin*) echo "windows" ;;
        *) echo "unknown" ;;
    esac
}

OS=$(detect_os)
echo "Detected OS: $OS"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Python on different platforms
install_python() {
    echo "Checking Python installation..."

    # Prefer python found in PATH
    PYTHON_PATH=$(which python 2>/dev/null)
    if [ -n "$PYTHON_PATH" ]; then
        PYTHON_VERSION=$($PYTHON_PATH --version 2>&1)
        echo "Detected python in PATH: $PYTHON_PATH ($PYTHON_VERSION)"
        if [[ $PYTHON_VERSION == *"Python 3"* ]]; then
            export PYTHON_CMD="$PYTHON_PATH"
            return
        fi
    fi

    # Allow manual override
    if [ -n "$PYTHON_CMD" ]; then
        echo "Using manually specified Python: $PYTHON_CMD"
        PYTHON_VERSION=$($PYTHON_CMD --version 2>&1)
        if [[ $PYTHON_VERSION != *"Python 3"* ]]; then
            echo "ERROR: PYTHON_CMD does not point to Python 3: $PYTHON_VERSION"
            exit 1
        fi
        return
    fi

    if command_exists python3; then
        export PYTHON_CMD="python3"
        echo "Python3 found: $(python3 --version)"
    elif command_exists python; then
        PYTHON_VERSION=$(python --version 2>&1)
        if [[ $PYTHON_VERSION == *"Python 3"* ]]; then
            export PYTHON_CMD="python"
            echo "Python found: $PYTHON_VERSION"
        else
            echo "ERROR: Python 3 is required but Python 2 was found"
            exit 1
        fi
    elif [ "$OS" = "windows" ]; then
        # Try common Windows install locations
        WIN_PYTHON_PATHS=(
            "/c/Users/$USERNAME/AppData/Local/Programs/Python/Python3*/python.exe"
            "/c/Python3*/python.exe"
            "/c/Program Files/Python3*/python.exe"
        )
        for pathglob in "${WIN_PYTHON_PATHS[@]}"; do
            for pyexe in $(ls $pathglob 2>/dev/null); do
                PYTHON_VERSION=$($pyexe --version 2>&1)
                if [[ $PYTHON_VERSION == *"Python 3"* ]]; then
                    export PYTHON_CMD="$pyexe"
                    echo "Found Python 3 at $pyexe: $PYTHON_VERSION"
                    return
                fi
            done
        done
        echo "ERROR: Python 3 not found in PATH or common install locations."
        echo "You can manually set PYTHON_CMD to your python.exe path before running setup.sh."
        echo "Example: export PYTHON_CMD='/c/Users/YourName/AppData/Local/Programs/Python/Python3x/python.exe'"
        exit 1
    else
        echo "ERROR: Python 3 not found in your Bash/WSL environment."
        case $OS in
            macos)
                echo "  - Install via Homebrew: brew install python3"
                echo "  - Or download from: https://www.python.org/downloads/"
                ;;
            linux)
                echo "  - Ubuntu/Debian: sudo apt-get install python3 python3-pip python3-venv"
                echo "  - CentOS/RHEL: sudo yum install python3 python3-pip"
                ;;
        esac
        exit 1
    fi
}

# Function to install Go
install_go() {
    echo "Checking Go installation..."

    if command_exists go; then
        GO_VERSION=$(go version)
        echo "Go found: $GO_VERSION"

        # Check if Go version is compatible (1.24+ required)
        GO_VERSION_NUM=$(go version | grep -o 'go[0-9.]*' | sed 's/go//')
        REQUIRED_VERSION="1.24"

        if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION_NUM" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
            echo "Go version is compatible"
        else
            echo "WARNING: Go version $GO_VERSION_NUM found, but $REQUIRED_VERSION+ is recommended"
        fi
    else
        echo "ERROR: Go not found. Please install Go manually:"
        case $OS in
            macos)
                echo "  - Install via Homebrew: brew install go"
                echo "  - Or download from: https://golang.org/dl/"
                ;;
            windows)
                echo "  - Download from: https://golang.org/dl/"
                echo "  - Or install via Chocolatey: choco install golang"
                ;;
            linux)
                echo "  - Download from: https://golang.org/dl/"
                echo "  - Or use your package manager"
                ;;
        esac
        exit 1
    fi
}

# Function to set up Go dependencies
setup_go() {
    echo "Setting up Go dependencies..."

    if [ ! -f "go.mod" ]; then
        echo "ERROR: go.mod not found. Make sure you're in the project root directory."
        exit 1
    fi

    echo "Downloading Go modules..."
    go mod download

    echo "Verifying Go modules..."
    go mod verify

    echo "Go dependencies installed successfully!"
}

# Function to create consolidated requirements file
create_consolidated_requirements() {
    echo "Creating consolidated Python requirements..."

    # Find all requirements.txt files in scripts subfolders
    REQUIREMENTS_FILES=$(find scripts -type f -name 'requirements.txt')
    if [ -z "$REQUIREMENTS_FILES" ]; then
        echo "No requirements.txt files found in scripts subfolders."
        touch requirements-consolidated.txt
        return
    fi

    # Concatenate, deduplicate, and write to requirements-consolidated.txt
    (
        echo "# Consolidated requirements from all scripts"
        echo "# Generated by setup script"
        for reqfile in $REQUIREMENTS_FILES; do
            cat "$reqfile"
            echo "" # Ensure newline between files
        done | grep -v '^#' | grep -v '^$' | sort | uniq
    ) > requirements-consolidated.txt

    echo "Consolidated requirements file created from:"
    echo "$REQUIREMENTS_FILES"
}

# Function to set up environment file
setup_env() {
    echo "Setting up environment configuration..."

    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            echo "Copying .env.example to .env..."
            cp .env.example .env
            echo "Environment file created from template"
            echo "Please edit .env with your specific configuration values"
        else
            echo "WARNING: No .env.example file found to copy"
        fi
    else
        echo "Environment file .env already exists"
    fi
}

# Function to set up Python virtual environment
setup_python() {
    echo "Setting up Python virtual environment..."

    # Check if Python supports venv before creating
    if ! $PYTHON_CMD -m venv --help >/dev/null 2>&1; then
        echo "ERROR: $PYTHON_CMD does not support 'venv' or is not a valid Python 3 interpreter."
        echo "Please ensure you have Python 3 installed and available in your PATH."
        case $OS in
            windows)
                echo "If you are on Windows, try running this setup in PowerShell or CMD instead of Bash."
                echo "Or add your Python installation to your Bash PATH."
                ;;
            *)
                echo "Install Python 3 and ensure it is available as 'python3' or 'python' in your shell."
                ;;
        esac
        exit 1
    fi

    # Create virtual environment if it doesn't exist
    if [ ! -d "venv" ]; then
        echo "Creating virtual environment..."
        $PYTHON_CMD -m venv venv
    else
        echo "Virtual environment already exists"
    fi

    # Activate virtual environment
    echo "Activating virtual environment..."
    case $OS in
        windows)
            source venv/Scripts/activate
            ;;
        *)
            source venv/bin/activate
            ;;
    esac

    # Upgrade pip
    echo "Upgrading pip..."
    python -m pip install --upgrade pip

    # Create consolidated requirements
    create_consolidated_requirements

    # Install requirements
    echo "Installing Python packages..."
    pip install -r requirements-consolidated.txt

    echo "Python environment setup complete!"
}

# Function to display usage instructions
show_usage() {
    echo ""
    echo "Setup complete! Here's how to use your development environment:"
    echo ""
    echo "IMPORTANT: Make sure to configure your .env file with the correct values before running the application."
    echo ""
    echo "Project Structure:"
    echo "  - Go API: cmd/api/main.go"
    echo "  - Scraper: cmd/scraper/main.go"
    echo "  - Python Scripts: scripts/*/main.py"
    echo ""
    echo "Running the application:"
    echo "  - Go API: go run cmd/api/main.go"
    echo "  - Go Scraper: go run cmd/scraper/main.go"
    echo ""
    echo "Python Scripts (activate environment first):"
    case $OS in
        windows)
            echo "  - Windows CMD: activate-env.bat && python scripts/coursebook/main.py"
            echo "  - Git Bash/WSL: source activate-env.sh && python scripts/coursebook/main.py"
            ;;
        *)
            echo "  - Activate: source activate-env.sh"
            echo "  - Run: python scripts/coursebook/main.py"
            echo "  - Run: python scripts/grades/main.py"
            echo "  - Run: python scripts/professors/main.py"
            ;;
    esac
    echo ""
    echo "Development commands:"
    echo "  - Go build: go build -o bin/ ./cmd/..."
    echo "  - Go test: go test ./..."
    echo "  - Go format: go fmt ./..."
    echo ""
}

# Main execution
main() {
    echo "Starting setup process..."
    echo ""

    # Check prerequisites
    install_python
    install_go

    # Set up dependencies
    setup_go
    setup_env
    setup_python


    # Clean up temporary files
    rm -f requirements-consolidated.txt

    # Show usage instructions
    show_usage

    echo "Setup completed successfully!"
}

# Run main function
main "$@"
