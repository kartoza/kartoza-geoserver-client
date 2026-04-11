"""URL configuration for GeoWebCache app."""

from django.urls import path

from . import views

urlpatterns = [
    # GWC Layers
    path(
        "gwc/layers/<str:conn_id>/<str:workspace>",
        views.GWCLayerListView.as_view(),
        name="gwc-layer-list",
    ),
    path(
        "gwc/layers/<str:conn_id>/<str:workspace>/<str:layer>",
        views.GWCLayerDetailView.as_view(),
        name="gwc-layer-detail",
    ),
    # Seeding
    path(
        "gwc/seed/<str:conn_id>/<str:workspace>/<str:layer>",
        views.GWCSeedView.as_view(),
        name="gwc-seed",
    ),
    # Truncating
    path(
        "gwc/truncate/<str:conn_id>/<str:workspace>/<str:layer>",
        views.GWCTruncateView.as_view(),
        name="gwc-truncate",
    ),
    # Grid Sets
    path(
        "gwc/gridsets/<str:conn_id>",
        views.GWCGridSetListView.as_view(),
        name="gwc-gridset-list",
    ),
    path(
        "gwc/gridsets/<str:conn_id>/<str:gridset>",
        views.GWCGridSetDetailView.as_view(),
        name="gwc-gridset-detail",
    ),
    # Disk Quota
    path(
        "gwc/diskquota/<str:conn_id>",
        views.GWCDiskQuotaView.as_view(),
        name="gwc-diskquota",
    ),
    # Mass Truncate
    path(
        "gwc/masstruncate/<str:conn_id>",
        views.GWCMassTruncateView.as_view(),
        name="gwc-masstruncate",
    ),
]
