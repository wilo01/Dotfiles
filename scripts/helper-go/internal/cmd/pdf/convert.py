#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = ["pymupdf4llm>=0.0.17"]
# ///
"""PDF -> LLM-friendly markdown converter, driven by `hlp pdf`.

Usage: convert.py <input.pdf> <output-dir> <basename> [--raw]

Exit codes:
  0 success (one-line JSON summary on stdout)
  3 scanned / image-only PDF (no usable text layer)
"""
import datetime
import json
import os
import pathlib
import re
import sys

# Below this many extracted characters per page the PDF is treated as scanned.
SCANNED_CHARS_PER_PAGE = 20

IMAGE_REF = re.compile(r"!\[[^\]]*\]\([^)]*\)")
UNDERLINE_TAG = re.compile(r"</?u>")
DOT_LEADER = re.compile(r"\s*\.{5,}\s*")


def clean(markdown: str) -> str:
    markdown = UNDERLINE_TAG.sub("", markdown)
    return DOT_LEADER.sub(" — ", markdown)


def front_matter(title: str, source: str, pages: int) -> str:
    lines = [
        "---",
        f"title: {json.dumps(title)}",
        f"source: {json.dumps(source)}",
        f"pages: {pages}",
        f"converted: {datetime.date.today().isoformat()}",
        "converter: pymupdf4llm",
        "---",
        "",
        "",
    ]
    return "\n".join(lines)


def main() -> None:
    args = [a for a in sys.argv[1:] if a != "--raw"]
    raw = "--raw" in sys.argv[1:]
    if len(args) != 3:
        print(__doc__, file=sys.stderr)
        sys.exit(2)
    pdf_path = os.path.abspath(args[0])
    out_dir = pathlib.Path(args[1]).resolve()
    basename = args[2]

    images_dir = out_dir / "images"
    images_dir.mkdir(parents=True, exist_ok=True)

    import pymupdf4llm

    # Relative image_path + cwd inside out_dir => portable "images/..." links.
    os.chdir(out_dir)
    chunks = pymupdf4llm.to_markdown(
        pdf_path,
        page_chunks=True,
        write_images=True,
        image_path="images",
        image_format="png",
    )

    pages = len(chunks)
    total_chars = sum(len(IMAGE_REF.sub("", c["text"]).strip()) for c in chunks)
    if total_chars < SCANNED_CHARS_PER_PAGE * pages:
        print(
            f"extracted only {total_chars} chars across {pages} pages",
            file=sys.stderr,
        )
        sys.exit(3)

    if raw:
        markdown = "".join(c["text"] for c in chunks)
    else:
        parts = []
        for i, chunk in enumerate(chunks, 1):
            page_no = chunk.get("metadata", {}).get("page", i)
            parts.append(f"<!-- page {page_no} -->\n\n{chunk['text'].strip()}\n")
        markdown = clean("\n".join(parts))
        meta = chunks[0].get("metadata", {}) if chunks else {}
        title = (meta.get("title") or "").strip() or basename
        markdown = front_matter(title, pdf_path, pages) + markdown

    md_file = out_dir / f"{basename}.md"
    md_file.write_text(markdown, encoding="utf-8")

    image_count = sum(1 for p in images_dir.iterdir() if p.is_file())
    if image_count == 0:
        images_dir.rmdir()
    table_rows = sum(
        1 for line in markdown.splitlines() if line.lstrip().startswith("|")
    )
    print(
        json.dumps(
            {
                "markdown": str(md_file),
                "pages": pages,
                "chars": total_chars,
                "images": image_count,
                "table_rows": table_rows,
            }
        )
    )


if __name__ == "__main__":
    main()
