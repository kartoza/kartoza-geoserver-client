"""URL configuration for bridge app."""

from django.urls import path

from . import views

urlpatterns = [
    # Create PostGIS store from pg_service
    path(
        "bridge/<str:conn_id>/postgis-store",
        views.BridgePostGISStoreView.as_view(),
        name="bridge-postgis-store",
    ),
    # List publishable tables
    path(
        "bridge/<str:conn_id>/<str:workspace>/<str:store>/tables",
        views.BridgePublishableTablesView.as_view(),
        name="bridge-publishable-tables",
    ),
    # Publish single table as layer
    path(
        "bridge/<str:conn_id>/<str:workspace>/<str:store>/publish",
        views.BridgePublishLayerView.as_view(),
        name="bridge-publish-layer",
    ),
    # Batch publish tables as layers
    path(
        "bridge/<str:conn_id>/<str:workspace>/<str:store>/publish/batch",
        views.BridgeBatchPublishView.as_view(),
        name="bridge-batch-publish",
    ),
]
