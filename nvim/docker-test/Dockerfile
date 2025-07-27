FROM fedora:41

# Install system dependencies
RUN dnf update -y && \
    dnf install -y \
        git \
        curl \
        wget \
        gcc \
        gcc-c++ \
        make \
        cmake \
        ninja-build \
        nodejs \
        npm \
        ripgrep \
        fd-find \
        tree \
        which \
        unzip \
        tar \
        gzip \
        python3 \
        python3-pip && \
    dnf clean all

# Install latest Neovim from GitHub releases (as root)
RUN NVIM_VERSION=$(curl -s https://api.github.com/repos/neovim/neovim/releases/latest | grep '"tag_name"' | cut -d'"' -f4) && \
    wget https://github.com/neovim/neovim/releases/download/${NVIM_VERSION}/nvim-linux-x86_64.tar.gz && \
    tar xzf nvim-linux-x86_64.tar.gz && \
    mv nvim-linux-x86_64 /opt/nvim && \
    ln -sf /opt/nvim/bin/nvim /usr/local/bin/nvim && \
    rm nvim-linux-x86_64.tar.gz

# Get user ID from build args to match host user
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=nvimuser

# Create user with matching UID/GID
RUN groupadd -g ${GROUP_ID} ${USERNAME} && \
    useradd -u ${USER_ID} -g ${GROUP_ID} -m -s /bin/bash ${USERNAME}

# Switch to user
USER ${USERNAME}
WORKDIR /home/${USERNAME}

# Create .config directory
RUN mkdir -p /home/${USERNAME}/.config

# Copy nvim configuration
COPY --from=nvim --chown=${USER_ID}:${GROUP_ID} .config/nvim /home/${USERNAME}/.config/nvim

# Set environment variables
ENV PATH="/usr/local/bin:${PATH}"
ENV EDITOR=nvim

# Create a startup script
RUN echo '#!/bin/bash' > /home/${USERNAME}/start.sh && \
    echo 'echo "=== Docker Neovim Test Environment ==="' >> /home/${USERNAME}/start.sh && \
    echo 'echo "Neovim version: $(nvim --version | head -1)"' >> /home/${USERNAME}/start.sh && \
    echo 'echo "Config location: ~/.config/nvim"' >> /home/${USERNAME}/start.sh && \
    echo 'echo "To start Neovim: nvim"' >> /home/${USERNAME}/start.sh && \
    echo 'echo "To run health check: nvim +checkhealth +qa"' >> /home/${USERNAME}/start.sh && \
    echo 'echo "======================================="' >> /home/${USERNAME}/start.sh && \
    echo 'exec "$@"' >> /home/${USERNAME}/start.sh && \
    chmod +x /home/${USERNAME}/start.sh

# Set working directory for file editing
WORKDIR /workspace

ENTRYPOINT ["/home/nvimuser/start.sh"]
CMD ["/bin/bash"]