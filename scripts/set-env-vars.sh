SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)
export PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo PROJECT_DIR=$PROJECT_DIR
