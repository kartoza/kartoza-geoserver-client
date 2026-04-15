"""Shared utility functions for Kartoza CloudBench."""

import fcntl
import os
from contextlib import contextmanager
from pathlib import Path
from typing import Generator

# Config directory name — kept here to avoid circular imports with config.py
CONFIG_DIR = "kartoza-cloudbench"


def get_xdg_config_path(filename: str, user_id: str = "default") -> str:
    """Build an XDG-compliant config file path for the given user.

    Path: ${XDG_CONFIG_HOME}/<user_id>/kartoza-cloudbench/<filename>

    Raises ValueError when CLOUDBENCH_MUST_AUTHENTICATED is True and
    user_id is still "default" (i.e., no authenticated user on the thread).
    """
    from django.conf import settings

    if settings.CLOUDBENCH_MUST_AUTHENTICATED and user_id == "default":
        raise ValueError("User ID is required for authenticated access.")

    config_home = os.environ.get("XDG_CONFIG_HOME") or os.path.join(
        str(Path.home()), ".config"
    )
    return os.path.join(config_home, user_id, CONFIG_DIR, filename)

def get_xdg_data_path(filename: str, user_id: str = "default") -> str:
    """Build an XDG-compliant config file path for the given user.

    Path: ${XDG_DATA_HOME}/<user_id>/kartoza-cloudbench/<filename>

    Raises ValueError when CLOUDBENCH_MUST_AUTHENTICATED is True and
    user_id is still "default" (i.e., no authenticated user on the thread).
    """
    from django.conf import settings

    if settings.CLOUDBENCH_MUST_AUTHENTICATED and user_id == "default":
        raise ValueError("User ID is required for authenticated access.")

    config_home = os.environ.get("XDG_DATA_HOME") or os.path.join(
        str(Path.home()), ".config"
    )
    return os.path.join(config_home, user_id, CONFIG_DIR, filename)

def get_xdg_cache_path(filename: str, user_id: str = "default") -> str:
    """Build an XDG-compliant config file path for the given user.

    Path: ${XDG_CACHE_HOME}/<user_id>/kartoza-cloudbench/<filename>

    Raises ValueError when CLOUDBENCH_MUST_AUTHENTICATED is True and
    user_id is still "default" (i.e., no authenticated user on the thread).
    """
    config_home = os.environ.get("XDG_CACHE_HOME") or os.path.join(
        str(Path.home()), ".config"
    )
    return os.path.join(config_home, user_id, CONFIG_DIR, filename)


@contextmanager
def file_lock(path: str, exclusive: bool = True) -> Generator[None, None, None]:
    """Context manager for cross-process file locking.

    Uses a sibling .lock file so the data file itself is never held open
    while locked. Acquires an exclusive lock for writes, shared for reads.

    Usage:
        with file_lock(path):           # exclusive (write)
            ...
        with file_lock(path, exclusive=False):  # shared (read)
            ...
    """
    lock_path = path + ".lock"
    mode = fcntl.LOCK_EX if exclusive else fcntl.LOCK_SH
    with open(lock_path, "w") as lock_file:
        fcntl.flock(lock_file, mode)
        try:
            yield
        finally:
            fcntl.flock(lock_file, fcntl.LOCK_UN)