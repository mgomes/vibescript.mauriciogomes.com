#!/usr/bin/env python3

from __future__ import annotations

import json
import sys
import urllib.parse
import urllib.request

USER_AGENT = "Mozilla/5.0 Codex"
API_BASE = "https://rosettacode.org/w/api.php"
BLACKLIST = [
    "active directory",
    "animation",
    "window",
    "gui",
    "web",
    "http",
    "https",
    "ftp",
    "file",
    "directory",
    "database",
    "socket",
    "server",
    "client",
    "sound",
    "image",
    "bitmap",
    "png",
    "jpg",
    "svg",
    "html",
    "xml",
    "json",
    "csv",
    "sqlite",
    "mysql",
    "postgres",
    "open",
    "read",
    "write",
    "append",
    "rename",
    "create",
    "delete",
    "terminal",
    "concurrent",
    "parallel",
    "distributed",
    "browser",
    "opengl",
    "vulkan",
    "sdl",
    "x11",
    "gtk",
    "wx",
    "tk",
]


def fetch_json(params: dict[str, str]) -> dict:
    url = API_BASE + "?" + urllib.parse.urlencode(params)
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def fetch_ruby_category_titles() -> list[str]:
    titles: list[str] = []
    params = {
        "action": "query",
        "list": "categorymembers",
        "cmtitle": "Category:Ruby",
        "cmlimit": "500",
        "format": "json",
    }
    continuation: dict[str, str] = {}

    while True:
        payload = fetch_json(params | continuation)
        titles.extend(item["title"] for item in payload["query"]["categorymembers"])
        if "continue" not in payload:
            return titles
        continuation = payload["continue"]


def candidate_titles(titles: list[str]) -> list[str]:
    candidates = []
    for title in titles:
        lowered = title.lower()
        if "/" in title:
            continue
        if any(term in lowered for term in BLACKLIST):
            continue
        candidates.append(title)
    return candidates


def fetch_page_lengths(titles: list[str]) -> list[dict[str, object]]:
    ranked: list[dict[str, object]] = []
    for index in range(0, len(titles), 50):
        batch = titles[index : index + 50]
        payload = fetch_json(
            {
                "action": "query",
                "prop": "info",
                "titles": "|".join(batch),
                "format": "json",
            }
        )
        for page in payload["query"]["pages"].values():
            ranked.append(
                {
                    "title": page["title"],
                    "length": page.get("length", 0),
                    "url": "https://rosettacode.org/wiki/" + page["title"].replace(" ", "_"),
                }
            )
    ranked.sort(key=lambda item: (-int(item["length"]), str(item["title"])))
    return ranked


def main() -> int:
    output_path = sys.argv[1] if len(sys.argv) > 1 else "docs/rosettacode_ruby_top_250.json"
    titles = fetch_ruby_category_titles()
    candidates = candidate_titles(titles)
    ranked = fetch_page_lengths(candidates)[:250]

    payload = {
        "source": "https://rosettacode.org/wiki/Category:Ruby",
        "selection_heuristic": "Top Ruby-category tasks after side-effect-heavy keyword filtering, ranked by Rosetta Code page length as a popularity proxy.",
        "total_ruby_tasks": len(titles),
        "candidate_tasks": len(candidates),
        "tasks": ranked,
    }

    with open(output_path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2)
        handle.write("\n")

    print(f"wrote {len(ranked)} tasks to {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
