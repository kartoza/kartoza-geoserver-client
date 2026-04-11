"""Views for chunked file upload.

Provides endpoints for:
- Initializing upload sessions
- Uploading file chunks
- Completing uploads
- Tracking progress
- Canceling uploads
"""

import hashlib
import os
import shutil
import threading
import uuid
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path

from django.conf import settings
from rest_framework import status
from rest_framework.parsers import FormParser, MultiPartParser
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.config import get_cache_dir
from apps.core.exceptions import UploadError
from apps.geoserver.client import get_geoserver_client


@dataclass
class UploadSession:
    """Represents an active upload session."""

    id: str
    filename: str
    file_size: int
    chunk_size: int
    total_chunks: int
    received_chunks: set = field(default_factory=set)
    temp_dir: Path = field(default_factory=Path)
    workspace: str = ""
    connection_id: str = ""
    store_name: str = ""
    created_at: datetime = field(default_factory=datetime.now)
    completed: bool = False
    error: str = ""

    @property
    def progress(self) -> float:
        """Calculate upload progress percentage."""
        if self.total_chunks == 0:
            return 0.0
        return len(self.received_chunks) / self.total_chunks * 100

    @property
    def bytes_received(self) -> int:
        """Calculate bytes received so far."""
        return len(self.received_chunks) * self.chunk_size

    def is_complete(self) -> bool:
        """Check if all chunks have been received."""
        return len(self.received_chunks) == self.total_chunks


