#!/bin/bash

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${PROJECT_ROOT}/bin"

COMPONENTS=("server_api" "server_web" "client" "server_cli")

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo "=== Building RATFF (Cross-Platform) ==="
echo ""

total=$(( ${#PLATFORMS[@]} * ${#COMPONENTS[@]} ))
current=0

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%%/*}"
    GOARCH="${platform##*/}"

    echo "--- ${GOOS}/${GOARCH} ---"

    for comp in "${COMPONENTS[@]}"; do
        current=$((current + 1))

        if [[ "${GOOS}" == "windows" ]]; then
            ext=".exe"
        else
            ext=""
        fi

        out_dir="${OUTPUT_DIR}/${GOOS}/${GOARCH}"
        mkdir -p "${out_dir}"

        echo "  [${current}/${total}] ${comp} -> ${out_dir}/${comp}${ext}"
        GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o "${out_dir}/${comp}${ext}" "${PROJECT_ROOT}/${comp}"
    done
    echo ""
done

echo "=== Build Complete ==="
echo "Output directory: ${OUTPUT_DIR}"
find "${OUTPUT_DIR}" -type f -executable -o -name "*.exe" | sort | while read -r f; do
    echo "  $(basename "$(dirname "$(dirname "$f")")")/$(basename "$(dirname "$f")")/$(basename "$f")  ($(du -h "$f" | cut -f1))"
done