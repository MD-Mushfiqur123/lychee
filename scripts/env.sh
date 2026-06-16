# Common environment setup across build*.sh scripts

export VERSION=${VERSION:-$(git describe --tags --first-parent --abbrev=7 --long --dirty --always | sed -e "s/^v//g")}
export GOFLAGS="'-ldflags=-w -s \"-X=github.com/MD-Mushfiqur123/lychee/version.Version=$VERSION\" \"-X=github.com/MD-Mushfiqur123/lychee/server.mode=release\"'"
# TODO - consider `docker buildx ls --format=json` to autodiscover platform capability
PLATFORM=${PLATFORM:-"linux/arm64,linux/amd64"}
DOCKER_ORG=${DOCKER_ORG:-"lychee"}
FINAL_IMAGE_REPO=${FINAL_IMAGE_REPO:-"${DOCKER_ORG}/lychee"}
LYCHEE_COMMON_BUILD_ARGS="--build-arg=GOFLAGS"

add_build_arg() {
    eval "_value=\"\${$1:-}\""
    if [ -n "$_value" ]; then
        LYCHEE_COMMON_BUILD_ARGS="$LYCHEE_COMMON_BUILD_ARGS --build-arg=$1"
    fi
}

for arg in \
    CGO_CFLAGS \
    CGO_CXXFLAGS \
    CMAKEVERSION \
    NINJAVERSION \
    ROCMVERSION \
    JETPACK5VERSION \
    JETPACK6VERSION \
    CUDA12VERSION \
    CUDA13VERSION \
    VULKANVERSION \
    MLX_CUDA_RAM_MB \
    APT_MIRROR \
    LYCHEE_MLX_BUILD_JOBS \
    LYCHEE_MLX_NVCC_THREADS
do
    add_build_arg "$arg"
done

# Forward local MLX source overrides as Docker build contexts
if [ -n "${LYCHEE_MLX_SOURCE:-}" ]; then
    LYCHEE_COMMON_BUILD_ARGS="$LYCHEE_COMMON_BUILD_ARGS --build-context local-mlx=$(cd "$LYCHEE_MLX_SOURCE" && pwd)"
fi
if [ -n "${LYCHEE_MLX_C_SOURCE:-}" ]; then
    LYCHEE_COMMON_BUILD_ARGS="$LYCHEE_COMMON_BUILD_ARGS --build-context local-mlx-c=$(cd "$LYCHEE_MLX_C_SOURCE" && pwd)"
fi
echo "Building Lychee"
echo "VERSION=$VERSION"
echo "PLATFORM=$PLATFORM"
