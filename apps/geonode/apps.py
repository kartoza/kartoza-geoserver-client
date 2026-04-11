"""Django app configuration for GeoNode app."""

from django.apps import AppConfig


class GeonodeConfig(AppConfig):
    """Configuration for the GeoNode integration app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "apps.geonode"
    verbose_name = "GeoNode Integration"
