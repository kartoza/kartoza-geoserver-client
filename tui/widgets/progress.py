"""Progress indicator widgets for Kartoza CloudBench TUI."""

from textual.app import ComposeResult
from textual.containers import Horizontal
from textual.widgets import Label, ProgressBar, Static


class ProgressIndicator(Static):
    """A progress indicator widget with label and percentage."""

    DEFAULT_CSS = """
    ProgressIndicator {
        height: 3;
        padding: 0 1;
    }

    ProgressIndicator .progress-label {
        width: 20;
    }

    ProgressIndicator ProgressBar {
        width: 1fr;
    }

    ProgressIndicator .progress-percentage {
        width: 8;
        text-align: right;
    }
    """

    def __init__(
        self,
        label: str = "Progress",
        total: float = 100.0,
        **kwargs,
    ):
        """Initialize progress indicator.

        Args:
            label: Label text
            total: Total value (100 for percentage)
            **kwargs: Additional arguments passed to Static
        """
        super().__init__(**kwargs)
        self._label = label
        self._total = total
        self._progress = 0.0

    def compose(self) -> ComposeResult:
        """Create the progress indicator layout."""
        with Horizontal():
            yield Label(self._label, classes="progress-label")
            yield ProgressBar(total=self._total, id="progress-bar")
            yield Label("0%", classes="progress-percentage", id="progress-pct")

    def update_progress(self, value: float) -> None:
        """Update the progress value.

        Args:
            value: Current progress value
        """
        self._progress = min(value, self._total)

        bar = self.query_one("#progress-bar", ProgressBar)
        bar.update(progress=self._progress)

        pct = self.query_one("#progress-pct", Label)
        percentage = (self._progress / self._total) * 100 if self._total > 0 else 0
        pct.update(f"{percentage:.0f}%")

    @property
    def progress(self) -> float:
        """Get current progress value."""
        return self._progress

    @property
    def is_complete(self) -> bool:
        """Check if progress is complete."""
        return self._progress >= self._total
