#!/bin/bash

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="${PROJECT_ROOT}/bin"

ALL_COMPONENTS=("server_api" "server_web" "client" "server_cli")

ALL_PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Parse arguments
if [ $# -eq 0 ]; then
    PLATFORMS=("${ALL_PLATFORMS[@]}")
    COMPONENTS=("${ALL_COMPONENTS[@]}")
else
    TARGET_PLATFORM="$1"
    shift

    # Check platform
    found=false
    for p in "${ALL_PLATFORMS[@]}"; do
        if [[ "$p" == "$TARGET_PLATFORM" ]]; then
            found=true
            break
        fi
    done

    if [[ "$found" == false ]]; then
        echo "Invalid platform: $TARGET_PLATFORM"
        echo "Available platforms:"
        printf '  %s\n' "${ALL_PLATFORMS[@]}"
        exit 1
    fi

    PLATFORMS=("$TARGET_PLATFORM")

    # Remaining arguments are components
    if [ $# -gt 0 ]; then
        COMPONENTS=()

        for c in "$@"; do
            found=false

            for ac in "${ALL_COMPONENTS[@]}"; do
                if [[ "$ac" == "$c" ]]; then
                    found=true
                    break
                fi
            done

            if [[ "$found" == false ]]; then
                echo "Invalid component: $c"
                echo "Available components:"
                printf '  %s\n' "${ALL_COMPONENTS[@]}"
                exit 1
            fi

            COMPONENTS+=("$c")
        done
    else
        COMPONENTS=("${ALL_COMPONENTS[@]}")
    fi
fi


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

        GOOS="${GOOS}" \
        GOARCH="${GOARCH}" \
        go build -o "${out_dir}/${comp}${ext}" \
        "${PROJECT_ROOT}/${comp}"
    done

    echo ""
done


echo "=== Build Complete ==="
echo "Output directory: ${OUTPUT_DIR}"

find "${OUTPUT_DIR}" -type f \( -executable -o -name "*.exe" \) | sort |
while read -r f; do
    echo "  $(basename "$(dirname "$(dirname "$f")")")/$(basename "$(dirname "$f")")/$(basename "$f")  ($(du -h "$f" | cut -f1))"
done