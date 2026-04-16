"""Views for GeoNode integration.

Provides endpoints for:
- GeoNode connection management
- Layer and map listing
- Resource browsing
"""

import uuid

import httpx
from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.config import GeoNodeConnection, get_config
from .client import GeoNodeClient, get_geonode_client
from .utilities import (
    RESOURCE_TYPE_DETAIL_REQUEST_MAP, RESOURCE_TYPE_LIST_REQUEST_MAP
)


class GeoNodeConnectionListView(APIView):
    """List and create GeoNode connections."""

    def get(self, request):
        """List all GeoNode connections."""
        config = get_config(request.user.id)
        connections = config.list_geonode_connections()
        return Response([
            {
                "id": c.id,
                "name": c.name,
                "url": c.url,
                "username": c.username,
            }
            for c in connections
        ])

    def post(self, request):
        """Create a new GeoNode connection."""
        data = request.data
        conn = GeoNodeConnection(
            id=str(uuid.uuid4()),
            name=data.get("name", ""),
            url=data.get("url", ""),
            username=data.get("username", ""),
            password=data.get("password", ""),
            api_key=data.get("apiKey", ""),
        )

        config = get_config(request.user.id)
        config.add_geonode_connection(conn)

        return Response(
            {
                "id": conn.id,
                "name": conn.name,
                "url": conn.url,
            },
            status=status.HTTP_201_CREATED,
        )


class GeoNodeConnectionTestView(APIView):
    """Test GeoNode connection."""

    def post(self, request):
        """Test connection parameters."""
        data = request.data

        client = GeoNodeClient(
            url=data.get("url", ""),
            username=data.get("username"),
            password=data.get("password"),
            api_key=data.get("apiKey"),
        )

        success, message = client.test_connection()

        if success:
            return Response({"status": "success", "message": message})
        return Response(
            {"status": "error", "message": message},
            status=status.HTTP_400_BAD_REQUEST,
        )


class GeoNodeConnectionDetailView(APIView):
    """Get, update, or delete a GeoNode connection."""

    def get(self, request, conn_id):
        """Get connection details."""
        config = get_config(request.user.id)
        conn = config.get_geonode_connection(conn_id)
        if not conn:
            return Response(
                {"error": "Connection not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        return Response({
            "connection": {
                "id": conn.id,
                "name": conn.name,
                "url": conn.url,
                "username": conn.username,
            }
        })

    def put(self, request, conn_id):
        """Update a connection."""
        config = get_config(request.user.id)
        conn = config.get_geonode_connection(conn_id)
        if not conn:
            return Response(
                {"error": "Connection not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        data = request.data
        conn.name = data.get("name", conn.name)
        conn.url = data.get("url", conn.url)
        conn.username = data.get("username", conn.username)
        if "password" in data:
            conn.password = data["password"]
        if "apiKey" in data:
            conn.api_key = data["apiKey"]

        config.update_geonode_connection(conn)

        return Response({"status": "updated"})

    def delete(self, request, conn_id):
        """Delete a connection."""
        config = get_config(request.user.id)
        if not config.delete_geonode_connection(conn_id):
            return Response(
                {"error": "Connection not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        return Response(status=status.HTTP_204_NO_CONTENT)


class GeoNodeResourceListView(APIView):
    """List layers for a connection."""

    def get(self, request, conn_id, resource_type):
        """List all layers."""
        page = int(request.query_params.get("page", 1))
        page_size = int(request.query_params.get("pageSize", 20))
        category = request.query_params.get("category")
        owner = request.query_params.get("owner")

        try:
            client = get_geonode_client(conn_id, str(request.user.id))
            result = client.list_resources(
                resource_type=resource_type,
                page=page,
                page_size=page_size,
                category=category,
                owner=owner,
            )
            return Response(result)
        except ValueError as e:
            return Response(
                {"error": str(e)},
                status=status.HTTP_404_NOT_FOUND,
            )
        except Exception as e:
            return Response(
                {"error": str(e)},
                status=status.HTTP_502_BAD_GATEWAY,
            )


class GeoNodeResourceDetailView(APIView):
    """Get layer details."""

    def get(self, request, conn_id, resource_type, resource_id):
        """Get layer information."""
        resource_type = RESOURCE_TYPE_DETAIL_REQUEST_MAP.get(
            resource_type, resource_type
        )
        try:
            client = get_geonode_client(conn_id, str(request.user.id))
            resource = client.get_resource(
                resource_type, int(resource_id)
            )
            return Response(
                {resource_type.rstrip("s"): resource.to_dict()}
            )
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                return Response(
                    {"error": "Resource not found"},
                    status=status.HTTP_404_NOT_FOUND,
                )
            return Response(
                {"error": str(e)},
                status=status.HTTP_502_BAD_GATEWAY,
            )
        except Exception as e:
            return Response(
                {"error": str(e)},
                status=status.HTTP_502_BAD_GATEWAY,
            )


class GeoNodeCategoryListView(APIView):
    """List resource categories."""

    def get(self, request, conn_id):
        """List all categories."""
        try:
            client = get_geonode_client(conn_id, str(request.user.id))
            categories = client.list_categories()
            return Response({"categories": categories})
        except ValueError as e:
            return Response(
                {"error": str(e)},
                status=status.HTTP_404_NOT_FOUND,
            )
        except Exception as e:
            return Response(
                {"error": str(e)},
                status=status.HTTP_502_BAD_GATEWAY,
            )
