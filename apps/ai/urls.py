"""URL configuration for AI app."""

from django.urls import path

from . import views

urlpatterns = [
    # Providers
    path(
        "ai/providers",
        views.AIProvidersView.as_view(),
        name="ai-providers",
    ),
    # Query generation
    path(
        "ai/query",
        views.AIQueryView.as_view(),
        name="ai-query",
    ),
    # Explanation
    path(
        "ai/explain",
        views.AIExplainView.as_view(),
        name="ai-explain",
    ),
    # Execute
    path(
        "ai/execute",
        views.AIExecuteView.as_view(),
        name="ai-execute",
    ),
    # Suggestions
    path(
        "ai/suggest",
        views.AISuggestView.as_view(),
        name="ai-suggest",
    ),
    # Schema context
    path(
        "ai/schema/<str:service_name>",
        views.AISchemaContextView.as_view(),
        name="ai-schema-context",
    ),
]
