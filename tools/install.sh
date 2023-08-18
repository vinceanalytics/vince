#!/usr/bin/env bash
set -euo pipefail

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: Please install bun using Windows Subsystem for Linux'
    exit 1
fi

# Reset
Color_Off=''

# Regular Colors
Red=''
Green=''
Dim='' # White

# Bold
Bold_White=''
Bold_Green=''

if [[ -t 1 ]]; then
    # Reset
    Color_Off='\033[0m' # Text Reset

    # Regular Colors
    Red='\033[0;31m'   # Red
    Green='\033[0;32m' # Green
    Dim='\033[0;2m'    # White

    # Bold
    Bold_Green='\033[1;32m' # Bold Green
    Bold_White='\033[1m'    # Bold White
fi

error() {
    echo -e "${Red}error${Color_Off}:" "$@" >&2
    exit 1
}

info() {
    echo -e "${Dim}$@ ${Color_Off}"
}

info_bold() {
    echo -e "${Bold_White}$@ ${Color_Off}"
}

success() {
    echo -e "${Green}$@ ${Color_Off}"
}

command -v unzip >/dev/null ||
    error 'unzip is required to install bun (see: https://github.com/oven-sh/bun#unzip-is-required)'

if [[ $# -gt 2 ]]; then
    error 'Too many arguments, only 2 are allowed. The first can be a specific tag of bun to install. (e.g. "bun-v0.1.4") The second can be a build variant of bun to install. (e.g. "debug-info")'
fi

case $(uname -ms) in
'Darwin x86_64')
    target=darwin_x86_64
    ;;
'Darwin arm64')
    target=darwin_arm64
    ;;
'Linux aarch64' | 'Linux arm64')
    target=linux_arm64
    ;;
'Linux x86_64' | *)
    target=linux_x86_64
    ;;
esac

if [[ $target = darwin_x86_64 ]]; then
    # Is this process running in Rosetta?
    # redirect stderr to devnull to avoid error message when not running in Rosetta
    if [[ $(sysctl -n sysctl.proc_translated 2>/dev/null) = 1 ]]; then
        target=darwin_arm64
        info "Your shell is running in Rosetta 2. Downloading vince for $target instead"
    fi
fi

GITHUB=${GITHUB-"https://github.com"}

github_repo="$GITHUB/vinceanalytics/vince"




exe_name=vince


if [[ $# = 0 ]]; then
    vince_uri=$github_repo/releases/latest/download/vince_$target.tar.gz
else
    vince_uri=$github_repo/releases/download/$1/vince_$target.tar.gz
fi

install_env=VINCE_INSTALL
bin_env=\$$install_env/bin

install_dir=${!install_env:-$HOME/.vince}
bin_dir=$install_dir/bin
exe=$bin_dir/vince

if [[ ! -d $bin_dir ]]; then
    mkdir -p "$bin_dir" ||
        error "Failed to create install directory \"$bin_dir\""
fi

curl --fail --location --progress-bar --output "$exe.tar.gz" "$vince_uri" ||
    error "Failed to download vince from \"$vince_uri\""

tar -xf  "$exe.tar.gz" -C "$bin_dir"||
    error 'Failed to extract vince'

rm -rf "$exe.tar.gz"

tildify() {
    if [[ $1 = $HOME/* ]]; then
        local replacement=\~/

        echo "${1/$HOME\//$replacement}"
    else
        echo "$1"
    fi
}

success "vince was installed successfully to $Bold_Green$(tildify "$exe")"