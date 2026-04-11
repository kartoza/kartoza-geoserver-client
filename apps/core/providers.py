"""Provider configuration management for Kartoza CloudBench.

Controls which provider types are enabled/disabled at startup.
Uses a separate JSON config file for provider settings.
"""

import json
import os
import threading
from pathlib import Path
from typing import Any

from pydantic import BaseModel, Field

# Provider configuration file name
PROVIDERS_FILE = "providers.json"


class ProviderConfig(BaseModel):
    """Configuration for a single provider type."""

    id: str
    name: str
    description: str
    enabled: bool = True
    experimental: bool = False


class ProvidersConfig(BaseModel):
    """Main providers configuration."""

    providers: list[ProviderConfig] = Field(default_factory=list)

    class Config:
        """Pydantic configuration."""

        extra = "allow"


# Default provider configurations
DEFAULT_PROVIDERS: list[dict[str, Any]] = [
    {
        "id": "geoserver",
        "name": "GeoServer",
        "description": "OGC-compliant geospatial server for publishing and sharing geospatial data",
        "enabled": True,
        "experimental": False,
    },
    {
        "id": "postgres",
        "name": "PostgreSQL",
        "description": "PostgreSQL database connections via pg_service.conf",
        "enabled": True,
        "experimental": False,
    },
    {
        "id": "geonode",
        "name": "GeoNode",
        "description": "Open source geospatial content management system",
        "enabled": True,
        "experimental": False,
    },
    {
        "id": "s3",
        "name": "S3 Storage",
        "description": "S3-compatible object storage (MinIO, AWS S3, etc.)",
        "enabled": False,
        "experimental": True,
    },
    {
        "id": "iceberg",
        "name": "Apache Iceberg",
        "description": "Apache Iceberg data lakehouse tables",
        "enabled": False,
        "experimental": True,
    },
    {
        "id": "qgis",
        "name": "QGIS Projects",
        "description": "Local QGIS project files",
        "enabled": False,
        "experimental": True,
    },
    {
        "id": "qfieldcloud",
        "name": "QFieldCloud",
        "description": "Cloud-based mobile GIS data synchronization",
        "enabled": False,
        "experimental": True,
    },
    {
        "id": "mergin",
        "name": "Mergin Maps",
        "description": "Field data collection and synchronization platform",
        "enabled": False,
        "experimental": True,
    },
]


class ProvidersManager:
    """Thread-safe providers configuration manager.

    Manages provider enablement/disablement settings.
    """

    _instance: "ProvidersManager | None" = None
    _lock = threading.RLock()

    def __new__(cls) -> "ProvidersManager":
        """Singleton pattern for providers manager."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._config = None
                    cls._instance._initialized = False
        return cls._instance

    @property
    def config(self) -> ProvidersConfig:
        """Get the current providers configuration, loading if necessary."""
        with self._lock:
            if self._config is None:
                self._config = self._load()
            return self._config

    def reload(self) -> ProvidersConfig:
        """Force reload providers configuration from disk."""
        with self._lock:
            self._config = self._load()
            return self._config

    def save(self) -> None:
        """Save providers configuration to disk atomically."""
        with self._lock:
            if self._config is None:
                return

            path = self._config_path()
            os.makedirs(os.path.dirname(path), exist_ok=True)

            # Atomic write using temp file
            tmp_path = path + ".tmp"
            with open(tmp_path, "w") as f:
                json.dump(self._config.model_dump(by_alias=True), f, indent=2)

            os.replace(tmp_path, path)

    def _config_path(self) -> str:
        """Get the path to the providers config file."""
        from .config import CONFIG_DIR

        config_home = os.environ.get("XDG_CONFIG_HOME")
        if not config_home:
            config_home = os.path.join(str(Path.home()), ".config")
        return os.path.join(config_home, CONFIG_DIR, PROVIDERS_FILE)

    def _load(self) -> ProvidersConfig:
        """Load providers configuration from disk."""
        path = self._config_path()

        if not os.path.exists(path):
            # Create default config
            config = ProvidersConfig(
                providers=[ProviderConfig(**p) for p in DEFAULT_PROVIDERS]
            )
            self._config = config
            self.save()
            return config

        try:
            with open(path) as f:
                data = json.load(f)

            # Merge with defaults to ensure new providers are added
            loaded_config = ProvidersConfig.model_validate(data)
            loaded_ids = {p.id for p in loaded_config.providers}

            # Add any new default providers that don't exist in saved config
            for default_provider in DEFAULT_PROVIDERS:
                if default_provider["id"] not in loaded_ids:
                    loaded_config.providers.append(
                        ProviderConfig(**default_provider)
                    )

            return loaded_config
        except (json.JSONDecodeError, ValueError):
            # Corrupted config, return default
            return ProvidersConfig(
                providers=[ProviderConfig(**p) for p in DEFAULT_PROVIDERS]
            )

    def get_provider(self, provider_id: str) -> ProviderConfig | None:
        """Get a provider configuration by ID."""
        for provider in self.config.providers:
            if provider.id == provider_id:
                return provider
        return None

    def is_provider_enabled(self, provider_id: str) -> bool:
        """Check if a provider is enabled."""
        provider = self.get_provider(provider_id)
        return provider.enabled if provider else False

    def list_providers(self) -> list[ProviderConfig]:
        """List all provider configurations."""
        return list(self.config.providers)

    def list_enabled_providers(self) -> list[ProviderConfig]:
        """List only enabled provider configurations."""
        return [p for p in self.config.providers if p.enabled]

    def set_provider_enabled(self, provider_id: str, enabled: bool) -> bool:
        """Set the enabled state for a provider."""
        with self._lock:
            for provider in self.config.providers:
                if provider.id == provider_id:
                    provider.enabled = enabled
                    self.save()
                    return True
            return False

    def get_enabled_provider_ids(self) -> set[str]:
        """Get a set of enabled provider IDs for quick lookup."""
        return {p.id for p in self.config.providers if p.enabled}


# Global providers manager instance
def get_providers_manager() -> ProvidersManager:
    """Get the providers manager singleton."""
    return ProvidersManager()


def is_provider_enabled(provider_id: str) -> bool:
    """Check if a provider is enabled."""
    return get_providers_manager().is_provider_enabled(provider_id)
