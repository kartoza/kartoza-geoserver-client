"""Views for GeoServer REST API operations."""

from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.exceptions import GeoServerError

from .client import get_geoserver_client


# === Workspaces ===


class WorkspaceListView(APIView):
    """List workspaces for a connection."""

    def get(self, request, conn_id):
        """List all workspaces."""
        try:
            client = get_geoserver_client(conn_id)
            workspaces = client.list_workspaces()
            return Response(workspaces)
        except GeoServerError:
            # Return empty array on error to prevent frontend crashes
            return Response([])


class WorkspaceDetailView(APIView):
    """Get workspace details."""

    def get(self, request, conn_id, workspace):
        """Get workspace details."""
        try:
            client = get_geoserver_client(conn_id)
            ws = client.get_workspace(workspace)
            return Response({"workspace": ws})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Data Stores ===


class DataStoreListView(APIView):
    """List data stores in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all data stores."""
        try:
            client = get_geoserver_client(conn_id)
            datastores = client.list_datastores(workspace)
            return Response(datastores)
        except GeoServerError:
            return Response([])


class DataStoreDetailView(APIView):
    """Get, update, or delete a data store."""

    def get(self, request, conn_id, workspace, store):
        """Get data store details."""
        try:
            client = get_geoserver_client(conn_id)
            ds = client.get_datastore(workspace, store)
            return Response({"dataStore": ds})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def delete(self, request, conn_id, workspace, store):
        """Delete a data store."""
        try:
            client = get_geoserver_client(conn_id)
            recurse = request.query_params.get("recurse", "false").lower() == "true"
            client.delete_datastore(workspace, store, recurse=recurse)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


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


# === Coverage Stores ===


class CoverageStoreListView(APIView):
    """List coverage stores in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all coverage stores."""
        try:
            client = get_geoserver_client(conn_id)
            stores = client.list_coveragestores(workspace)
            return Response(stores)
        except GeoServerError:
            return Response([])


class CoverageStoreDetailView(APIView):
    """Get, update, or delete a coverage store."""

    def get(self, request, conn_id, workspace, store):
        """Get coverage store details."""
        try:
            client = get_geoserver_client(conn_id)
            cs = client.get_coveragestore(workspace, store)
            return Response({"coverageStore": cs})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def delete(self, request, conn_id, workspace, store):
        """Delete a coverage store."""
        try:
            client = get_geoserver_client(conn_id)
            recurse = request.query_params.get("recurse", "false").lower() == "true"
            client.delete_coveragestore(workspace, store, recurse=recurse)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Feature Types ===


class FeatureTypeListView(APIView):
    """List feature types in a data store."""

    def get(self, request, conn_id, workspace, store):
        """List all feature types."""
        try:
            client = get_geoserver_client(conn_id)
            featuretypes = client.list_featuretypes(workspace, store)
            return Response(featuretypes)
        except GeoServerError:
            return Response([])

    def post(self, request, conn_id, workspace, store):
        """Publish a feature type."""
        try:
            client = get_geoserver_client(conn_id)
            name = request.data.get("name")
            native_name = request.data.get("nativeName", name)
            title = request.data.get("title", name)
            srs = request.data.get("srs", "EPSG:4326")

            if not name:
                return Response(
                    {"error": "name is required"}, status=status.HTTP_400_BAD_REQUEST
                )

            client.create_featuretype(workspace, store, name, native_name, title, srs)
            return Response({"message": f"Feature type {name} published"}, status=status.HTTP_201_CREATED)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


class FeatureTypeDetailView(APIView):
    """Get or delete a feature type."""

    def get(self, request, conn_id, workspace, store, featuretype):
        """Get feature type details."""
        try:
            client = get_geoserver_client(conn_id)
            ft = client.get_featuretype(workspace, store, featuretype)
            return Response({"featureType": ft})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def delete(self, request, conn_id, workspace, store, featuretype):
        """Delete a feature type."""
        try:
            client = get_geoserver_client(conn_id)
            recurse = request.query_params.get("recurse", "false").lower() == "true"
            client.delete_featuretype(workspace, store, featuretype, recurse=recurse)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Coverages ===


class CoverageListView(APIView):
    """List coverages in a coverage store."""

    def get(self, request, conn_id, workspace, store):
        """List all coverages."""
        try:
            client = get_geoserver_client(conn_id)
            coverages = client.list_coverages(workspace, store)
            return Response(coverages)
        except GeoServerError:
            return Response([])


class CoverageDetailView(APIView):
    """Get coverage details."""

    def get(self, request, conn_id, workspace, store, coverage):
        """Get coverage details."""
        try:
            client = get_geoserver_client(conn_id)
            cov = client.get_coverage(workspace, store, coverage)
            return Response({"coverage": cov})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Layers ===


class LayerListView(APIView):
    """List layers in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all layers."""
        try:
            client = get_geoserver_client(conn_id)
            layers = client.list_layers(workspace)
            return Response(layers)
        except GeoServerError:
            return Response([])


