"""Client managers for thread-safe connection pooling.

Provides cached client instances for GeoServer, S3, and other services.
"""

import threading
from typing import Any

import httpx


class ClientManager:
    """Thread-safe manager for HTTP client instances.

    Caches httpx clients by connection ID for reuse across requests.
    """

    _instance: "ClientManager | None" = None
    _lock = threading.RLock()

    def __new__(cls) -> "ClientManager":
        """Singleton pattern for client manager."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._clients: dict[str, httpx.Client] = {}
                    cls._instance._async_clients: dict[str, httpx.AsyncClient] = {}
        return cls._instance

    def get_client(
        self,
        conn_id: str,
        base_url: str,
        username: str | None = None,
        password: str | None = None,
        **kwargs: Any,
    ) -> httpx.Client:
        """Get or create a synchronous HTTP client.

        Args:
            conn_id: Connection identifier for caching
            base_url: Base URL for the client
            username: Optional username for basic auth
            password: Optional password for basic auth
            **kwargs: Additional arguments passed to httpx.Client

        Returns:
            Cached or new httpx.Client instance
        """
        with self._lock:
            if conn_id not in self._clients:
                auth = None
                if username and password:
                    auth = httpx.BasicAuth(username, password)

                self._clients[conn_id] = httpx.Client(
                    base_url=base_url,
                    auth=auth,
                    timeout=httpx.Timeout(30.0, connect=10.0),
                    follow_redirects=True,
                    **kwargs,
                )
            return self._clients[conn_id]

    def get_async_client(
        self,
        conn_id: str,
        base_url: str,
        username: str | None = None,
        password: str | None = None,
        **kwargs: Any,
    ) -> httpx.AsyncClient:
        """Get or create an asynchronous HTTP client.

        Args:
            conn_id: Connection identifier for caching
            base_url: Base URL for the client
            username: Optional username for basic auth
            password: Optional password for basic auth
            **kwargs: Additional arguments passed to httpx.AsyncClient

        Returns:
            Cached or new httpx.AsyncClient instance
        """
        with self._lock:
            if conn_id not in self._async_clients:
                auth = None
                if username and password:
                    auth = httpx.BasicAuth(username, password)

                self._async_clients[conn_id] = httpx.AsyncClient(
                    base_url=base_url,
                    auth=auth,
                    timeout=httpx.Timeout(30.0, connect=10.0),
                    follow_redirects=True,
                    **kwargs,
                )
            return self._async_clients[conn_id]

    def remove_client(self, conn_id: str) -> None:
        """Remove a cached client."""
        with self._lock:
            if conn_id in self._clients:
                self._clients[conn_id].close()
                del self._clients[conn_id]
            if conn_id in self._async_clients:
                # Note: async client should be closed in async context
                del self._async_clients[conn_id]

    def clear_all(self) -> None:
        """Close and remove all cached clients."""
        with self._lock:
            for client in self._clients.values():
                client.close()
            self._clients.clear()
            self._async_clients.clear()


# Global client manager instance
client_manager = ClientManager()
