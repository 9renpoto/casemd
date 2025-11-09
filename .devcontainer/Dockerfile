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

USER vscode
RUN NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" \
	&& eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" \
	&& brew install typos-cli lefthook \
	&& echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/vscode/.bashrc

USER vscode
