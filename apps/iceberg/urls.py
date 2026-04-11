"""URL configuration for Apache Iceberg app."""

from django.urls import path

from . import views

urlpatterns = [
    # Iceberg Connections
    path(
        "iceberg/connections",
        views.IcebergConnectionListView.as_view(),
        name="iceberg-connection-list",
    ),
    path(
        "iceberg/connections/test",
        views.IcebergConnectionTestView.as_view(),
        name="iceberg-connection-test",
    ),
    path(
        "iceberg/connections/<str:conn_id>",
        views.IcebergConnectionDetailView.as_view(),
        name="iceberg-connection-detail",
    ),
    # Config
    path(
        "iceberg/connections/<str:conn_id>/config",
        views.IcebergConfigView.as_view(),
        name="iceberg-config",
    ),
    # Namespaces
    path(
        "iceberg/connections/<str:conn_id>/namespaces",
        views.IcebergNamespaceListView.as_view(),
        name="iceberg-namespace-list",
    ),
    path(
        "iceberg/connections/<str:conn_id>/namespaces/<str:namespace>",
        views.IcebergNamespaceDetailView.as_view(),
        name="iceberg-namespace-detail",
    ),
    # Tables
    path(
        "iceberg/connections/<str:conn_id>/namespaces/<str:namespace>/tables",
        views.IcebergTableListView.as_view(),
        name="iceberg-table-list",
    ),
    path(
        "iceberg/connections/<str:conn_id>/namespaces/<str:namespace>/tables/<str:table>",
        views.IcebergTableDetailView.as_view(),
        name="iceberg-table-detail",
    ),
    path(
        "iceberg/connections/<str:conn_id>/namespaces/<str:namespace>/tables/<str:table>/metadata",
        views.IcebergTableMetadataView.as_view(),
        name="iceberg-table-metadata",
    ),
]
