#!/bin/bash

echo "🔧 Configurando Git LFS para archivos grandes..."

# Verificar si Git LFS está instalado
if ! command -v git-lfs &> /dev/null; then
    echo "❌ Git LFS no está instalado."
    echo ""
    echo "Instálalo con:"
    echo "  macOS:   brew install git-lfs"
    echo "  Ubuntu:  sudo apt-get install git-lfs"
    echo "  Fedora:  sudo dnf install git-lfs"
    exit 1
fi

# Inicializar Git LFS
git lfs install

# Trackear archivos grandes
git lfs track "internal/geometry/datosLC.json"
git lfs track "internal/geometry/latlonCPWorldEste.json"

# Agregar .gitattributes
git add .gitattributes

echo ""
echo "✅ Git LFS configurado correctamente"
echo ""
echo "Archivos trackeados:"
echo "  - internal/geometry/datosLC.json (~64MB)"
echo "  - internal/geometry/latlonCPWorldEste.json (~49MB)"
echo ""
echo "Ahora puedes hacer commit y push normalmente:"
echo "  git add internal/geometry/*.json"
echo "  git commit -m 'Add geographic polygon data with LFS'"
echo "  git push"
