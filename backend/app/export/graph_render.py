"""Render a dungeon graph to a PNG (used when embedding in PDF exports)."""
from __future__ import annotations

import base64
from io import BytesIO

import matplotlib

matplotlib.use("Agg")  # headless backend — must be set before pyplot import

import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
import networkx as nx

from app.models import Dungeon

# Palette mirrors frontend's role colors (App.tsx DungeonMap)
ROLE_COLORS: dict[str, str] = {
    "entrance": "#4a9f5c",
    "boss_room": "#c93d28",
    "vault": "#c9a227",
    "trap_room": "#d4763b",
    "secret_room": "#c9a227",
    "shrine": "#5a8f9a",
    "lair": "#9e2f20",
    "rest_area": "#4a9f5c",
    "guard_post": "#6b7280",
    "puzzle_room": "#8b7355",
    "armory": "#6b7280",
    "corridor": "#44403a",
}

DEFAULT_COLOR = "#44403a"
BG = "#1a1814"
NODE_TEXT = "#efe5d1"
EDGE_COLOR = "#6b6357"
LOCKED_COLOR = "#c9a227"
HIDDEN_COLOR = "#8a7a3d"


def render_graph_png(dungeon: Dungeon, *, dpi: int = 150) -> bytes:
    """Render the dungeon graph to PNG bytes."""
    if not dungeon.rooms:
        # Still return a valid (blank) PNG so the template never breaks
        fig, ax = plt.subplots(figsize=(6, 4), dpi=dpi, facecolor=BG)
        ax.set_axis_off()
        ax.text(0.5, 0.5, "(empty dungeon)", color=NODE_TEXT, ha="center", va="center")
        buf = BytesIO()
        fig.savefig(buf, format="png", facecolor=BG)
        plt.close(fig)
        return buf.getvalue()

    G = nx.Graph()
    for room in dungeon.rooms:
        G.add_node(room.room_id, role=room.role.value)
    for edge in dungeon.edges:
        G.add_edge(edge.from_room, edge.to_room, locked=edge.is_locked, hidden=edge.is_hidden)

    # Layered layout rooted at entrance (room 0 if present, else first)
    root = 0 if any(r.room_id == 0 for r in dungeon.rooms) else dungeon.rooms[0].room_id
    try:
        pos = _layered_layout(G, root)
    except Exception:
        pos = nx.spring_layout(G, seed=dungeon.seed)

    n_rooms = len(dungeon.rooms)
    fig_w = max(8.0, min(14.0, 1.1 * (max(len(layer) for layer in _bfs_layers(G, root)) or 1)))
    fig_h = max(5.0, min(11.0, 1.0 * len(_bfs_layers(G, root))))
    fig, ax = plt.subplots(figsize=(fig_w, fig_h), dpi=dpi, facecolor=BG)
    ax.set_facecolor(BG)
    ax.set_axis_off()

    # Draw edges, grouped by style so dashed ones look right
    solid = [(u, v) for u, v, d in G.edges(data=True) if not d.get("locked") and not d.get("hidden")]
    locked = [(u, v) for u, v, d in G.edges(data=True) if d.get("locked") and not d.get("hidden")]
    hidden = [(u, v) for u, v, d in G.edges(data=True) if d.get("hidden")]
    if solid:
        nx.draw_networkx_edges(G, pos, edgelist=solid, ax=ax, edge_color=EDGE_COLOR, width=1.8, alpha=0.9)
    if locked:
        nx.draw_networkx_edges(G, pos, edgelist=locked, ax=ax, edge_color=LOCKED_COLOR, width=2.0, style="dashed", alpha=0.9)
    if hidden:
        nx.draw_networkx_edges(G, pos, edgelist=hidden, ax=ax, edge_color=HIDDEN_COLOR, width=1.4, style="dotted", alpha=0.7)

    # Draw nodes
    room_by_id = {r.room_id: r for r in dungeon.rooms}
    node_colors = [ROLE_COLORS.get(room_by_id[n].role.value, DEFAULT_COLOR) for n in G.nodes()]
    node_edges = [
        "#f5c542" if room_by_id[n].is_boss_room else "#2a2620"
        for n in G.nodes()
    ]
    node_widths = [2.5 if room_by_id[n].is_boss_room else 1.2 for n in G.nodes()]
    nx.draw_networkx_nodes(
        G, pos, ax=ax, node_color=node_colors, node_size=900,
        edgecolors=node_edges, linewidths=node_widths, alpha=0.95,
    )
    nx.draw_networkx_labels(
        G, pos, ax=ax,
        labels={n: str(n) for n in G.nodes()},
        font_color=NODE_TEXT, font_size=10, font_weight="bold",
    )

    # Legend
    legend_roles = ["entrance", "boss_room", "vault", "trap_room", "shrine", "lair", "rest_area", "corridor"]
    handles = [
        mpatches.Patch(color=ROLE_COLORS[role], label=role.replace("_", " ").title())
        for role in legend_roles
    ]
    legend = ax.legend(
        handles=handles, loc="lower center", bbox_to_anchor=(0.5, -0.08),
        ncol=4, frameon=False, fontsize=8,
    )
    for text in legend.get_texts():
        text.set_color(NODE_TEXT)

    ax.set_title(
        f"{dungeon.name}  ·  {n_rooms} rooms",
        color=NODE_TEXT, fontsize=12, pad=12,
    )

    buf = BytesIO()
    fig.savefig(buf, format="png", facecolor=BG, bbox_inches="tight", pad_inches=0.3)
    plt.close(fig)
    return buf.getvalue()


def render_graph_data_uri(dungeon: Dungeon) -> str:
    """Render the graph and return a data: URI — handy for HTML templates."""
    png = render_graph_png(dungeon)
    b64 = base64.b64encode(png).decode("ascii")
    return f"data:image/png;base64,{b64}"


def _bfs_layers(G: nx.Graph, root: int) -> list[list[int]]:
    visited: set[int] = {root}
    layers: list[list[int]] = [[root]]
    while True:
        last = layers[-1]
        nxt: list[int] = []
        for node in last:
            for nb in G.neighbors(node):
                if nb not in visited:
                    visited.add(nb)
                    nxt.append(nb)
        if not nxt:
            break
        layers.append(nxt)
    # Any disconnected nodes go at the bottom so they still render
    leftover = [n for n in G.nodes() if n not in visited]
    if leftover:
        layers.append(leftover)
    return layers


def _layered_layout(G: nx.Graph, root: int) -> dict[int, tuple[float, float]]:
    layers = _bfs_layers(G, root)
    pos: dict[int, tuple[float, float]] = {}
    h = len(layers)
    for li, layer in enumerate(layers):
        w = len(layer)
        y = (h - 1 - li) / max(h - 1, 1)  # top = entrance
        for ni, node in enumerate(layer):
            x = (ni + 1) / (w + 1)
            pos[node] = (x, y)
    return pos
