"""URL configuration for search app."""

from django.urls import path

from . import views

urlpatterns = [
    path(
        "search",
        views.SearchView.as_view(),
        name="search",
    ),
    path(
        "search/suggestions",
        views.SearchSuggestionsView.as_view(),
        name="search-suggestions",
    ),
]
