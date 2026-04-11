"""Serializers for connections app."""

from rest_framework import serializers

from apps.core.config import Connection


class ConnectionSerializer(serializers.Serializer):
    """Serializer for GeoServer connections."""

    id = serializers.CharField(read_only=True)
    name = serializers.CharField(max_length=255)
    url = serializers.URLField()
    username = serializers.CharField(max_length=255)
    password = serializers.CharField(max_length=255, write_only=True)
    is_active = serializers.BooleanField(default=False)

    def create(self, validated_data):
        """Create a new connection."""
        return Connection(**validated_data)

    def update(self, instance, validated_data):
        """Update an existing connection."""
        for key, value in validated_data.items():
            setattr(instance, key, value)
        return instance


class ConnectionTestSerializer(serializers.Serializer):
    """Serializer for connection test requests."""

    url = serializers.URLField()
    username = serializers.CharField(max_length=255)
    password = serializers.CharField(max_length=255)


class ConnectionResponseSerializer(serializers.Serializer):
    """Serializer for connection responses (includes password masked)."""

    id = serializers.CharField()
    name = serializers.CharField()
    url = serializers.URLField()
    username = serializers.CharField()
    is_active = serializers.BooleanField()

    # Don't include password in responses
