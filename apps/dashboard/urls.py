"""URL configuration for dashboard app."""

from django.urls import path

from . import views

urlpatterns = [
    path(
        "dashboard",
        views.DashboardView.as_view(),
        name="dashboard",
    ),
    path(
        "dashboard/server",
        views.DashboardServerView.as_view(),
        name="dashboard-server",
    ),
    path(
        "dashboard/connections",
        views.DashboardConnectionsView.as_view(),
        name="dashboard-connections",
    ),
    path(
        "dashboard/geoserver/<str:conn_id>",
        views.DashboardGeoServerView.as_view(),
        name="dashboard-geoserver",
    ),
]
