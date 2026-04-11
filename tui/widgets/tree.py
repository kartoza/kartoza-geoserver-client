"""Resource tree widget for Kartoza CloudBench TUI."""

from textual.widgets import Tree
from textual.widgets.tree import TreeNode


class ResourceTreeWidget(Tree):
    """A tree widget for browsing geospatial resources.

    Supports lazy loading of children and different node types.
    """

    DEFAULT_CSS = """
    ResourceTreeWidget {
        background: $surface;
        padding: 1;
    }

    ResourceTreeWidget > .tree--cursor {
        background: $primary-darken-1;
    }
    """

    # Node type icons using Nerd Font codepoints
    ICONS = {
        "workspace": "\uf07b",  # folder
        "datastore": "\uf1c0",  # database
        "coveragestore": "\uf03e",  # image
        "featuretype": "\uf0ac",  # globe
        "coverage": "\uf279",  # map
        "layer": "\uf5fd",  # layer-group
        "style": "\uf1fc",  # paint-brush
        "layergroup": "\uf5fd",  # layer-group
        "bucket": "\uf0c2",  # cloud
        "object": "\uf15b",  # file
        "schema": "\uf0e8",  # sitemap
        "table": "\uf0ce",  # table
        "column": "\uf0db",  # columns
        "default": "\uf15b",  # file
    }

    def __init__(self, label: str = "Resources", **kwargs):
        """Initialize the resource tree.

        Args:
            label: Root node label
            **kwargs: Additional arguments passed to Tree
        """
        super().__init__(label, **kwargs)

    def get_icon(self, node_type: str) -> str:
        """Get the icon for a node type.

        Args:
            node_type: Type of the node

        Returns:
            Icon string (Nerd Font codepoint)
        """
        return self.ICONS.get(node_type, self.ICONS["default"])

    def add_resource_node(
        self,
        parent: TreeNode,
        label: str,
        node_type: str,
        data: dict | None = None,
        expandable: bool = False,
    ) -> TreeNode:
        """Add a resource node to the tree.

        Args:
            parent: Parent node
            label: Node label
            node_type: Type of the node (for icon selection)
            data: Optional data to attach to the node
            expandable: Whether the node can be expanded

        Returns:
            The created TreeNode
        """
        icon = self.get_icon(node_type)
        full_label = f"{icon} {label}"

        node_data = data or {}
        node_data["type"] = node_type

        if expandable:
            return parent.add(full_label, data=node_data)
        else:
            return parent.add_leaf(full_label, data=node_data)
