"""GeoNode API client.

Provides a client for interacting with GeoNode REST API
for managing geospatial data catalog and services.
"""

import threading
from dataclasses import dataclass
from typing import Any

import httpx

from apps.core.config import get_config


@dataclass
class GeoNodeLayer:
    """GeoNode layer information."""

    pk: int
    uuid: str
    name: str
    title: str
    abstract: str = ""
    category: str = ""
    owner: str = ""
    date: str = ""
    thumbnail_url: str = ""
    detail_url: str = ""

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        return {
            "pk": self.pk,
            "uuid": self.uuid,
            "name": self.name,
            "title": self.title,
            "abstract": self.abstract,
            "category": self.category,
            "owner": self.owner,
            "date": self.date,
            "thumbnailUrl": self.thumbnail_url,
            "detailUrl": self.detail_url,
        }


@dataclass
class GeoNodeMap:
    """GeoNode map information."""

    pk: int
    uuid: str
    title: str
    abstract: str = ""
    owner: str = ""
    date: str = ""
    thumbnail_url: str = ""
    detail_url: str = ""

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        return {
            "pk": self.pk,
            "uuid": self.uuid,
            "title": self.title,
            "abstract": self.abstract,
            "owner": self.owner,
            "date": self.date,
            "thumbnailUrl": self.thumbnail_url,
            "detailUrl": self.detail_url,
        }


