# Go-focused development container with CLI tooling support.
FROM mcr.microsoft.com/devcontainers/go:2-1.25-bookworm

# hadolint ignore=DL3008
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

USER root
RUN apt-get update \
	&& curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
	&& apt-get install -y --no-install-recommends nodejs \
	&& npm install -g @evilmartians/lefthook cspell@^6.14.3 @secretlint/secretlint-rule-preset-recommend@^5.0.1 secretlint@^5.0.3 \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

USER vscode
