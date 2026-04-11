# Installation

## Prerequisites

- Python 3.12 or higher
- Node.js 22 or higher (for frontend development)
- PostgreSQL (optional, for production database)

## Using Nix (Recommended)

The easiest way to get started is using Nix, which provides a reproducible development environment with all dependencies.

```bash
# Clone the repository
git clone https://github.com/kartoza/kartoza-cloudbench.git
cd kartoza-cloudbench

# Enter the development shell
nix develop

# Run database migrations
python manage.py migrate

# Start the development server
python manage.py runserver 8080
```

## Using pip

```bash
# Clone the repository
git clone https://github.com/kartoza/kartoza-cloudbench.git
cd kartoza-cloudbench

# Create a virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install the package
pip install -e .

# Run database migrations
python manage.py migrate

# Build the frontend (optional, pre-built assets included)
cd web
npm install
npm run build
cd ..

# Start the server
python manage.py runserver 8080
```

## Docker

```bash
# Build and run with Docker Compose
docker compose up -d
```

See [Docker Deployment](../admin-guide/docker.md) for detailed instructions.

## Verify Installation

1. Open [http://localhost:8080](http://localhost:8080) in your browser
2. You should see the CloudBench web interface
3. Add a GeoServer connection to start managing your geospatial data

## Next Steps

- [Quick Start Guide](quickstart.md) - Get up and running quickly
- [Configuration](configuration.md) - Configure CloudBench for your environment