class GeoNodeClient:
    """Client for GeoNode REST API."""

    def __init__(
        self,
        url: str,
        username: str | None = None,
        password: str | None = None,
        api_key: str | None = None,
    ):
        """Initialize GeoNode client.

        Args:
            url: GeoNode server URL
            username: Username for authentication
            password: Password for authentication
            api_key: API key (alternative to username/password)
        """
        self.url = url.rstrip("/")
        self.username = username
        self.password = password
        self.api_key = api_key

        # Create HTTP client
        headers = {}
        if api_key:
            headers["Authorization"] = f"ApiKey {api_key}"

        auth = None
        if username and password:
            auth = httpx.BasicAuth(username, password)

        self.client = httpx.Client(
            base_url=f"{self.url}/api/v2",
            timeout=30.0,
            headers=headers,
            auth=auth,
        )

    def test_connection(self) -> tuple[bool, str]:
        """Test the connection.

        Returns:
            Tuple of (success, message)
        """
        try:
            response = self.client.get("/")
            response.raise_for_status()

            return True, f"Connected to GeoNode at {self.url}"
        except httpx.HTTPStatusError as e:
            return False, f"HTTP error: {e.response.status_code}"
        except Exception as e:
            return False, str(e)

    def list_layers(
        self,
        page: int = 1,
        page_size: int = 20,
        category: str | None = None,
        owner: str | None = None,
    ) -> dict[str, Any]:
        """List layers.

        Args:
            page: Page number
            page_size: Items per page
            category: Filter by category
            owner: Filter by owner

        Returns:
            Dictionary with layers and pagination info
        """
        params = {
            "page": page,
            "page_size": page_size,
        }
        if category:
            params["filter[category__identifier]"] = category
        if owner:
            params["filter[owner__username]"] = owner

        response = self.client.get("/datasets", params=params)
        response.raise_for_status()

        data = response.json()
        layers = []

        for item in data.get("datasets", []):
            layers.append(
                GeoNodeLayer(
                    pk=item.get("pk", 0),
                    uuid=item.get("uuid", ""),
                    name=item.get("name", ""),
                    title=item.get("title", ""),
                    abstract=item.get("abstract", ""),
                    category=item.get("category", {}).get("identifier", ""),
                    owner=item.get("owner", {}).get("username", ""),
                    date=item.get("date", ""),
                    thumbnail_url=item.get("thumbnail_url", ""),
                    detail_url=item.get("detail_url", ""),
                )
            )

        return {
            "layers": [l.to_dict() for l in layers],
            "total": data.get("total", 0),
            "page": data.get("page", 1),
            "pageSize": data.get("page_size", 20),
        }

    def get_layer(self, layer_id: int) -> GeoNodeLayer | None:
        """Get a specific layer.

        Args:
            layer_id: Layer primary key

        Returns:
            GeoNodeLayer or None
        """
        response = self.client.get(f"/datasets/{layer_id}")

        if response.status_code == 404:
            return None

        response.raise_for_status()
        item = response.json().get("dataset", {})

        return GeoNodeLayer(
            pk=item.get("pk", 0),
            uuid=item.get("uuid", ""),
            name=item.get("name", ""),
            title=item.get("title", ""),
            abstract=item.get("abstract", ""),
            category=item.get("category", {}).get("identifier", ""),
            owner=item.get("owner", {}).get("username", ""),
            date=item.get("date", ""),
            thumbnail_url=item.get("thumbnail_url", ""),
            detail_url=item.get("detail_url", ""),
        )

    def list_maps(
        self,
        page: int = 1,
        page_size: int = 20,
        owner: str | None = None,
    ) -> dict[str, Any]:
        """List maps.

        Args:
            page: Page number
            page_size: Items per page
            owner: Filter by owner

        Returns:
            Dictionary with maps and pagination info
        """
        params = {
            "page": page,
            "page_size": page_size,
        }
        if owner:
            params["filter[owner__username]"] = owner

        response = self.client.get("/maps", params=params)
        response.raise_for_status()

        data = response.json()
        maps = []

        for item in data.get("maps", []):
            maps.append(
                GeoNodeMap(
                    pk=item.get("pk", 0),
                    uuid=item.get("uuid", ""),
                    title=item.get("title", ""),
                    abstract=item.get("abstract", ""),
                    owner=item.get("owner", {}).get("username", ""),
                    date=item.get("date", ""),
                    thumbnail_url=item.get("thumbnail_url", ""),
                    detail_url=item.get("detail_url", ""),
                )
            )

        return {
            "maps": [m.to_dict() for m in maps],
            "total": data.get("total", 0),
            "page": data.get("page", 1),
            "pageSize": data.get("page_size", 20),
        }

    def get_map(self, map_id: int) -> GeoNodeMap | None:
        """Get a specific map.

        Args:
            map_id: Map primary key

        Returns:
            GeoNodeMap or None
        """
        response = self.client.get(f"/maps/{map_id}")

        if response.status_code == 404:
            return None

        response.raise_for_status()
        item = response.json().get("map", {})

        return GeoNodeMap(
            pk=item.get("pk", 0),
            uuid=item.get("uuid", ""),
            title=item.get("title", ""),
            abstract=item.get("abstract", ""),
            owner=item.get("owner", {}).get("username", ""),
            date=item.get("date", ""),
            thumbnail_url=item.get("thumbnail_url", ""),
            detail_url=item.get("detail_url", ""),
        )

    def list_categories(self) -> list[dict[str, Any]]:
        """List resource categories.

        Returns:
            List of categories
        """
        response = self.client.get("/categories")
        response.raise_for_status()

        return response.json().get("categories", [])

    def list_users(
        self,
        page: int = 1,
        page_size: int = 20,
    ) -> dict[str, Any]:
        """List users.

        Args:
            page: Page number
            page_size: Items per page

        Returns:
            Dictionary with users and pagination info
        """
        params = {
            "page": page,
            "page_size": page_size,
        }

        response = self.client.get("/users", params=params)
        response.raise_for_status()

        return response.json()


class GeoNodeClientManager:
    """Thread-safe manager for GeoNode clients."""

    _instance: "GeoNodeClientManager | None" = None
    _lock = threading.RLock()

    def __new__(cls) -> "GeoNodeClientManager":
        """Ensure singleton instance."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._clients: dict[str, GeoNodeClient] = {}
        return cls._instance

    def get_client(self, connection_id: str) -> GeoNodeClient:
        """Get or create a GeoNode client.

        Args:
            connection_id: Connection ID

        Returns:
            GeoNodeClient instance

        Raises:
            ValueError: If connection not found
        """
        with self._lock:
            if connection_id in self._clients:
                return self._clients[connection_id]

            config = get_config()
            conn = config.get_geonode_connection(connection_id)
            if not conn:
                raise ValueError(f"GeoNode connection not found: {connection_id}")

            client = GeoNodeClient(
                url=conn.url,
                username=conn.username,
                password=conn.password,
                api_key=conn.api_key,
            )

            self._clients[connection_id] = client
            return client

    def remove_client(self, connection_id: str) -> None:
        """Remove a cached client."""
        with self._lock:
            self._clients.pop(connection_id, None)


def get_geonode_client(connection_id: str) -> GeoNodeClient:
    """Get a GeoNode client for a connection."""
    manager = GeoNodeClientManager()
    return manager.get_client(connection_id)
