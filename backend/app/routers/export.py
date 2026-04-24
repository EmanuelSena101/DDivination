"""API routes for exporting generated dungeons as PDF or Markdown."""
from __future__ import annotations

import re
from urllib.parse import quote

from fastapi import APIRouter, HTTPException
from fastapi.responses import Response

from app.export.markdown import render_markdown
from app.models import Dungeon

router = APIRouter(prefix="/api/export", tags=["export"])


def _safe_slug(name: str, *, fallback: str = "dungeon") -> str:
    slug = re.sub(r"[^a-zA-Z0-9]+", "-", name.strip()).strip("-").lower()
    return slug or fallback


def _content_disposition(filename: str) -> str:
    # RFC 5987 to handle non-ASCII names safely
    ascii_fallback = filename.encode("ascii", "ignore").decode("ascii") or "dungeon"
    return f"attachment; filename=\"{ascii_fallback}\"; filename*=UTF-8''{quote(filename)}"


@router.post("/markdown")
async def export_markdown(dungeon: Dungeon) -> Response:
    """Render a dungeon to a GM-facing Markdown document."""
    content = render_markdown(dungeon)
    filename = f"{_safe_slug(dungeon.name)}.md"
    return Response(
        content=content,
        media_type="text/markdown; charset=utf-8",
        headers={"Content-Disposition": _content_disposition(filename)},
    )


@router.post("/pdf")
async def export_pdf(dungeon: Dungeon) -> Response:
    """Render a dungeon to a GM-facing PDF document."""
    try:
        from app.export.pdf import render_pdf
    except ImportError as exc:
        raise HTTPException(
            status_code=503,
            detail=(
                "PDF export unavailable: WeasyPrint is not installed or its native "
                "dependencies (pango, cairo) are missing. "
                f"Original error: {exc}"
            ),
        ) from exc

    try:
        pdf_bytes = render_pdf(dungeon)
    except OSError as exc:
        # WeasyPrint raises OSError when libpango/cairo can't be loaded
        raise HTTPException(
            status_code=503,
            detail=(
                "PDF rendering failed — system libraries for WeasyPrint are missing. "
                f"Original error: {exc}"
            ),
        ) from exc

    filename = f"{_safe_slug(dungeon.name)}.pdf"
    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={"Content-Disposition": _content_disposition(filename)},
    )
