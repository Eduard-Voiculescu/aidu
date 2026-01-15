# syntax=docker/dockerfile:1
FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    curl \
    git \
    jq \
    vim \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code

RUN npm install -g opencode-ai

RUN useradd -m -s /bin/bash claude && \
    mkdir -p /workspace && \
    chown -R claude:claude /workspace

USER claude
WORKDIR /workspace

CMD ["/bin/bash"]
