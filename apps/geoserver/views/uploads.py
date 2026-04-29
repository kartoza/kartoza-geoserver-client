"""File upload views for GeoServer API."""

from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.exceptions import GeoServerError

from ..client import get_geoserver_client
from .base import handle_geoserver_error


class UploadShapefileView(APIView):
    """Upload shapefile to create a data store."""

    def post(self, request, conn_id, workspace):
        """Upload a shapefile ZIP."""
        try:
            client = get_geoserver_client(conn_id, str(request.user.id))
            store_name = request.data.get("name")
            charset = request.data.get("charset", "UTF-8")

            if not store_name:
                return Response(
                    {"error": "name is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            # Get uploaded file
            file = request.FILES.get("file")
            if not file:
                return Response(
                    {"error": "file is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.upload_shapefile(workspace, store_name, file.read(), charset)
            return Response(
                {"message": f"Shapefile uploaded as {store_name}"},
                status=status.HTTP_201_CREATED,
            )
        except GeoServerError as e:
            return handle_geoserver_error(e)


class UploadGeoTiffView(APIView):
    """Upload GeoTIFF to create a coverage store."""

    def post(self, request, conn_id, workspace):
        """Upload a GeoTIFF file."""
        try:
            client = get_geoserver_client(conn_id, str(request.user.id))
            store_name = request.data.get("name")

            if not store_name:
                return Response(
                    {"error": "name is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            # Get uploaded file
            file = request.FILES.get("file")
            if not file:
                return Response(
                    {"error": "file is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.upload_geotiff(workspace, store_name, file.read())
            return Response(
                {"message": f"GeoTIFF uploaded as {store_name}"},
                status=status.HTTP_201_CREATED,
            )
        except GeoServerError as e:
            return handle_geoserver_error(e)


class UploadGeoPackageView(APIView):
    """Upload GeoPackage to create a data store."""

    def post(self, request, conn_id, workspace):
        """Upload a GeoPackage file."""
        try:
            client = get_geoserver_client(conn_id, str(request.user.id))
            store_name = request.data.get("name")

            if not store_name:
                return Response(
                    {"error": "name is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            # Get uploaded file
            file = request.FILES.get("file")
            if not file:
                return Response(
                    {"error": "file is required"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            client.upload_geopackage(workspace, store_name, file.read())
            return Response(
                {"message": f"GeoPackage uploaded as {store_name}"},
                status=status.HTTP_201_CREATED,
            )
        except GeoServerError as e:
            return handle_geoserver_error(e)
