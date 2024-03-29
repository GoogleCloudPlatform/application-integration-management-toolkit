#!/bin/sh
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e

# Determines the operating system.
OS="$(uname)"
if [ "${OS}" = "Darwin" ] ; then
  OSEXT="Darwin"
else
  OSEXT="Linux"
fi

# Determine the latest INTEGRATIONCLI version by version number ignoring alpha, beta, and rc versions.
if [ "${INTEGRATIONCLI_VERSION}" = "" ] ; then
  INTEGRATIONCLI_VERSION="$(curl -sL https://api.github.com/repos/GoogleCloudPlatform/application-integration-management-toolkit/releases/latest | grep tag_name | sed -E 's/.*"([^"]+)".*/\1/')"
fi

LOCAL_ARCH=$(uname -m)
if [ "${TARGET_ARCH}" ]; then
    LOCAL_ARCH=${TARGET_ARCH}
fi

case "${LOCAL_ARCH}" in
  x86_64|amd64|arm64)
    INTEGRATIONCLI_ARCH=x86_64
    ;;
  armv8*|aarch64*)
    INTEGRATIONCLI_ARCH=arm64
    ;;
  *)
    echo "This system's architecture, ${LOCAL_ARCH}, isn't supported"
    exit 1
    ;;
esac

if [ "${INTEGRATIONCLI_VERSION}" = "" ] ; then
  printf "Unable to get latest INTEGRATIONCLI version. Set INTEGRATIONCLI_VERSION env var and re-run. For example: export INTEGRATIONCLI_VERSION=v1.104"
  exit 1;
fi

# Downloads the INTEGRATIONCLI binary archive.
tmp=$(mktemp -d /tmp/integrationcli.XXXXXX)
NAME="integrationcli_$INTEGRATIONCLI_VERSION"

cd "$tmp" || exit
URL="https://github.com/GoogleCloudPlatform/application-integration-management-toolkit/releases/download/${INTEGRATIONCLI_VERSION}/integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}.zip"
SIG_URL="https://github.com/GoogleCloudPlatform/application-integration-management-toolkit/releases/download/${INTEGRATIONCLI_VERSION}/integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}.zip.sig"
COSIGN_PUBLIC_KEY="https://raw.githubusercontent.com/GoogleCloudPlatform/application-integration-management-toolkit/main/cosign.pub"

download_cli() {
  printf "\nDownloading %s from %s ...\n" "$NAME" "$URL"
  if ! curl -o /dev/null -sIf "$URL"; then
    printf "\n%s is not found, please specify a valid INTEGRATIONCLI_VERSION and TARGET_ARCH\n" "$URL"
    exit 1
  fi
  curl -fsLO -H 'Cache-Control: no-cache, no-store' "$URL"
  filename="integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}.zip"
  # Check if cosign is installed
  set +e # disable exit on error
  cosign version 2>&1 >/dev/null
  RESULT=$?
  set -e # re-enable exit on error
  if [ $RESULT -eq 0 ]; then
    echo "Verifying the signature of the binary " "$filename"
    echo "Downloading the cosign public key"
    curl -fsLO -H 'Cache-Control: no-cache, no-store' "$COSIGN_PUBLIC_KEY"
    echo "Downloading the signature file " "$SIG_URL"
    curl -fsLO -H 'Cache-Control: no-cache, no-store' "$SIG_URL"
    sig_filename="integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}.zip.sig"
    echo "Verifying the signature"
    cosign verify-blob --key cosign.pub --signature "$sig_filename" "$filename"
    rm "$sig_filename"
    rm cosign.pub
  else
    echo "cosign is not installed, skipping signature verification"
  fi
  unzip "${filename}"
  rm "${filename}"
}


download_cli

printf ""
printf "\nintegrationcli %s Download Complete!\n" "$INTEGRATIONCLI_VERSION"
printf "\n"
printf "integrationcli has been successfully downloaded into the %s folder on your system.\n" "$tmp"
printf "\n"

# setup INTEGRATIONCLI
cd "$HOME" || exit
mkdir -p "$HOME/.integrationcli/bin"

mv "${tmp}/integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}/integrationcli" "$HOME/.integrationcli/bin"
mv "${tmp}/integrationcli_${OSEXT}_${INTEGRATIONCLI_ARCH}/LICENSE.txt" "$HOME/.integrationcli/LICENSE.txt"

printf "Copied integrationcli into the $HOME/.integrationcli/bin folder.\n"
chmod +x "$HOME/.integrationcli/bin/integrationcli"
rm -r "${tmp}"

# Print message
printf "\n"
printf "Please add integrationcli to your path:"
printf "\n"
printf "  export PATH=\$PATH:\$HOME/.integrationcli/bin \n"
printf "\n"

export PATH=$PATH:$HOME/.integrationcli/bin
