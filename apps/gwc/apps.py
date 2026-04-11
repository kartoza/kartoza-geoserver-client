"""Django app configuration for GeoWebCache app."""

from django.apps import AppConfig


class GwcConfig(AppConfig):
    """Configuration for the GeoWebCache app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "apps.gwc"
    verbose_name = "GeoWebCache Management"
