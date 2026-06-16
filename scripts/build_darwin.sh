#!/bin/sh

# Note:
#  While testing, if you double-click on the Lychee.app
#  some state is left on MacOS and subsequent attempts
#  to build again will fail with:
#
#    hdiutil: create failed - Operation not permitted
#
#  To work around, specify another volume name with:
#
#    VOL_NAME="$(date)" ./scripts/build_darwin.sh
#
VOL_NAME=${VOL_NAME:-"Lychee"}
export VERSION=${VERSION:-$(git describe --tags --first-parent --abbrev=7 --long --dirty --always | sed -e "s/^v//g")}
export CGO_CFLAGS="-O3 -mmacosx-version-min=14.0"
export CGO_CXXFLAGS="-O3 -mmacosx-version-min=14.0"
export CGO_LDFLAGS="-mmacosx-version-min=14.0"

set -e

status() { echo >&2 ">>> $@"; }
usage() {
    echo "usage: $(basename $0) [build package app sign]"
    exit 1
}

mkdir -p dist

ARCHS="arm64 amd64"
while getopts "a:h" OPTION; do
    case $OPTION in
        a) ARCHS=$OPTARG ;;
        h) usage ;;
    esac
done

shift $(( $OPTIND - 1 ))

_build_darwin() {
    SOURCE_BUILD=build/darwin-sources
    status "Preparing shared native sources"
    cmake -S . -B "$SOURCE_BUILD" -DLYCHEE_MLX_BACKENDS=metal_v3 -DLYCHEE_LLAMA_BACKENDS=
    cmake --build "$SOURCE_BUILD" --target lychee-llama-cpp-source --target lychee-mlx-sources
    LLAMA_CPP_SHARED_SRC="$(pwd)/$SOURCE_BUILD/_deps/llama_cpp-src"
    MLX_SHARED_SRC="$(pwd)/$SOURCE_BUILD/_deps/mlx-src"
    MLX_C_SHARED_SRC="$(pwd)/$SOURCE_BUILD/_deps/mlx-c-src"

    for ARCH in $ARCHS; do
        status "Building darwin $ARCH"
        INSTALL_PREFIX=dist/darwin-$ARCH/
        BUILD_DIR=build/darwin-$ARCH

        if [ "$ARCH" = "amd64" ]; then
            CMAKE_ARCH=x86_64
            MLX_BACKENDS=metal_v3
            MLX_EXTRA_ARGS="-DMLX_ENABLE_X64_MAC=ON"
            MLX_CGO_CFLAGS="-O3 -mmacosx-version-min=14.0"
            MLX_CGO_LDFLAGS="-ldl -lc++ -framework Accelerate -mmacosx-version-min=14.0"
        else
            CMAKE_ARCH=arm64
            MLX_BACKENDS="metal_v3;metal_v4"
            MLX_EXTRA_ARGS=
            MLX_CGO_CFLAGS="-O3 -mmacosx-version-min=14.0"
            MLX_CGO_LDFLAGS="-lc++ -framework Metal -framework Foundation -framework Accelerate -mmacosx-version-min=14.0"
        fi

        cmake -S . -B "$BUILD_DIR" \
            -DCMAKE_BUILD_TYPE=Release \
            -DCMAKE_OSX_ARCHITECTURES=$CMAKE_ARCH \
            -DCMAKE_OSX_DEPLOYMENT_TARGET=14.0 \
            -DCMAKE_INSTALL_PREFIX=$INSTALL_PREFIX \
            -DLYCHEE_PAYLOAD_INSTALL_PREFIX=$INSTALL_PREFIX \
            -DLYCHEE_GO_OUTPUT=$INSTALL_PREFIX/lychee \
            -DLYCHEE_VERSION="$VERSION" \
            -DLYCHEE_MLX_BACKENDS="$MLX_BACKENDS" \
            -DLYCHEE_LLAMA_BACKENDS= \
            -DFETCHCONTENT_SOURCE_DIR_LLAMA_CPP=$LLAMA_CPP_SHARED_SRC \
            -DFETCHCONTENT_SOURCE_DIR_MLX=$MLX_SHARED_SRC \
            -DFETCHCONTENT_SOURCE_DIR_MLX-C=$MLX_C_SHARED_SRC \
            $MLX_EXTRA_ARGS

        GOOS=darwin GOARCH=$ARCH CGO_ENABLED=1 CGO_CFLAGS="$MLX_CGO_CFLAGS" CGO_LDFLAGS="$MLX_CGO_LDFLAGS" \
            cmake --build "$BUILD_DIR" --target lychee-local --target lychee-mlx-backends --parallel
    done
}

