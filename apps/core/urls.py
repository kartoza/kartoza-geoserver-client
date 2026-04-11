"""URL configuration for core app."""

from django.urls import path

from . import views

urlpatterns = [
    path("settings/", views.SettingsView.as_view(), name="settings"),
    path("providers/", views.ProvidersView.as_view(), name="providers"),
]
