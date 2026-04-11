"""Base view utilities for GeoServer views."""

from rest_framework import status
from rest_framework.response import Response

from apps.core.exceptions import GeoServerError


def handle_geoserver_error(error: GeoServerError) -> Response:
    """Convert GeoServerError to Response.

    Args:
        error: The GeoServer error

    Returns:
        Response with error details
    """
    return Response(
        {"error": error.message},
        status=error.status_code or status.HTTP_502_BAD_GATEWAY,
    )


def get_recurse_param(request) -> bool:
    """Extract recurse parameter from request.

    Args:
        request: The HTTP request

    Returns:
        Boolean value of recurse parameter
    """
    return request.query_params.get("recurse", "false").lower() == "true"
