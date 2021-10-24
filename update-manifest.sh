#!/bin/bash
set -eu

# matches to git tags, e.g. "v0.0.3"
VER="$1"

# ex.) https://github.com/shusugmt/kubectl-sealer/releases/download/v0.0.3/kubectl-sealer_v0.0.3_checksums.txt
TARGETS="$(curl -L https://github.com/shusugmt/kubectl-sealer/releases/download/${VER}/kubectl-sealer_${VER}_checksums.txt)"

IFS=$'\n'
for TARGET in $TARGETS; do
  # ex.) 50636519e0e2bcfa836478388561a8b8b4c9ef0a1c6c0bd2926f189a3ef3408f  kubectl-sealer_v0.0.3_linux_arm64.tar.gz
  if [[ "${TARGET}" =~ ^([0-9a-f]+)[\ ]+kubectl-sealer_${VER}_([0-9a-z]+)_([0-9a-z]+)\.tar\.gz$ ]]; then
    SHA="${BASH_REMATCH[1]}"
    OS="${BASH_REMATCH[2]}"
    ARCH="${BASH_REMATCH[3]}"

    # ex.) https://github.com/shusugmt/kubectl-sealer/releases/download/v0.0.3/kubectl-sealer_v0.0.3_linux_arm64.tar.gz
    URI="https://github.com/shusugmt/kubectl-sealer/releases/download/${VER}/kubectl-sealer_${VER}_${OS}_${ARCH}.tar.gz"
    OS=$OS ARCH=$ARCH URI=$URI yq e '(.spec.platforms[] | select(.selector.matchLabels.os == env(OS)) | select(.selector.matchLabels.arch == env(ARCH))) |= .uri    = env(URI)' plugins/sealer.yaml -i
    OS=$OS ARCH=$ARCH SHA=$SHA yq e '(.spec.platforms[] | select(.selector.matchLabels.os == env(OS)) | select(.selector.matchLabels.arch == env(ARCH))) |= .sha256 = env(SHA)' plugins/sealer.yaml -i
  fi
done
