# Go-focused development container with CLI tooling support.
FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

# hadolint ignore=DL3008
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG TYPOS_VERSION=1.39.0

USER root
RUN apt-get update \
	&& curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& npm install -g @evilmartians/lefthook @secretlint/secretlint-rule-preset-recommend@^5.0.1 secretlint@^5.0.3 \
	&& curl -fsSL "https://github.com/crate-ci/typos/releases/download/v${TYPOS_VERSION}/typos-v${TYPOS_VERSION}-x86_64-unknown-linux-musl.tar.gz" -o /tmp/typos.tar.gz \
	&& tar -xzf /tmp/typos.tar.gz -C /tmp typos \
	&& install -m 0755 /tmp/typos /usr/local/bin/typos \
	&& rm -rf /tmp/typos /tmp/typos.tar.gz /tmp/doc \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

USER vscode
