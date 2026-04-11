"""URL configuration for QFieldCloud app."""

from django.urls import path

from . import views

urlpatterns = [
    # QFieldCloud Connections
    path(
        "qfieldcloud/connections",
        views.QFieldCloudConnectionListView.as_view(),
        name="qfieldcloud-connection-list",
    ),
    path(
        "qfieldcloud/connections/test",
        views.QFieldCloudConnectionTestView.as_view(),
        name="qfieldcloud-connection-test",
    ),
    path(
        "qfieldcloud/connections/<str:conn_id>",
        views.QFieldCloudConnectionDetailView.as_view(),
        name="qfieldcloud-connection-detail",
    ),
    # Projects
    path(
        "qfieldcloud/connections/<str:conn_id>/projects",
        views.QFieldCloudProjectListView.as_view(),
        name="qfieldcloud-project-list",
    ),
    path(
        "qfieldcloud/connections/<str:conn_id>/projects/<str:project_id>",
        views.QFieldCloudProjectDetailView.as_view(),
        name="qfieldcloud-project-detail",
    ),
    path(
        "qfieldcloud/connections/<str:conn_id>/projects/<str:project_id>/files",
        views.QFieldCloudProjectFilesView.as_view(),
        name="qfieldcloud-project-files",
    ),
    path(
        "qfieldcloud/connections/<str:conn_id>/projects/<str:project_id>/status",
        views.QFieldCloudProjectStatusView.as_view(),
        name="qfieldcloud-project-status",
    ),
]
