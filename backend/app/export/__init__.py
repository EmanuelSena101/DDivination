"""Export module: render generated dungeons as Markdown or PDF.

PDF rendering requires matplotlib + weasyprint with native libs, so those are
imported lazily by the PDF path only. Markdown has no heavy deps.
"""
from app.export.markdown import render_markdown

__all__ = ["render_markdown"]
