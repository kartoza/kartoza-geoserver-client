"""Views for core app - settings endpoint."""

from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from .config import config_manager


class SettingsView(APIView):
    """API endpoint for application settings."""

    def get(self, request):
        """Get application settings.

        Returns theme, ping interval, and other app-wide settings.
        """
        config = config_manager.config
        return Response(
            {
                "theme": config.theme,
                "pingIntervalSecs": config.ping_interval_secs,
                "lastLocalPath": config.last_local_path,
            }
        )

    def put(self, request):
        """Update application settings.

        Expected body:
        {
            "theme": "default",
            "pingIntervalSecs": 60,
            "lastLocalPath": "/path/to/dir"
        }
        """
        data = request.data

        if "theme" in data:
            config_manager.config.theme = data["theme"]

        if "pingIntervalSecs" in data:
            interval = int(data["pingIntervalSecs"])
            # Clamp to valid range
            interval = max(10, min(600, interval))
            config_manager.config.ping_interval_secs = interval

        if "lastLocalPath" in data:
            config_manager.config.last_local_path = data["lastLocalPath"]

        config_manager.save()

        return Response(
            {
                "theme": config_manager.config.theme,
                "pingIntervalSecs": config_manager.config.ping_interval_secs,
                "lastLocalPath": config_manager.config.last_local_path,
            }
        )
