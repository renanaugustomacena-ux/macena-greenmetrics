#!/usr/bin/env bash
# Devcontainer post-create — install per-language tooling not covered by features.
# Doctrine: Rule 17 (DX), Rule 23 (tooling discipline).
set -Eeuo pipefail

echo "[devcontainer] installing Go tooling…"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/google/go-licenses@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install gotest.tools/gotestsum@latest

echo "[devcontainer] installing python tooling…"
pip install --user pre-commit yamllint

echo "[devcontainer] installing security + policy tooling…"
# conftest
curl -sL https://github.com/open-policy-agent/conftest/releases/latest/download/conftest_linux_x86_64.tar.gz | tar xz -C /tmp
sudo mv /tmp/conftest /usr/local/bin/
# kubeconform
curl -sL https://github.com/yannh/kubeconform/releases/latest/download/kubeconform-linux-amd64.tar.gz | tar xz -C /tmp
sudo mv /tmp/kubeconform /usr/local/bin/
# trivy
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
# cosign
curl -sL "https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64" -o /tmp/cosign
chmod +x /tmp/cosign && sudo mv /tmp/cosign /usr/local/bin/
# syft
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
# osv-scanner
go install github.com/google/osv-scanner/cmd/osv-scanner@latest
# hadolint
sudo curl -sL https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64 -o /usr/local/bin/hadolint
sudo chmod +x /usr/local/bin/hadolint
# actionlint
go install github.com/rhysd/actionlint/cmd/actionlint@latest
# tfsec
curl -sL https://github.com/aquasecurity/tfsec/releases/latest/download/tfsec-linux-amd64 -o /tmp/tfsec
chmod +x /tmp/tfsec && sudo mv /tmp/tfsec /usr/local/bin/
# argocd CLI
sudo curl -sL https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64 -o /usr/local/bin/argocd
sudo chmod +x /usr/local/bin/argocd
# k6
sudo curl -sL https://github.com/grafana/k6/releases/latest/download/k6-v0.55.0-linux-amd64.tar.gz | sudo tar xz -C /usr/local/bin/ --strip-components=1 k6-v0.55.0-linux-amd64/k6 || true
# k9s
go install github.com/derailed/k9s@latest

echo "[devcontainer] installing pre-commit hooks…"
cd "/workspaces/$(basename "$(pwd)")"
pre-commit install || true
pre-commit install --hook-type commit-msg || true

echo "[devcontainer] post-create complete."
