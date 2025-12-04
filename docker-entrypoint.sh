#!/bin/sh
set -e

# ===========================================
# Docker Entrypoint Script for Daily Reminder Bot
# Handles configuration generation and key management
# ===========================================

CONFIG_DIR="/app/configs"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
PRIVATE_KEY_FILE="${CONFIG_DIR}/ed25519-private.pem"
PUBLIC_KEY_FILE="${CONFIG_DIR}/ed25519-public.pem"

echo "=========================================="
echo "Daily Reminder Bot - Container Startup"
echo "=========================================="

# Create config directory if it doesn't exist
mkdir -p "${CONFIG_DIR}"

# ===========================================
# Handle Ed25519 Key Pair
# ===========================================
handle_keys() {
    echo "[Keys] Checking Ed25519 key configuration..."
    
    # If QWEATHER_PRIVATE_KEY is provided (base64 or raw PEM), use it
    if [ -n "${QWEATHER_PRIVATE_KEY}" ]; then
        echo "[Keys] Using provided private key from environment variable"
        
        # Check if it looks like base64 (no newlines, no "BEGIN")
        if echo "${QWEATHER_PRIVATE_KEY}" | grep -q "BEGIN"; then
            # It's raw PEM format
            echo "${QWEATHER_PRIVATE_KEY}" > "${PRIVATE_KEY_FILE}"
        else
            # It's base64 encoded
            echo "${QWEATHER_PRIVATE_KEY}" | base64 -d > "${PRIVATE_KEY_FILE}"
        fi
        
        chmod 600 "${PRIVATE_KEY_FILE}"
        echo "[Keys] Private key saved to ${PRIVATE_KEY_FILE}"
    
    # If private key file doesn't exist, generate new key pair
    elif [ ! -f "${PRIVATE_KEY_FILE}" ]; then
        echo "[Keys] No private key found. Generating new Ed25519 key pair..."
        
        # Generate Ed25519 private key in PKCS8 format
        openssl genpkey -algorithm Ed25519 -out "${PRIVATE_KEY_FILE}"
        
        # Extract public key
        openssl pkey -in "${PRIVATE_KEY_FILE}" -pubout -out "${PUBLIC_KEY_FILE}"
        
        chmod 600 "${PRIVATE_KEY_FILE}"
        chmod 644 "${PUBLIC_KEY_FILE}"
        
        echo "[Keys] ============================================"
        echo "[Keys] NEW KEY PAIR GENERATED!"
        echo "[Keys] ============================================"
        echo "[Keys] Private key: ${PRIVATE_KEY_FILE}"
        echo "[Keys] Public key:  ${PUBLIC_KEY_FILE}"
        echo ""
        echo "[Keys] IMPORTANT: You need to upload the public key to QWeather console"
        echo "[Keys] and set QWEATHER_KEY_ID and QWEATHER_PROJECT_ID accordingly."
        echo ""
        echo "[Keys] Public Key Content (upload this to QWeather):"
        echo "--------------------------------------------"
        cat "${PUBLIC_KEY_FILE}"
        echo "--------------------------------------------"
        echo ""
        echo "[Keys] To export private key for backup (base64 encoded):"
        echo "cat ${PRIVATE_KEY_FILE} | base64"
        echo "[Keys] ============================================"
    else
        echo "[Keys] Using existing private key at ${PRIVATE_KEY_FILE}"
    fi
}

# ===========================================
# Generate Configuration File
# ===========================================
generate_config() {
    echo "[Config] Generating configuration file..."
    
    cat > "${CONFIG_FILE}" << EOF
# Auto-generated configuration from environment variables
# Generated at: $(date -u '+%Y-%m-%d %H:%M:%S UTC')

telegram:
  token: "${TELEGRAM_TOKEN}"
  api_endpoint: "${TELEGRAM_API_ENDPOINT}"

qweather:
  auth_mode: "${QWEATHER_AUTH_MODE}"
  private_key_path: "${PRIVATE_KEY_FILE}"
  key_id: "${QWEATHER_KEY_ID}"
  project_id: "${QWEATHER_PROJECT_ID}"
  api_key: "${QWEATHER_API_KEY}"
  base_url: "${QWEATHER_BASE_URL}"

openai:
  enabled: ${OPENAI_ENABLED}
  api_key: "${OPENAI_API_KEY}"
  base_url: "${OPENAI_BASE_URL}"
  model: "${OPENAI_MODEL}"
  max_tokens: ${OPENAI_MAX_TOKENS}
  temperature: ${OPENAI_TEMPERATURE}
  timeout: ${OPENAI_TIMEOUT}
  max_retries: ${OPENAI_MAX_RETRIES}

holiday:
  api_url: "${HOLIDAY_API_URL}"
  cache_ttl: ${HOLIDAY_CACHE_TTL}

database:
  type: "${DATABASE_TYPE}"
  path: "${DATABASE_PATH}"
  host: "${DATABASE_HOST}"
  port: ${DATABASE_PORT}
  user: "${DATABASE_USER}"
  password: "${DATABASE_PASSWORD}"
  dbname: "${DATABASE_NAME}"
  charset: "${DATABASE_CHARSET}"

scheduler:
  timezone: "${SCHEDULER_TIMEZONE}"

logger:
  level: "${LOGGER_LEVEL}"
  format: "${LOGGER_FORMAT}"
EOF

    echo "[Config] Configuration saved to ${CONFIG_FILE}"
}

# ===========================================
# Validate Configuration
# ===========================================
validate_config() {
    echo "[Validate] Checking required configuration..."
    
    local has_error=0
    
    # Check Telegram token
    if [ -z "${TELEGRAM_TOKEN}" ]; then
        echo "[ERROR] TELEGRAM_TOKEN is required"
        has_error=1
    fi
    
    # Check QWeather configuration based on auth mode
    if [ "${QWEATHER_AUTH_MODE}" = "jwt" ]; then
        if [ -z "${QWEATHER_KEY_ID}" ]; then
            echo "[ERROR] QWEATHER_KEY_ID is required for JWT auth mode"
            has_error=1
        fi
        if [ -z "${QWEATHER_PROJECT_ID}" ]; then
            echo "[ERROR] QWEATHER_PROJECT_ID is required for JWT auth mode"
            has_error=1
        fi
        if [ -z "${QWEATHER_BASE_URL}" ]; then
            echo "[ERROR] QWEATHER_BASE_URL is required"
            has_error=1
        fi
    elif [ "${QWEATHER_AUTH_MODE}" = "api_key" ]; then
        if [ -z "${QWEATHER_API_KEY}" ]; then
            echo "[ERROR] QWEATHER_API_KEY is required for api_key auth mode"
            has_error=1
        fi
        if [ -z "${QWEATHER_BASE_URL}" ]; then
            echo "[ERROR] QWEATHER_BASE_URL is required"
            has_error=1
        fi
    fi
    
    if [ $has_error -eq 1 ]; then
        echo ""
        echo "[Validate] Configuration validation failed. Please check the errors above."
        exit 1
    fi
    
    echo "[Validate] Configuration validation passed"
}

# ===========================================
# Main Execution
# ===========================================

# Handle key pair generation/import
handle_keys

# Generate configuration from environment variables
generate_config

# Validate configuration
validate_config

echo ""
echo "[Start] Starting Daily Reminder Bot..."
echo "=========================================="

# Execute the main application
exec /app/daily-reminder-bot -config "${CONFIG_FILE}"