_merge_darwin_payload() {
    status "Preparing universal Darwin runtime payload"
    rm -rf dist/darwin/lib
    mkdir -p dist/darwin/lib/lychee

    for ROOT in dist/darwin-amd64/lib/lychee dist/darwin-arm64/lib/lychee; do
        [ -d "$ROOT" ] || continue
        for F in "$ROOT"/*; do
            [ -e "$F" ] || continue
            BASE=$(basename "$F")
            case "$BASE" in
                llama-server|llama-quantize|mlx_*) continue ;;
            esac
            [ -e "dist/darwin/lib/lychee/$BASE" ] || cp -P "$F" dist/darwin/lib/lychee/
        done
    done

    for VARIANT in dist/darwin-arm64/lib/lychee/mlx_metal_v*/; do
        [ -d "$VARIANT" ] || continue
        VNAME=$(basename "$VARIANT")
        DEST=dist/darwin/lib/lychee/$VNAME
        AMD_VARIANT=dist/darwin-amd64/lib/lychee/$VNAME
        [ -d "$AMD_VARIANT" ] || AMD_VARIANT=dist/darwin-amd64/lib/lychee
        mkdir -p "$DEST"

        for LIB in libmlx.dylib libmlxc.dylib; do
            if [ -f "$AMD_VARIANT/$LIB" ] && [ -f "$VARIANT$LIB" ]; then
                lipo -create -output "$DEST/$LIB" "$AMD_VARIANT/$LIB" "$VARIANT$LIB"
            elif [ -f "$VARIANT$LIB" ]; then
                cp "$VARIANT$LIB" "$DEST/"
            elif [ -f "$AMD_VARIANT/$LIB" ]; then
                cp "$AMD_VARIANT/$LIB" "$DEST/"
            fi
        done

        for F in "$VARIANT"*; do
            [ -f "$F" ] && [ ! -L "$F" ] || continue
            case "$(basename "$F")" in
                libmlx.dylib|libmlxc.dylib) continue ;;
            esac
            cp "$F" "$DEST/"
        done
    done
}

_prepare_darwin_runtime() {
    status "Creating universal binary..."
    mkdir -p dist/darwin
    lipo -create -output dist/darwin/lychee dist/darwin-amd64/lychee dist/darwin-arm64/lychee
    chmod +x dist/darwin/lychee
    lipo dist/darwin/lychee -verify_arch x86_64 arm64

    lipo -create -output dist/darwin/llama-server dist/darwin-amd64/lib/lychee/llama-server dist/darwin-arm64/lib/lychee/llama-server
    chmod +x dist/darwin/llama-server
    lipo dist/darwin/llama-server -verify_arch x86_64 arm64

    lipo -create -output dist/darwin/llama-quantize dist/darwin-amd64/lib/lychee/llama-quantize dist/darwin-arm64/lib/lychee/llama-quantize
    chmod +x dist/darwin/llama-quantize
    lipo dist/darwin/llama-quantize -verify_arch x86_64 arm64

    _merge_darwin_payload
}

_create_darwin_runtime_tarball() {
    status "Creating universal tarball..."
    rm -f dist/lychee-darwin.tar dist/lychee-darwin.tgz
    tar -cf dist/lychee-darwin.tar --strip-components 2 dist/darwin/lychee dist/darwin/llama-server dist/darwin/llama-quantize
    tar -rf dist/lychee-darwin.tar --strip-components 4 dist/darwin/lib/lychee
    gzip -9vc <dist/lychee-darwin.tar >dist/lychee-darwin.tgz
}

_package_darwin_runtime() {
    _prepare_darwin_runtime
    _create_darwin_runtime_tarball
}

_sign_darwin() {
    _prepare_darwin_runtime
    if [ -n "$APPLE_IDENTITY" ]; then
        for F in dist/darwin/lychee dist/darwin/llama-server dist/darwin/llama-quantize dist/darwin/lib/lychee/* dist/darwin/lib/lychee/mlx_metal_v*/*; do
            [ -f "$F" ] && [ ! -L "$F" ] || continue
            codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime "$F"
        done

        # create a temporary zip for notarization
        TEMP=$(mktemp -u).zip
        ditto -c -k --keepParent dist/darwin/lychee "$TEMP"
        xcrun notarytool submit "$TEMP" --wait --timeout 20m --apple-id $APPLE_ID --password $APPLE_PASSWORD --team-id $APPLE_TEAM_ID
        rm -f "$TEMP"
    fi

    _create_darwin_runtime_tarball
}

