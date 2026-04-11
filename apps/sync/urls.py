"""URL configuration for sync app."""

from django.urls import path

from . import views

urlpatterns = [
    # Sync configurations
    path(
        "sync/configs",
        views.SyncConfigListView.as_view(),
        name="sync-config-list",
    ),
    path(
        "sync/configs/<str:config_id>",
        views.SyncConfigDetailView.as_view(),
        name="sync-config-detail",
    ),
    # Start sync
    path(
        "sync/start",
        views.SyncStartView.as_view(),
        name="sync-start",
    ),
    # Status
    path(
        "sync/status",
        views.SyncStatusView.as_view(),
        name="sync-status-list",
    ),
    path(
        "sync/status/<str:job_id>",
        views.SyncStatusView.as_view(),
        name="sync-status-detail",
    ),
]