class UploadSessionManager:
    """Thread-safe manager for upload sessions."""

    _instance: "UploadSessionManager | None" = None
    _lock = threading.RLock()

    def __new__(cls) -> "UploadSessionManager":
        """Singleton pattern for session manager."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._sessions: dict[str, UploadSession] = {}
        return cls._instance

    def create_session(
        self,
        filename: str,
        file_size: int,
        chunk_size: int,
        workspace: str = "",
        connection_id: str = "",
        store_name: str = "",
    ) -> UploadSession:
        """Create a new upload session."""
        with self._lock:
            session_id = str(uuid.uuid4())

            # Calculate total chunks
            total_chunks = (file_size + chunk_size - 1) // chunk_size

            # Create temp directory for chunks
            cache_dir = get_cache_dir()
            temp_dir = cache_dir / "uploads" / session_id
            temp_dir.mkdir(parents=True, exist_ok=True)

            session = UploadSession(
                id=session_id,
                filename=filename,
                file_size=file_size,
                chunk_size=chunk_size,
                total_chunks=total_chunks,
                temp_dir=temp_dir,
                workspace=workspace,
                connection_id=connection_id,
                store_name=store_name,
            )

            self._sessions[session_id] = session
            return session

    def get_session(self, session_id: str) -> UploadSession | None:
        """Get an upload session by ID."""
        with self._lock:
            return self._sessions.get(session_id)

    def delete_session(self, session_id: str) -> None:
        """Delete an upload session and clean up temp files."""
        with self._lock:
            session = self._sessions.pop(session_id, None)
            if session and session.temp_dir.exists():
                shutil.rmtree(session.temp_dir, ignore_errors=True)

    def add_chunk(self, session_id: str, chunk_index: int, data: bytes) -> None:
        """Add a chunk to an upload session."""
        with self._lock:
            session = self._sessions.get(session_id)
            if not session:
                raise UploadError("Session not found", session_id)

            # Write chunk to temp file
            chunk_path = session.temp_dir / f"chunk_{chunk_index:06d}"
            with open(chunk_path, "wb") as f:
                f.write(data)

            session.received_chunks.add(chunk_index)

    def assemble_file(self, session_id: str) -> Path:
        """Assemble chunks into final file."""
        with self._lock:
            session = self._sessions.get(session_id)
            if not session:
                raise UploadError("Session not found", session_id)

            if not session.is_complete():
                missing = set(range(session.total_chunks)) - session.received_chunks
                raise UploadError(f"Missing chunks: {missing}", session_id)

            # Assemble file
            final_path = session.temp_dir / session.filename
            with open(final_path, "wb") as outfile:
                for i in range(session.total_chunks):
                    chunk_path = session.temp_dir / f"chunk_{i:06d}"
                    with open(chunk_path, "rb") as chunk:
                        outfile.write(chunk.read())

            session.completed = True
            return final_path


# Global session manager
session_manager = UploadSessionManager()


class UploadInitView(APIView):
    """Initialize a chunked upload session."""

    def post(self, request):
        """Create a new upload session.

        Expected body:
        {
            "filename": "data.shp.zip",
            "fileSize": 1048576,
            "chunkSize": 524288,
            "workspace": "topp",
            "connectionId": "conn_123",
            "storeName": "my_store"
        }
        """
        filename = request.data.get("filename")
        file_size = request.data.get("fileSize")
        chunk_size = request.data.get("chunkSize", 5 * 1024 * 1024)  # Default 5MB
        workspace = request.data.get("workspace", "")
        connection_id = request.data.get("connectionId", "")
        store_name = request.data.get("storeName", "")

        if not filename or not file_size:
            return Response(
                {"error": "filename and fileSize are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        # Check file size limit
        max_size = getattr(settings, "UPLOAD_MAX_FILE_SIZE", 10 * 1024 * 1024 * 1024)
        if file_size > max_size:
            return Response(
                {"error": f"File too large. Maximum size is {max_size} bytes"},
                status=status.HTTP_413_REQUEST_ENTITY_TOO_LARGE,
            )

        session = session_manager.create_session(
            filename=filename,
            file_size=file_size,
            chunk_size=chunk_size,
            workspace=workspace,
            connection_id=connection_id,
            store_name=store_name,
        )

        return Response(
            {
                "sessionId": session.id,
                "filename": session.filename,
                "fileSize": session.file_size,
                "chunkSize": session.chunk_size,
                "totalChunks": session.total_chunks,
            },
            status=status.HTTP_201_CREATED,
        )


class UploadChunkView(APIView):
    """Upload a single chunk."""

    parser_classes = [MultiPartParser, FormParser]

    def post(self, request):
        """Upload a chunk.

        Expected form data:
        - sessionId: Upload session ID
        - chunkIndex: Zero-based chunk index
        - chunk: The chunk data (file)
        """
        session_id = request.data.get("sessionId")
        chunk_index = request.data.get("chunkIndex")
        chunk_file = request.FILES.get("chunk")

        if not session_id or chunk_index is None or not chunk_file:
            return Response(
                {"error": "sessionId, chunkIndex, and chunk are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        try:
            chunk_index = int(chunk_index)
        except ValueError:
            return Response(
                {"error": "chunkIndex must be an integer"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        session = session_manager.get_session(session_id)
        if not session:
            return Response(
                {"error": "Upload session not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        if chunk_index < 0 or chunk_index >= session.total_chunks:
            return Response(
                {"error": f"Invalid chunk index. Must be 0-{session.total_chunks - 1}"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        try:
            chunk_data = chunk_file.read()
            session_manager.add_chunk(session_id, chunk_index, chunk_data)

            return Response(
                {
                    "sessionId": session_id,
                    "chunkIndex": chunk_index,
                    "receivedChunks": len(session.received_chunks),
                    "totalChunks": session.total_chunks,
                    "progress": session.progress,
                }
            )
        except UploadError as e:
            return Response({"error": str(e)}, status=status.HTTP_400_BAD_REQUEST)


class UploadCompleteView(APIView):
    """Complete a chunked upload and publish to GeoServer."""

    def post(self, request):
        """Complete the upload.

        Expected body:
        {
            "sessionId": "uuid",
            "publish": true,
            "storeName": "my_store"
        }
        """
        session_id = request.data.get("sessionId")
        publish = request.data.get("publish", True)
        store_name = request.data.get("storeName")

        if not session_id:
            return Response(
                {"error": "sessionId is required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        session = session_manager.get_session(session_id)
        if not session:
            return Response(
                {"error": "Upload session not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        if not session.is_complete():
            missing = session.total_chunks - len(session.received_chunks)
            return Response(
                {"error": f"Upload incomplete. Missing {missing} chunks."},
                status=status.HTTP_400_BAD_REQUEST,
            )

        try:
            # Assemble the file
            file_path = session_manager.assemble_file(session_id)

            result = {
                "sessionId": session_id,
                "filename": session.filename,
                "fileSize": session.file_size,
                "path": str(file_path),
            }

            # Publish to GeoServer if requested
            if publish and session.connection_id and session.workspace:
                final_store_name = store_name or session.store_name or Path(session.filename).stem

                client = get_geoserver_client(session.connection_id)

                # Read the file
                with open(file_path, "rb") as f:
                    data = f.read()

                # Determine file type and upload
                filename_lower = session.filename.lower()
                if filename_lower.endswith(".zip") or filename_lower.endswith(".shp"):
                    client.upload_shapefile(session.workspace, final_store_name, data)
                    result["storeType"] = "shapefile"
                elif filename_lower.endswith(".tif") or filename_lower.endswith(".tiff"):
                    client.upload_geotiff(session.workspace, final_store_name, data)
                    result["storeType"] = "geotiff"
                elif filename_lower.endswith(".gpkg"):
                    client.upload_geopackage(session.workspace, final_store_name, data)
                    result["storeType"] = "geopackage"
                else:
                    result["warning"] = f"Unknown file type: {session.filename}"

                result["storeName"] = final_store_name
                result["workspace"] = session.workspace
                result["published"] = True

            return Response(result)

        except Exception as e:
            return Response(
                {"error": f"Failed to complete upload: {str(e)}"},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR,
            )
        finally:
            # Clean up session
            session_manager.delete_session(session_id)


class UploadProgressView(APIView):
    """Get upload progress for a session."""

    def get(self, request, session_id):
        """Get upload progress."""
        session = session_manager.get_session(session_id)
        if not session:
            return Response(
                {"error": "Upload session not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        return Response(
            {
                "sessionId": session.id,
                "filename": session.filename,
                "fileSize": session.file_size,
                "receivedChunks": len(session.received_chunks),
                "totalChunks": session.total_chunks,
                "bytesReceived": session.bytes_received,
                "progress": session.progress,
                "completed": session.completed,
                "error": session.error,
            }
        )


class UploadCancelView(APIView):
    """Cancel an upload session."""

    def delete(self, request, session_id):
        """Cancel and clean up an upload session."""
        session = session_manager.get_session(session_id)
        if not session:
            return Response(
                {"error": "Upload session not found"},
                status=status.HTTP_404_NOT_FOUND,
            )

        session_manager.delete_session(session_id)
        return Response(status=status.HTTP_204_NO_CONTENT)


class SimpleUploadView(APIView):
    """Simple (non-chunked) file upload."""

    parser_classes = [MultiPartParser, FormParser]

    def post(self, request):
        """Upload a file directly.

        Expected form data:
        - file: The file to upload
        - workspace: Target workspace
        - connectionId: GeoServer connection ID
        - storeName: Optional store name (defaults to filename)
        """
        uploaded_file = request.FILES.get("file")
        workspace = request.data.get("workspace")
        connection_id = request.data.get("connectionId")
        store_name = request.data.get("storeName")

        if not uploaded_file:
            return Response(
                {"error": "file is required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        if not workspace or not connection_id:
            return Response(
                {"error": "workspace and connectionId are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        filename = uploaded_file.name
        final_store_name = store_name or Path(filename).stem

        try:
            client = get_geoserver_client(connection_id)
            data = uploaded_file.read()

            result = {
                "filename": filename,
                "fileSize": len(data),
                "workspace": workspace,
                "storeName": final_store_name,
            }

            # Determine file type and upload
            filename_lower = filename.lower()
            if filename_lower.endswith(".zip") or filename_lower.endswith(".shp"):
                client.upload_shapefile(workspace, final_store_name, data)
                result["storeType"] = "shapefile"
            elif filename_lower.endswith(".tif") or filename_lower.endswith(".tiff"):
                client.upload_geotiff(workspace, final_store_name, data)
                result["storeType"] = "geotiff"
            elif filename_lower.endswith(".gpkg"):
                client.upload_geopackage(workspace, final_store_name, data)
                result["storeType"] = "geopackage"
            else:
                return Response(
                    {"error": f"Unsupported file type: {filename}"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            result["published"] = True
            return Response(result, status=status.HTTP_201_CREATED)

        except Exception as e:
            return Response(
                {"error": f"Upload failed: {str(e)}"},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR,
            )
