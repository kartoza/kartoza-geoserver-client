"""URL configuration for GeoNode app."""

from django.urls import path

from . import views

urlpatterns = [
    # GeoNode Connections
    path(
        "geonode/connections",
        views.GeoNodeConnectionListView.as_view(),
        name="geonode-connection-list",
    ),
    path(
        "geonode/connections/test",
        views.GeoNodeConnectionTestView.as_view(),
        name="geonode-connection-test",
    ),
    path(
        "geonode/connections/<str:conn_id>",
        views.GeoNodeConnectionDetailView.as_view(),
        name="geonode-connection-detail",
    ),
    # Layers
    path(
        "geonode/connections/<str:conn_id>/layers",
        views.GeoNodeLayerListView.as_view(),
        name="geonode-layer-list",
    ),
    path(
        "geonode/connections/<str:conn_id>/layers/<str:layer_id>",
        views.GeoNodeLayerDetailView.as_view(),
        name="geonode-layer-detail",
    ),
    # Maps
    path(
        "geonode/connections/<str:conn_id>/maps",
        views.GeoNodeMapListView.as_view(),
        name="geonode-map-list",
    ),
    path(
        "geonode/connections/<str:conn_id>/maps/<str:map_id>",
        views.GeoNodeMapDetailView.as_view(),
        name="geonode-map-detail",
    ),
    # Categories
    path(
        "geonode/connections/<str:conn_id>/categories",
        views.GeoNodeCategoryListView.as_view(),
        name="geonode-category-list",
    ),
]
