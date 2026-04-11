"""Data store views for GeoServer API."""

from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.exceptions import GeoServerError

from ..client import get_geoserver_client
from .base import get_recurse_param, handle_geoserver_error


class DataStoreListView(APIView):
    """List and create data stores in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all data stores."""
        try:
            client = get_geoserver_client(conn_id)
            datastores = client.list_datastores(workspace)
            return Response(datastores)
        except GeoServerError:
            return Response([])

    def post(self, request, conn_id, workspace):
        """Create a new data store."""
        try:
            client = get_geoserver_client(conn_id)
            name = request.data.get("name")
            connection_params = request.data.get("connectionParameters", {})
            description = request.data.get("description", "")
            enabled = request.data.get("enabled", True)

            if not name:
                return Response(
                    {"error": "name is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            if not connection_params:
                return Response(
                    {"error": "connectionParameters is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.create_datastore(
                workspace, name, connection_params, description, enabled
            )
            return Response(
                {"message": f"Data store {name} created"},
                status=status.HTTP_201_CREATED,
            )
        except GeoServerError as e:
            return handle_geoserver_error(e)


class DataStoreDetailView(APIView):
    """Get, update, or delete a data store."""

    def get(self, request, conn_id, workspace, store):
        """Get data store details."""
        try:
            client = get_geoserver_client(conn_id)
            ds = client.get_datastore(workspace, store)
            return Response({"dataStore": ds})
        except GeoServerError as e:
            return handle_geoserver_error(e)

    def put(self, request, conn_id, workspace, store):
        """Update a data store."""
        try:
            client = get_geoserver_client(conn_id)
            # For updates, we need to rebuild connection parameters
            connection_params = request.data.get("connectionParameters")
            description = request.data.get("description")
            enabled = request.data.get("enabled")

            # Build update payload
            payload = {"dataStore": {"name": store}}
            if connection_params:
                entries = [{"@key": k, "$": v} for k, v in connection_params.items()]
                payload["dataStore"]["connectionParameters"] = {"entry": entries}
            if description is not None:
                payload["dataStore"]["description"] = description
            if enabled is not None:
                payload["dataStore"]["enabled"] = enabled

            client._request(
                "PUT",
                f"/rest/workspaces/{workspace}/datastores/{store}.json",
                json=payload,
            )
            return Response({"message": "Data store updated"})
        except GeoServerError as e:
            return handle_geoserver_error(e)

    def delete(self, request, conn_id, workspace, store):
        """Delete a data store."""
        try:
            client = get_geoserver_client(conn_id)
            recurse = get_recurse_param(request)
            client.delete_datastore(workspace, store, recurse=recurse)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return handle_geoserver_error(e)


class DataStoreAvailableView(APIView):
    """List available (unpublished) feature types in a data store."""

    def get(self, request, conn_id, workspace, store):
        """List available feature types."""
        try:
            client = get_geoserver_client(conn_id)
            available = client.list_available_featuretypes(workspace, store)
            return Response(available)
        except GeoServerError:
            return Response([])
