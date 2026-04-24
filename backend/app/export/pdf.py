"""Render a Dungeon to a GM-facing PDF using Jinja2 + WeasyPrint."""
from __future__ import annotations

from pathlib import Path

from jinja2 import Environment, FileSystemLoader, select_autoescape

from app.export.graph_render import render_graph_data_uri
from app.models import Dungeon

_TEMPLATE_DIR = Path(__file__).parent / "templates"


def _title_clean(value: str) -> str:
    return value.replace("_", " ").title()


_env = Environment(
    loader=FileSystemLoader(str(_TEMPLATE_DIR)),
    autoescape=select_autoescape(enabled_extensions=("html", "j2"), default_for_string=True),
)
_env.filters["title_clean"] = _title_clean


def render_pdf(dungeon: Dungeon) -> bytes:
    """Render a dungeon as a PDF document. Returns PDF bytes."""
    # Imported lazily so the module can be imported (tests, markdown-only) without
    # having the native weasyprint dependencies installed.
    from weasyprint import HTML

    template = _env.get_template("dungeon_gm.html.j2")
    stylesheet = (_TEMPLATE_DIR / "dungeon_gm.css").read_text(encoding="utf-8")
    graph_data_uri = render_graph_data_uri(dungeon)

    html = template.render(
        dungeon=dungeon,
        cfg=dungeon.config,
        graph_data_uri=graph_data_uri,
        stylesheet=stylesheet,
    )
    return HTML(string=html).write_pdf()
