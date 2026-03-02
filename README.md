# DurableCompose

A tool for defining and running multi-container Docker applications using the [Compose file format](https://compose-spec.io).

DurableCompose enables you to define multi-container application configurations in a single `compose.yaml` file and orchestrate them with simple commands. Built on proven foundations with a focus on long-term maintainability and reliability.

## Features

- **Simple Configuration**: Define multi-container applications in a single YAML file
- **Complete Lifecycle Management**: Start, stop, rebuild, and manage services with intuitive commands
- **Watch Mode**: Automatically rebuild and restart services when code changes
- **Service Dependencies**: Define and manage dependencies between services
- **Network Isolation**: Automatic network creation for service communication
- **Volume Management**: Persistent data storage and sharing between containers
- **Compose Spec Compliant**: Full compatibility with the [Compose specification](https://compose-spec.io)

## Quick Start

```bash
# Install DurableCompose (see Installation section for more options)
# On macOS/Linux via Homebrew
brew install durablecompose

# On Windows via winget
winget install DurableProgramming.DurableCompose

# Verify installation
durablecompose version
```

Create a `compose.yaml` file:

```yaml
services:
  web:
    build: .
    ports:
      - "5000:5000"
    volumes:
      - .:/code
  redis:
    image: redis:alpine
```

Start your application:

```bash
durablecompose up
```

## Installation

### Recommended Methods

<details>
<summary><b>macOS</b></summary>

#### Homebrew (Recommended)
```bash
brew install durablecompose
```

#### Standalone Binary
```bash
# Intel Macs
curl -L -o durablecompose https://github.com/durable_oss/durablecompose/releases/latest/download/durablecompose-darwin-amd64
chmod +x durablecompose
sudo mv durablecompose /usr/local/bin/

# Apple Silicon Macs
curl -L -o durablecompose https://github.com/durable_oss/durablecompose/releases/latest/download/durablecompose-darwin-arm64
chmod +x durablecompose
sudo mv durablecompose /usr/local/bin/
```

</details>

<details>
<summary><b>Windows</b></summary>

#### winget (Recommended)
```powershell
winget install DurableProgramming.DurableCompose
```

#### Chocolatey
```powershell
choco install durablecompose
```

#### Standalone Binary
1. Download [durablecompose.exe](https://github.com/durable_oss/durablecompose/releases/latest/download/durablecompose-windows-amd64.exe)
2. Move to `C:\Program Files\DurableCompose\`
3. Add to PATH: System Properties → Environment Variables → Path

</details>

<details>
<summary><b>Linux</b></summary>

#### Ubuntu/Debian (APT)
```bash
sudo add-apt-repository ppa:durableprogramming/durablecompose
sudo apt update
sudo apt install durablecompose
```

#### Arch Linux (AUR)
```bash
yay -S durablecompose
```

#### Fedora/RHEL
```bash
sudo dnf install durablecompose
```

#### Universal Binary (All Distributions)
```bash
curl -L -o durablecompose https://github.com/durable_oss/durablecompose/releases/latest/download/durablecompose-linux-amd64
chmod +x durablecompose
sudo mv durablecompose /usr/local/bin/
```

</details>

### From Source

```bash
git clone https://github.com/durable_oss/durablecompose.git
cd durablecompose
make build
```

The binary will be available in `bin/build/durablecompose`.

### Docker

```bash
docker pull durableprogramming/durablecompose:latest
docker run --rm durableprogramming/durablecompose:latest version
```

### Verification

After installation, verify it works:

```bash
durablecompose version
# Expected: durablecompose version vX.Y.Z

durablecompose --help
# Should display help information
```

## Usage

### Basic Commands

```bash
# Start services in foreground
durablecompose up

# Start services in background
durablecompose up -d

# Stop services
durablecompose down

# View service logs
durablecompose logs

# List running services
durablecompose ps

# Execute command in running service
durablecompose exec web sh
```

### Watch Mode

Automatically rebuild and restart services when files change:

```bash
durablecompose up --watch
```

Configure watch in `compose.yaml`:

```yaml
services:
  web:
    build: .
    develop:
      watch:
        - path: ./src
          action: sync
          target: /app/src
        - path: package.json
          action: rebuild
```

### Common Workflows

**Development Workflow**:
```bash
# Start with watch mode
durablecompose up --watch

# View logs from specific service
durablecompose logs -f web

# Rebuild after dependency changes
durablecompose up --build
```

**Production Deployment**:
```bash
# Pull latest images
durablecompose pull

# Start services in background
durablecompose up -d

# View status
durablecompose ps
```

See the [Compose Specification](https://compose-spec.io) for complete file format documentation.

## About This Project

DurableCompose is a product of **Durable Programming**, forked from docker-compose v5 with a focus on:

- **Long-term maintainability**: Building software that remains useful and maintainable over years
- **Pragmatic improvements**: Addressing real-world problems with proven solutions
- **Quality and reliability**: Comprehensive testing and careful consideration of edge cases
- **Community-driven development**: Open collaboration while maintaining commercial viability

This fork represents Durable Programming's commitment to sustainable software development—creating tools that solve genuine problems and can be maintained effectively over the long term.

### Durable Programming Philosophy

Durable Programming emphasizes:

1. **Practical problem-solving** over technology for its own sake
2. **Incremental improvement** rather than disruptive rewrites
3. **Modular, composable design** for flexibility and reusability
4. **Developer experience** through clear documentation and intuitive interfaces
5. **Security and stability** as non-negotiable requirements

## Contributing

We welcome contributions from the community! See [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Development setup instructions
- Code style guidelines
- Pull request process
- Testing requirements

To report issues or request features, use our [issue tracker](https://github.com/durable_oss/durablecompose/issues/new/choose).

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Support

- **Documentation**: [https://github.com/durable_oss/durablecompose/tree/main/docs](https://github.com/durable_oss/durablecompose/tree/main/docs)
- **Issues**: [GitHub Issues](https://github.com/durable_oss/durablecompose/issues)
- **Discussions**: [GitHub Discussions](https://github.com/durable_oss/durablecompose/discussions)
- **Commercial Support**: [commercial@durableprogramming.com](mailto:commercial@durableprogramming.com)

## Upstream

DurableCompose is forked from [docker-compose](https://github.com/docker/compose) v5. We maintain compatibility with the Compose file format and strive to contribute improvements back to the upstream project where appropriate.

---

**Table of Contents**
- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [About This Project](#about-this-project)
- [Contributing](#contributing)
- [License](#license)
- [Support](#support)
- [Upstream](#upstream)
