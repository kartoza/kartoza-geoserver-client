"""URL configuration for Terria app."""

from django.urls import path

from . import views

urlpatterns = [
    # Terria init config
    path(
        "terria/init",
        views.TerriaInitView.as_view(),
        name="terria-init",
    ),
    # Connection catalog
    path(
        "terria/connection/<str:conn_id>",
        views.TerriaConnectionCatalogView.as_view(),
        name="terria-connection-catalog",
    ),
    # Workspace catalog
    path(
        "terria/workspace/<str:conn_id>/<str:workspace>",
        views.TerriaWorkspaceCatalogView.as_view(),
        name="terria-workspace-catalog",
    ),
    # Layer catalog item
    path(
        "terria/layer/<str:conn_id>/<str:workspace>/<str:layer>",
        views.TerriaLayerCatalogView.as_view(),
        name="terria-layer-catalog",
    ),
    # Proxy
    path(
        "terria/proxy",
        views.TerriaProxyView.as_view(),
        name="terria-proxy",
    ),
]
