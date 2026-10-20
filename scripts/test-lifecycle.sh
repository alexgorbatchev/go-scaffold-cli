#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
just --justfile ./justfile setup
just --justfile ./justfile setup
just --justfile ./justfile run --help >/dev/null
mkdir -p .tmp
fixture="$(mktemp -d .tmp/lifecycle.XXXXXX)"
trap 'rm -rf "$fixture"' EXIT
just --justfile ./justfile run cli create "$fixture/cli" --name lifecycle-cli --module example.com/lifecycle-cli --author Test --no-git
(cd "$fixture/cli" && just --justfile ./justfile setup && just --justfile ./justfile setup && just --justfile ./justfile run status)
just --justfile ./justfile run lib create "$fixture/lib" --name lifecycle-lib --module example.com/lifecycle-lib --author Test --no-git
(cd "$fixture/lib" && just --justfile ./justfile setup && just --justfile ./justfile setup && just --justfile ./justfile test)