_build_macapp() {
    if ! command -v npm &> /dev/null; then
        echo "npm is not installed. Please install Node.js and npm first:"
        echo "   Visit: https://nodejs.org/"
        exit 1
    fi

    if ! command -v tsc &> /dev/null; then
        echo "Installing TypeScript compiler..."
        npm install -g typescript
    fi

    echo "Installing required Go tools..."

    cd app/ui/app
    npm install
    npm run build
    cd ../../..

    # Build the Lychee.app bundle
    rm -rf dist/Lychee.app
    cp -a ./app/darwin/Lychee.app dist/Lychee.app

    # update the modified date of the app bundle to now
    touch dist/Lychee.app

    go clean -cache
    GOARCH=amd64 CGO_ENABLED=1 GOOS=darwin go build -o dist/darwin-app-amd64 -ldflags="-s -w -X=github.com/MD-Mushfiqur123/lychee/app/version.Version=${VERSION}" ./app/cmd/app
    GOARCH=arm64 CGO_ENABLED=1 GOOS=darwin go build -o dist/darwin-app-arm64 -ldflags="-s -w -X=github.com/MD-Mushfiqur123/lychee/app/version.Version=${VERSION}" ./app/cmd/app
    mkdir -p dist/Lychee.app/Contents/MacOS
    lipo -create -output dist/Lychee.app/Contents/MacOS/Lychee dist/darwin-app-amd64 dist/darwin-app-arm64
    rm -f dist/darwin-app-amd64 dist/darwin-app-arm64

    # Create a mock Squirrel.framework bundle
    mkdir -p dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Versions/A/Resources/
    cp -a dist/Lychee.app/Contents/MacOS/Lychee dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Versions/A/Squirrel
    ln -s ../Squirrel dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Versions/A/Resources/ShipIt
    cp -a ./app/cmd/squirrel/Info.plist dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Versions/A/Resources/Info.plist
    ln -s A dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Versions/Current
    ln -s Versions/Current/Resources dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Resources
    ln -s Versions/Current/Squirrel dist/Lychee.app/Contents/Frameworks/Squirrel.framework/Squirrel

    # Update the version in the Info.plist
    plutil -replace CFBundleShortVersionString -string "$VERSION" dist/Lychee.app/Contents/Info.plist
    plutil -replace CFBundleVersion -string "$VERSION" dist/Lychee.app/Contents/Info.plist

    # Setup the lychee binaries
    mkdir -p dist/Lychee.app/Contents/Resources
    [ -d dist/darwin/lib/lychee ] || _merge_darwin_payload
    cp -a dist/darwin/lychee dist/Lychee.app/Contents/Resources/lychee
    cp dist/darwin/llama-server dist/Lychee.app/Contents/Resources/
    cp dist/darwin/llama-quantize dist/Lychee.app/Contents/Resources/
    if [ -d dist/darwin/lib/lychee ]; then
        cp -a dist/darwin/lib/lychee/. dist/Lychee.app/Contents/Resources/
    fi
    chmod a+x dist/Lychee.app/Contents/Resources/lychee

    # Sign
    if [ -n "$APPLE_IDENTITY" ]; then
        codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime dist/Lychee.app/Contents/Resources/lychee
        codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime dist/Lychee.app/Contents/Resources/llama-server
        codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime dist/Lychee.app/Contents/Resources/llama-quantize
        for lib in dist/Lychee.app/Contents/Resources/*.so dist/Lychee.app/Contents/Resources/*.dylib dist/Lychee.app/Contents/Resources/*.metallib dist/Lychee.app/Contents/Resources/mlx_metal_v*/*.dylib dist/Lychee.app/Contents/Resources/mlx_metal_v*/*.metallib dist/Lychee.app/Contents/Resources/mlx_metal_v*/*.so; do
            [ -f "$lib" ] || continue
            codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime "$lib"
        done
        codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier com.electron.lychee --deep --options=runtime dist/Lychee.app
    fi

    rm -f dist/Lychee-darwin.zip
    ditto -c -k --norsrc --keepParent dist/Lychee.app dist/Lychee-darwin.zip
    (cd dist/Lychee.app/Contents/Resources/; tar -cf - lychee llama-server llama-quantize *.so *.dylib *.metallib mlx_metal_v*/ 2>/dev/null) | gzip -9vc > dist/lychee-darwin.tgz

    # Notarize and Staple
    if [ -n "$APPLE_IDENTITY" ]; then
        $(xcrun -f notarytool) submit dist/Lychee-darwin.zip --wait --timeout 20m --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$APPLE_TEAM_ID"
        rm -f dist/Lychee-darwin.zip
        $(xcrun -f stapler) staple dist/Lychee.app
        ditto -c -k --norsrc --keepParent dist/Lychee.app dist/Lychee-darwin.zip

        rm -f dist/Lychee.dmg

        (cd dist && ../scripts/create-dmg.sh \
            --volname "${VOL_NAME}" \
            --volicon ../app/darwin/Lychee.app/Contents/Resources/icon.icns \
            --background ../app/assets/background.png \
            --window-pos 200 120 \
            --window-size 800 400 \
            --icon-size 128 \
            --icon "Lychee.app" 200 190 \
            --hide-extension "Lychee.app" \
            --app-drop-link 600 190 \
            --text-size 12 \
            "Lychee.dmg" \
            "Lychee.app" \
        ; )
        rm -f dist/rw*.dmg

        codesign -f --timestamp -s "$APPLE_IDENTITY" --identifier ai.lychee.lychee --options=runtime dist/Lychee.dmg
        $(xcrun -f notarytool) submit dist/Lychee.dmg --wait --timeout 20m --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$APPLE_TEAM_ID"
        $(xcrun -f stapler) staple dist/Lychee.dmg
    else
        echo "WARNING: Code signing disabled, this bundle will not work for upgrade testing"
    fi
}

if [ "$#" -eq 0 ]; then
    _build_darwin
    _sign_darwin
    _build_macapp
    exit 0
fi

for CMD in "$@"; do
    case $CMD in
        build) _build_darwin ;;
        package) _package_darwin_runtime ;;
        sign) _sign_darwin ;;
        app) _build_macapp ;;
        *) usage ;;
    esac
done
