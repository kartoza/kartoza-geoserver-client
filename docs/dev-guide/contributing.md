# Contributing

We welcome contributions to CloudBench!

## Development Setup

```bash
# Clone the repository
git clone https://github.com/kartoza/kartoza-cloudbench.git
cd kartoza-cloudbench

# Enter development environment
nix develop

# Run the server
python manage.py runserver 8080

# Run tests
pytest

# Lint code
ruff check .
```

## Code Style

- **Python**: Follow PEP 8, enforced by Ruff
- **TypeScript**: ESLint + Prettier
- **Commits**: Conventional commits format

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `pytest`
5. Run linter: `ruff check .`
6. Submit a pull request

## Testing

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=apps

# Run specific tests
pytest tests/unit/
```

## Documentation

Documentation uses MkDocs:

```bash
# Serve docs locally
mkdocs serve

# Build docs
mkdocs build
```

## Reporting Issues

Please use GitHub Issues for:

- Bug reports
- Feature requests
- Questions

Include:
- CloudBench version
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs

## Code of Conduct

Be respectful and inclusive. See CODE_OF_CONDUCT.md.
