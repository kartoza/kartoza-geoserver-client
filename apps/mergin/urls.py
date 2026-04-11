"""URL configuration for Mergin Maps app."""

from django.urls import path

from . import views

urlpatterns = [
    # Mergin Connections
    path(
        "mergin/connections",
        views.MerginConnectionListView.as_view(),
        name="mergin-connection-list",
    ),
    path(
        "mergin/connections/test",
        views.MerginConnectionTestView.as_view(),
        name="mergin-connection-test",
    ),
    path(
        "mergin/connections/<str:conn_id>",
        views.MerginConnectionDetailView.as_view(),
        name="mergin-connection-detail",
    ),
    # Projects
    path(
        "mergin/connections/<str:conn_id>/projects",
        views.MerginProjectListView.as_view(),
        name="mergin-project-list",
    ),
    path(
        "mergin/connections/<str:conn_id>/projects/<str:namespace>/<str:name>",
        views.MerginProjectDetailView.as_view(),
        name="mergin-project-detail",
    ),
    path(
        "mergin/connections/<str:conn_id>/projects/<str:namespace>/<str:name>/files",
        views.MerginProjectFilesView.as_view(),
        name="mergin-project-files",
    ),
    path(
        "mergin/connections/<str:conn_id>/projects/<str:namespace>/<str:name>/versions",
        views.MerginProjectVersionsView.as_view(),
        name="mergin-project-versions",
    ),
]
