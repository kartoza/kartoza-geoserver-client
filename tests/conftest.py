"""Pytest configuration and fixtures for Kartoza CloudBench tests."""

import os
import tempfile

import pytest


@pytest.fixture(scope="session")
def temp_config_dir():
    """Create a temporary config directory for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        # Set environment variables for testing
        os.environ["XDG_CONFIG_HOME"] = tmpdir
        os.environ["XDG_DATA_HOME"] = tmpdir
        os.environ["XDG_CACHE_HOME"] = tmpdir
        yield tmpdir


@pytest.fixture
def django_settings():
    """Configure Django settings for testing."""
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "cloudbench.settings.development")

    import django

    django.setup()


@pytest.fixture
def config_manager(temp_config_dir):
    """Get a fresh ConfigManager for testing."""
    from apps.core.config import ConfigManager

    # Reset the singleton
    ConfigManager._instance = None

    return ConfigManager()


@pytest.fixture
def sample_connection():
    """Create a sample GeoServer connection for testing."""
    from apps.core.config import Connection

    return Connection(
        name="Test GeoServer",
        url="http://localhost:8080/geoserver",
        username="admin",
        password="geoserver",
    )


@pytest.fixture
def sample_s3_connection():
    """Create a sample S3 connection for testing."""
    from apps.core.config import S3Connection

    return S3Connection(
        name="Test MinIO",
        endpoint="localhost:9000",
        access_key="minioadmin",
        secret_key="minioadmin",
        use_ssl=False,
        path_style=True,
    )
