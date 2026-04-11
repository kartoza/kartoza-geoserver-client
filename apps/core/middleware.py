"""Custom middleware for Kartoza CloudBench."""

from django.conf import settings


class COOPCOEPMiddleware:
    """Middleware to add Cross-Origin isolation headers.

    Required for SharedArrayBuffer support needed by QGIS-js WebAssembly.
    Sets Cross-Origin-Opener-Policy and Cross-Origin-Embedder-Policy headers.
    """

    def __init__(self, get_response):
        """Initialize middleware."""
        self.get_response = get_response

    def __call__(self, request):
        """Process request and add COOP/COEP headers to response."""
        response = self.get_response(request)

        # Only add headers if enabled in settings
        if getattr(settings, "COOP_COEP_ENABLED", True):
            # Cross-Origin-Opener-Policy: same-origin
            # Required for SharedArrayBuffer
            response["Cross-Origin-Opener-Policy"] = "same-origin"

            # Cross-Origin-Embedder-Policy: require-corp
            # Required for SharedArrayBuffer
            response["Cross-Origin-Embedder-Policy"] = "require-corp"

        return response
