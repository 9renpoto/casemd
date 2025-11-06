# Go-focused development container with CLI tooling support.
FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG TYPOS_VERSION=1.39.0
ARG LEFTHOOK_VERSION=2.0.2

USER root
RUN curl -fsSL "https://github.com/crate-ci/typos/releases/download/v${TYPOS_VERSION}/typos-v${TYPOS_VERSION}-x86_64-unknown-linux-musl.tar.gz" -o /tmp/typos.tar.gz \
	&& tar -xzf /tmp/typos.tar.gz -C /tmp typos \
	&& install -m 0755 /tmp/typos /usr/local/bin/typos \
	&& curl -fsSL "https://github.com/evilmartians/lefthook/releases/download/v${LEFTHOOK_VERSION}/lefthook_${LEFTHOOK_VERSION}_Linux_x86_64.gz" -o /tmp/lefthook.gz \
	&& gunzip /tmp/lefthook.gz \
	&& install -m 0755 "/tmp/lefthook_${LEFTHOOK_VERSION}_Linux_x86_64" /usr/local/bin/lefthook \
	&& rm -rf /tmp/typos /tmp/typos.tar.gz "/tmp/lefthook_${LEFTHOOK_VERSION}_Linux_x86_64"

USER vscode
