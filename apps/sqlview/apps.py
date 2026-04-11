"""Django app configuration for sqlview app."""

from django.apps import AppConfig


class SqlviewConfig(AppConfig):
    """Configuration for the SQL view publishing app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "apps.sqlview"
    verbose_name = "SQL View Publishing"
