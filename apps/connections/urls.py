"""URL configuration for connections app."""

from django.urls import path

from . import views

urlpatterns = [
    path("connections", views.ConnectionListView.as_view(), name="connection-list"),
    path("connections/test", views.ConnectionTestView.as_view(), name="connection-test"),
    path(
        "connections/<str:conn_id>",
        views.ConnectionDetailView.as_view(),
        name="connection-detail",
    ),
    path(
        "connections/<str:conn_id>/test",
        views.ConnectionTestExistingView.as_view(),
        name="connection-test-existing",
    ),
    path(
        "connections/<str:conn_id>/info",
        views.ConnectionInfoView.as_view(),
        name="connection-info",
    ),
]
