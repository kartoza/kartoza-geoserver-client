"""URL configuration for QGIS app."""

from django.urls import path

from . import views

urlpatterns = [
    # QGIS Projects
    path(
        "qgis/projects",
        views.QGISProjectListView.as_view(),
        name="qgis-project-list",
    ),
    path(
        "qgis/projects/<str:project_id>",
        views.QGISProjectDetailView.as_view(),
        name="qgis-project-detail",
    ),
    # SQL View Publishing
    path(
        "sqlview",
        views.SQLViewPublishView.as_view(),
        name="sqlview-publish",
    ),
    path(
        "sqlview/validate",
        views.SQLViewValidateView.as_view(),
        name="sqlview-validate",
    ),
]
