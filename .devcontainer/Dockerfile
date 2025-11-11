# Go-focused development container with CLI tooling support.
FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV HOMEBREW_PREFIX=/home/linuxbrew/.linuxbrew \
	HOMEBREW_CELLAR=/home/linuxbrew/.linuxbrew/Cellar \
	HOMEBREW_REPOSITORY=/home/linuxbrew/.linuxbrew/Homebrew \
	PATH=/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:$PATH \
	MANPATH=/home/linuxbrew/.linuxbrew/share/man:$MANPATH \
	INFOPATH=/home/linuxbrew/.linuxbrew/share/info:$INFOPATH

USER root
RUN apt-get update \
	&& apt-get install -y --no-install-recommends build-essential procps curl file git ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& mkdir -p /home/linuxbrew \
	&& chown -R vscode:vscode /home/linuxbrew

RUN set -euo pipefail \
	&& curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh -o /tmp/install.sh \
	&& chmod +x /tmp/install.sh \
	&& su vscode -c 'NONINTERACTIVE=1 /tmp/install.sh' \
	&& echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /etc/profile.d/brew.sh \
	&& rm /tmp/install.sh

COPY Brewfile /tmp/Brewfile
RUN set -euo pipefail \
	&& chown vscode:vscode /tmp/Brewfile \
	&& su vscode -c 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" && brew update && brew bundle --file=/tmp/Brewfile' \
	&& rm /tmp/Brewfile

USER vscode
