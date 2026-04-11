"""Django app configuration for upload app."""

from django.apps import AppConfig


class UploadConfig(AppConfig):
    """Configuration for the chunked file upload app."""

    default_auto_field = "django.db.models.BigAutoField"
    name = "apps.upload"
    verbose_name = "Chunked File Uploads"
