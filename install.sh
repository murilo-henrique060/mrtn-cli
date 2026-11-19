#!/usr/bin/env bash
#
# mrtn installer script (Linux & macOS)
# https://github.com/murilo-henrique060/mrtn-cli
#
set -euo pipefail

REPO="murilo-henrique060/mrtn-cli"
DEFAULT_VERSION="v1.1.0"
VERSION="${MRTN_VERSION:-$DEFAULT_VERSION}"

echo "============================================================"
echo "          Instalador do mrtn (v${VERSION#v})               "
echo "============================================================"

# 1. Detect OS
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS_RAW" in
    linux*)  OS="linux" ;;
    darwin*) OS="darwin" ;;
    *)
        echo "Erro: Sistema operacional '$OS_RAW' não é suportado por este instalador."
        exit 1
        ;;
esac

# 2. Detect Architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64)   ARCH="amd64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    *)
        echo "Erro: Arquitetura '$ARCH_RAW' não é suportada."
        exit 1
        ;;
esac

BINARY_NAME="mrtn-${VERSION}-${OS}-${ARCH}"
echo "Detectado: ${OS} (${ARCH})"

# 3. Choose install directory
if [ -z "${INSTALL_DIR:-}" ]; then
    if [ -w "/usr/local/bin" ] && [ "${MRTN_USER_INSTALL:-0}" != "1" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="${HOME}/.local/bin"
    fi
fi

mkdir -p "$INSTALL_DIR"
TARGET="${INSTALL_DIR}/mrtn"

# 4. Obtain binary (local or download)
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

LOCAL_BINARY="./dist/${VERSION}/${BINARY_NAME}"
if [ -f "$LOCAL_BINARY" ]; then
    echo "Usando binário local: $LOCAL_BINARY"
    cp "$LOCAL_BINARY" "$TARGET"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    echo "Baixando ${BINARY_NAME}..."
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/mrtn"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "${TMP_DIR}/mrtn" "$DOWNLOAD_URL"
    else
        echo "Erro: curl ou wget é necessário para baixar o binário."
        exit 1
    fi
    mv "${TMP_DIR}/mrtn" "$TARGET"
fi

chmod +x "$TARGET"

# 5. Check PATH configuration
PATH_CONFIGURED=0
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) PATH_CONFIGURED=1 ;;
esac

echo ""
echo "mrtn instalado com sucesso em: ${TARGET}"

if [ "$PATH_CONFIGURED" -eq 0 ]; then
    echo ""
    echo "ATENÇÃO: O diretório '${INSTALL_DIR}' não está no seu PATH."
    echo "Para poder executar 'mrtn' de qualquer lugar, adicione-o ao seu shell:"
    
    SHELL_NAME="$(basename "${SHELL:-bash}")"
    RC_FILE=""
    if [ "$SHELL_NAME" = "zsh" ]; then
        RC_FILE="${HOME}/.zshrc"
    elif [ "$SHELL_NAME" = "bash" ]; then
        if [ -f "${HOME}/.bashrc" ]; then
            RC_FILE="${HOME}/.bashrc"
        elif [ -f "${HOME}/.bash_profile" ]; then
            RC_FILE="${HOME}/.bash_profile"
        fi
    fi

    if [ -n "$RC_FILE" ] && [ -f "$RC_FILE" ]; then
        if ! grep -q 'export PATH=.*\.local/bin' "$RC_FILE"; then
            echo "  export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "$RC_FILE"
            echo "  -> Adicionado automaticamente em ${RC_FILE}."
            echo "  -> Execute 'source ${RC_FILE}' ou reinicie o terminal."
        fi
    else
        echo "  Adicione ao seu ~/.bashrc ou ~/.zshrc:"
        echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
fi

echo ""
echo "Verificação:"
"$TARGET" --version || true

echo ""
echo "Próximos passos:"
echo "  1. mrtn sync     # Sincroniza os casos de teste oficiais"
echo "  2. mrtn info     # Lista as provas e questões disponíveis"
echo "  3. mrtn run      # Executa e valida sua solução contra o juiz"
echo ""
echo "Instalação concluída com sucesso! 🎈"
