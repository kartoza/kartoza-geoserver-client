"""URL configuration for PostgreSQL app."""

from django.urls import path

from . import views

urlpatterns = [
    # PostgreSQL Services
    path(
        "pg/services",
        views.PGServiceListView.as_view(),
        name="pg-service-list",
    ),
    path(
        "pg/services/<str:name>",
        views.PGServiceDetailView.as_view(),
        name="pg-service-detail",
    ),
    path(
        "pg/services/<str:name>/test",
        views.PGServiceTestView.as_view(),
        name="pg-service-test",
    ),
    # Schema and Tables
    path(
        "pg/services/<str:name>/schemas",
        views.PGSchemaListView.as_view(),
        name="pg-schema-list",
    ),
    path(
        "pg/services/<str:name>/schemas/<str:schema>/tables",
        views.PGTableListView.as_view(),
        name="pg-table-list",
    ),
    path(
        "pg/services/<str:name>/schemas/<str:schema>/tables/<str:table>",
        views.PGTableDetailView.as_view(),
        name="pg-table-detail",
    ),
    # Queries
    path(
        "pg/services/<str:name>/query",
        views.PGQueryView.as_view(),
        name="pg-query",
    ),
    # Import
    path(
        "pg/import",
        views.PGImportView.as_view(),
        name="pg-import",
    ),
    path(
        "pg/import/raster",
        views.PGImportRasterView.as_view(),
        name="pg-import-raster",
    ),
    path(
        "pg/import/<str:job_id>",
        views.PGImportStatusView.as_view(),
        name="pg-import-status",
    ),
    # Layer Detection
    path(
        "pg/detect-layers",
        views.PGDetectLayersView.as_view(),
        name="pg-detect-layers",
    ),
]
