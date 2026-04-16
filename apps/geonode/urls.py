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
    # Categories
    path(
        "geonode/connections/<str:conn_id>/categories",
        views.GeoNodeCategoryListView.as_view(),
        name="geonode-category-list",
    ),
    # Resources
    path(
        "geonode/connections/<str:conn_id>/<str:resource_type>",
        views.GeoNodeResourceListView.as_view(),
        name="geonode-resource-list",
    ),
    path(
        "geonode/connections/<str:conn_id>/<str:resource_type>/<str:resource_id>",
        views.GeoNodeResourceDetailView.as_view(),
        name="geonode-resource-detail",
    ),
]
