"""Tests for the configuration manager."""

import json
import os
from pathlib import Path

import pytest


def test_config_manager_singleton(config_manager):
    """Test that ConfigManager is a singleton."""
    from apps.core.config import ConfigManager

    manager2 = ConfigManager()
    assert config_manager is manager2


def test_config_default_values(config_manager):
    """Test default configuration values."""
    config = config_manager.config

    assert config.connections == []
    assert config.active_connection == ""
    assert config.theme == "default"
    assert config.ping_interval_secs == 60
    assert config.s3_connections == []


def test_add_connection(config_manager, sample_connection):
    """Test adding a connection."""
    config_manager.add_connection(sample_connection)

    assert len(config_manager.config.connections) == 1
    assert config_manager.config.connections[0].name == "Test GeoServer"


def test_get_connection(config_manager, sample_connection):
    """Test getting a connection by ID."""
    config_manager.add_connection(sample_connection)

    conn = config_manager.get_connection(sample_connection.id)
    assert conn is not None
    assert conn.name == "Test GeoServer"


def test_remove_connection(config_manager, sample_connection):
    """Test removing a connection."""
    config_manager.add_connection(sample_connection)
    config_manager.remove_connection(sample_connection.id)

    assert len(config_manager.config.connections) == 0


def test_set_active_connection(config_manager, sample_connection):
    """Test setting the active connection."""
    config_manager.add_connection(sample_connection)
    config_manager.set_active_connection(sample_connection.id)

    assert config_manager.config.active_connection == sample_connection.id
    assert config_manager.get_active_connection() is not None


def test_add_s3_connection(config_manager, sample_s3_connection):
    """Test adding an S3 connection."""
    config_manager.add_s3_connection(sample_s3_connection)

    assert len(config_manager.config.s3_connections) == 1
    assert config_manager.config.s3_connections[0].name == "Test MinIO"


def test_config_save_load(config_manager, sample_connection, temp_config_dir):
    """Test saving and loading configuration."""
    config_manager.add_connection(sample_connection)
    config_manager.save()

    # Create a new manager and verify the connection was saved
    from apps.core.config import ConfigManager

    ConfigManager._instance = None
    new_manager = ConfigManager()

    assert len(new_manager.config.connections) == 1
    assert new_manager.config.connections[0].name == "Test GeoServer"


def test_config_file_location(temp_config_dir):
    """Test that config is stored in correct location."""
    from apps.core.config import ConfigManager

    ConfigManager._instance = None
    manager = ConfigManager()
    manager.save()

    config_path = Path(temp_config_dir) / "kartoza-cloudbench" / "config.json"
    assert config_path.exists()


def test_connection_serialization(sample_connection):
    """Test that connections can be serialized to JSON."""
    data = sample_connection.model_dump()

    assert data["name"] == "Test GeoServer"
    assert data["url"] == "http://localhost:8080/geoserver"
    assert "id" in data