class LayerDetailView(APIView):
    """Get, update, or delete a layer."""

    def get(self, request, conn_id, workspace, layer):
        """Get layer details."""
        try:
            client = get_geoserver_client(conn_id)
            lyr = client.get_layer(workspace, layer)
            return Response({"layer": lyr})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def put(self, request, conn_id, workspace, layer):
        """Update layer properties."""
        try:
            client = get_geoserver_client(conn_id)
            enabled = request.data.get("enabled")
            advertised = request.data.get("advertised")
            queryable = request.data.get("queryable")
            default_style = request.data.get("defaultStyle")

            client.update_layer(
                workspace,
                layer,
                enabled=enabled,
                advertised=advertised,
                queryable=queryable,
                default_style=default_style,
            )
            return Response({"message": "Layer updated"})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def delete(self, request, conn_id, workspace, layer):
        """Delete a layer."""
        try:
            client = get_geoserver_client(conn_id)
            recurse = request.query_params.get("recurse", "false").lower() == "true"
            client.delete_layer(workspace, layer, recurse=recurse)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


class LayerCountView(APIView):
    """Get feature count for a layer."""

    def get(self, request, conn_id, workspace, layer):
        """Get layer feature count."""
        try:
            client = get_geoserver_client(conn_id)
            count = client.get_layer_feature_count(workspace, layer)
            return Response({"count": count})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


class LayerMetadataView(APIView):
    """Get layer metadata including bounding box."""

    def get(self, request, conn_id, workspace, layer):
        """Get layer metadata."""
        try:
            client = get_geoserver_client(conn_id)
            metadata = client.get_layer_metadata(workspace, layer)
            return Response(metadata)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


class LayerStylesView(APIView):
    """Get or update layer style associations."""

    def get(self, request, conn_id, workspace, layer):
        """Get layer styles."""
        try:
            client = get_geoserver_client(conn_id)
            styles = client.get_layer_styles(workspace, layer)
            return Response(styles)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def put(self, request, conn_id, workspace, layer):
        """Update layer styles."""
        try:
            client = get_geoserver_client(conn_id)
            default_style = request.data.get("defaultStyle")
            additional_styles = request.data.get("additionalStyles", [])

            if not default_style:
                return Response(
                    {"error": "defaultStyle is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.update_layer_styles(workspace, layer, default_style, additional_styles)
            return Response({"message": "Styles updated"})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Styles ===


class StyleListView(APIView):
    """List styles in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all styles."""
        try:
            client = get_geoserver_client(conn_id)
            styles = client.list_styles(workspace)
            return Response(styles)
        except GeoServerError:
            return Response([])

    def post(self, request, conn_id, workspace):
        """Create a new style."""
        try:
            client = get_geoserver_client(conn_id)
            name = request.data.get("name")
            content = request.data.get("content")
            style_format = request.data.get("format", "sld")

            if not name or not content:
                return Response(
                    {"error": "name and content are required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.create_style(name, content, style_format, workspace)
            return Response({"message": f"Style {name} created"}, status=status.HTTP_201_CREATED)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


class StyleDetailView(APIView):
    """Get, update, or delete a style."""

    def get(self, request, conn_id, workspace, style):
        """Get style details and content."""
        try:
            client = get_geoserver_client(conn_id)
            style_info = client.get_style(style, workspace)
            content, style_format = client.get_style_content(style, workspace)
            return Response({
                "style": style_info,
                "content": content,
                "format": style_format,
            })
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def put(self, request, conn_id, workspace, style):
        """Update style content."""
        try:
            client = get_geoserver_client(conn_id)
            content = request.data.get("content")
            style_format = request.data.get("format", "sld")

            if not content:
                return Response(
                    {"error": "content is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.update_style_content(style, content, style_format, workspace)
            return Response({"message": "Style updated"})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )

    def delete(self, request, conn_id, workspace, style):
        """Delete a style."""
        try:
            client = get_geoserver_client(conn_id)
            purge = request.query_params.get("purge", "false").lower() == "true"
            client.delete_style(style, workspace, purge=purge)
            return Response(status=status.HTTP_204_NO_CONTENT)
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )


# === Layer Groups ===


class LayerGroupListView(APIView):
    """List layer groups in a workspace."""

    def get(self, request, conn_id, workspace):
        """List all layer groups."""
        try:
            client = get_geoserver_client(conn_id)
            groups = client.list_layergroups(workspace)
            return Response(groups)
        except GeoServerError:
            return Response([])


class LayerGroupDetailView(APIView):
    """Get layer group details."""

    def get(self, request, conn_id, workspace, layergroup):
        """Get layer group details."""
        try:
            client = get_geoserver_client(conn_id)
            group = client.get_layergroup(layergroup, workspace)
            return Response({"layerGroup": group})
        except GeoServerError as e:
            return Response(
                {"error": e.message}, status=e.status_code or status.HTTP_502_BAD_GATEWAY
            )
