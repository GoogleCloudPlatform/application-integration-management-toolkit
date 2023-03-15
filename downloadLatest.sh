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
  INTEGRATIONCLI_VERSION="$(curl -sL https://github.com/GoogleCloudPlatform/application-integration-management-toolkit/releases/latest | \
                  grep -i release | grep -v beta | grep -o 'v[0-9]\.[0-9]*' | head -1)"
  INTEGRATIONCLI_VERSION="${INTEGRATIONCLI_VERSION##*/}"
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

download_cli() {
  printf "\nDownloading %s from %s ...\n" "$NAME" "$URL"
  if ! curl -o /dev/null -sIf "$URL"; then
    printf "\n%s is not found, please specify a valid INTEGRATIONCLI_VERSION and TARGET_ARCH\n" "$URL"
    exit 1
  fi
  curl -fsLO -H 'Cache-Control: no-cache, no-store' "$URL"
  filename="integrationcli_${INTEGRATIONCLI_VERSION}_${OSEXT}_${INTEGRATIONCLI_ARCH}.zip"
  unzip "${filename}"
  rm "${filename}"
}


download_cli

printf ""
printf "\integrationcli %s Download Complete!\n" "$INTEGRATIONCLI_VERSION"
printf "\n"
printf "integrationcli has been successfully downloaded into the %s folder on your system.\n" "$tmp"
printf "\n"

# setup INTEGRATIONCLI
cd "$HOME" || exit
mkdir -p "$HOME/.integrationcli/bin"
mv "${tmp}/integrationcli_${INTEGRATIONCLI_VERSION}_${OSEXT}_${INTEGRATIONCLI_ARCH}/integrationcli" "$HOME/.integrationcli/bin"
printf "Copied integrationcli into the $HOME/.integrationcli/bin folder.\n"
chmod +x "$HOME/.integrationcli/bin/integrationcli"
rm -r "${tmp}"

# Print message
printf "\n"
printf "Added the integrationcli to your path with:"
printf "\n"
printf "  export PATH=\$PATH:\$HOME/.integrationcli/bin \n"
printf "\n"

export PATH=$PATH:$HOME/.integrationcli/bin
