#!/usr/bin/env python3
"""
Scrape Swedish recipes from ICA.se and generate a SQL seed migration.

Usage:
    source tools/.venv/bin/activate
    python3 tools/scrape_ica_recipes.py

Outputs: backend/migrations/012_seed_100_recipes.sql
"""

import json
import re
import sys
import time
import uuid
from dataclasses import dataclass, field

import requests
from bs4 import BeautifulSoup
from recipe_scrapers import scrape_html

# ── Existing recipe names to skip (from migrations 004 + 008) ──

EXISTING_NAMES = {
    "pasta carbonara",
    "kycklingwok",
    "tacos",
    "laxfilé med potatis",
    "köttfärssås",
    "pannkakor",
    "kycklinggryta med ris",
    "ärtsoppa med pannkakor",
    "falukorv med stuvade makaroner",
    "fiskpinnar med potatismos",
    "korvstroganoff med ris",
    "janssons frestelse",
    "pytt i panna",
    "vegetarisk pasta med pesto",
    "stekt fläsk med löksås",
    "köttbullar med gräddsås och potatis",
    "ugnsbakad torsk med rotfrukter",
    "chili con carne",
    "tomatsoppa med ostmacka",
    "kyckling med currysås och ris",
}

# ── Category pages to scrape ──

CATEGORIES = {
    "/recept/middag/": ["vardag"],
    "/recept/kyckling/middag/": ["kyckling"],
    "/recept/fisk/middag/": ["fisk"],
    "/recept/svensk/middag/": ["husmanskost"],
    "/recept/snabb/middag/": ["snabb", "vardag"],
    "/recept/soppa/": ["soppa"],
    "/recept/pasta/middag/": ["pasta"],
    "/recept/enkel/middag/": ["vardag", "enkel"],
    "/recept/vegetariskt/middag/": ["vegetariskt"],
    "/recept/kottfars/": ["köttfärs"],
    "/recept/gratang/": ["gratäng"],
    "/recept/gryta/": ["gryta"],
    "/recept/sallad/": ["sallad"],
    "/recept/vardag/middag/": ["vardag"],
    "/recept/lax/": ["fisk", "lax"],
    "/recept/kottbullar/": ["husmanskost"],
}

# ── Emoji mapping ──

EMOJI_KEYWORDS = [
    (["kyckling", "kycklingfilé", "kycklingfärs", "kycklingklubba", "kycklinglår"], "🍗"),
    (["lax", "torsk", "fisk", "räkor", "sej", "kolja", "pangasius", "tonfisk", "sill", "makrill"], "🐟"),
    (["pasta", "spaghetti", "penne", "fusilli", "tagliatelle", "linguine", "makaroner", "lasagne"], "🍝"),
    (["soppa", "buljong"], "🍲"),
    (["gryta", "stuvning"], "🥘"),
    (["taco", "burrito", "enchilada", "mexikan"], "🌮"),
    (["wok", "asiatisk", "teriyaki", "soja", "sweet chili"], "🥡"),
    (["sallad"], "🥗"),
    (["paj", "pajdeg", "quiche"], "🥧"),
    (["pannkak", "plättar", "crêpe"], "🥞"),
    (["vegetar", "vegan", "halloumi", "linser", "bönor", "tofu", "kikärt"], "🥬"),
    (["nöt", "biff", "entrecôte", "högrev", "oxfilé", "fläsk", "korv", "bacon", "köttfärs", "revben"], "🥩"),
    (["gratäng", "ugns"], "🧀"),
]

# ── Swedish unit patterns ──

UNIT_PATTERN = (
    r"(?:g|kg|dl|l|ml|cl|msk|tsk|krm|st|klyftor|klyfta|skivor|skiva|"
    r"knippe|näve|kvist|kvistar|påse|påsar|burk|burkar|port|portioner|"
    r"förp|förpackning|cm|matsked|tesked|nypa|nypor)"
)

# Fraction conversion
FRACTION_MAP = {
    "½": 0.5, "⅓": 0.333, "⅔": 0.667, "¼": 0.25, "¾": 0.75,
    "⅛": 0.125, "⅜": 0.375, "⅝": 0.625, "⅞": 0.875,
}


@dataclass
class Ingredient:
    name: str
    amount: float
    unit: str


@dataclass
class Recipe:
    title: str
    servings: int
    emoji: str
    tags: list[str]
    ingredients: list[Ingredient]
    instructions: list[str]
    source_url: str = ""


def parse_amount(s: str) -> float:
    """Parse Swedish amount string to float."""
    s = s.strip()
    if not s:
        return 1.0

    # Handle unicode fractions
    for frac, val in FRACTION_MAP.items():
        if frac in s:
            # e.g. "2½" -> 2.5
            rest = s.replace(frac, "").strip()
            if rest:
                try:
                    return float(rest) + val
                except ValueError:
                    return val
            return val

    # Handle "1/2", "2 1/2" style fractions
    m = re.match(r"^(\d+)\s+(\d+)/(\d+)$", s)
    if m:
        return int(m.group(1)) + int(m.group(2)) / int(m.group(3))

    m = re.match(r"^(\d+)/(\d+)$", s)
    if m:
        return int(m.group(1)) / int(m.group(2))

    # Handle ranges like "2-3" -> take the first number
    m = re.match(r"^(\d+(?:[.,]\d+)?)\s*[-–]\s*\d+", s)
    if m:
        return float(m.group(1).replace(",", "."))

    # Simple number
    try:
        return float(s.replace(",", "."))
    except ValueError:
        return 1.0


def normalize_unit(unit: str) -> str:
    """Normalize Swedish unit names."""
    unit = unit.lower().strip().rstrip(".")
    mapping = {
        "matsked": "msk",
        "tesked": "tsk",
        "klyfta": "klyftor",
        "skiva": "skivor",
        "kvist": "kvistar",
        "påsar": "påse",
        "burkar": "burk",
        "portioner": "port",
        "förpackning": "förp",
        "nypa": "krm",
        "nypor": "krm",
    }
    return mapping.get(unit, unit)


def clean_ingredient_name(name: str) -> str:
    """Clean up an ingredient name."""
    # Remove parenthetical notes like "(à 400 g)" or "(ca 200 g)"
    name = re.sub(r"\s*\(.*?\)\s*", " ", name)
    # Remove "à X g" suffixes
    name = re.sub(r"\s*à\s+\d+.*$", "", name)
    # Remove "eller ..." alternatives, and any preceding comma-separated fragments
    name = re.sub(r"\s+eller\s+.*$", "", name)
    # If stripping left just an adjective (e.g. "Frysta" from "Frysta eller färska X"), drop it
    name = re.sub(r"^(Frysta|Färska|Tinade|Kylda|Kokta|Torkade)\s*$", "", name, flags=re.IGNORECASE)
    # Remove trailing descriptors like "till garnering", "till servering"
    name = re.sub(r"\s+till\s+(garnering|servering).*$", "", name, flags=re.IGNORECASE)
    # Remove trailing "ev." or leading "ev "/"ev. "
    name = re.sub(r"^[Ee]v\.?\s+", "", name)
    name = re.sub(r"\s+ev\.?\s*$", "", name)
    # Remove ", gärna ..." and " gärna ..." advisory notes
    name = re.sub(r"[,\s]+gärna\b.*$", "", name, flags=re.IGNORECASE)
    # Remove trailing unclosed parenthesis and any content (artifact of à-stripping)
    name = re.sub(r"\s*\([^)]*$", "", name)
    # Remove trailing dash fragments from "eller" stripping: "Bland-, vego-" -> "Blandfärs"
    name = re.sub(r"[,\s]+\w+-$", "", name)
    # Remove trailing dash: "Nöt-" -> "Nötfärs" isn't possible, just strip the dash
    name = name.rstrip("-").strip()
    # Remove leading/trailing whitespace and commas
    name = name.strip().strip(",").strip()
    # Truncate overly long descriptive names (keep first meaningful part)
    if len(name) > 45:
        # Try to cut at a natural break point
        for sep in [",", " med ", " gärna "]:
            idx = name.find(sep)
            if 3 < idx < 45:
                name = name[:idx].strip()
                break
        else:
            name = name[:45].strip()
    # Capitalize first letter
    if name:
        name = name[0].upper() + name[1:]
    return name


def preprocess_ingredient(text: str) -> str:
    """Clean up raw ingredient text before parsing."""
    text = text.strip()
    # Strip leading bullet/dash (may be followed by spaces)
    text = re.sub(r"^[-–•]\s*", "", text)
    # Strip leading "ca " / "Ca " / "ca. " (repeat to handle "- ca X")
    text = re.sub(r"^[Cc]a\.?\s+", "", text)
    # Handle "à" constructs: "2 förp fetaost (à 150 g)" -> "2 förp fetaost"
    text = re.sub(r"\s*à\s+(?!\s*la\s).*$", "", text)
    # Handle weight ranges: "600 - ca 700 g X" -> "650 g X" (midpoint)
    m = re.match(r"^(\d+)\s*[-–]\s*(?:[Cc]a\.?\s+)?(\d+)\s+(g|kg|dl|ml|cl|l)\s+(.+)$", text)
    if m:
        lo, hi = int(m.group(1)), int(m.group(2))
        text = f"{(lo + hi) // 2} {m.group(3)} {m.group(4)}"
    else:
        # Handle count ranges: "2–3 tomater" -> "2 tomater"
        text = re.sub(r"^(\d+)\s*[–-]\s*\d+\s+", r"\1 ", text)
    # Truncate overly long names (descriptive recipe text)
    # This is handled post-parse in clean_ingredient_name
    return text.strip()


def parse_ingredient(text: str) -> Ingredient:
    """Parse a Swedish ingredient string into structured data."""
    text = preprocess_ingredient(text)

    # Fraction pattern: matches "1/2", "2 1/2", "½", "2½"
    FRAC = r"(?:\d+\s+\d+/\d+|\d+/\d+|\d*[½⅓⅔¼¾⅛⅜⅝⅞])"

    # Try explicit fraction + unit: "1/2 dl vitt vin", "2 1/2 dl kycklingbuljong"
    m = re.match(
        rf"^({FRAC})\s+({UNIT_PATTERN})\s+(.+)$",
        text,
        re.IGNORECASE,
    )
    if m:
        amount = parse_amount(m.group(1))
        unit = normalize_unit(m.group(2))
        name = clean_ingredient_name(m.group(3))
        return Ingredient(name=name, amount=amount, unit=unit)

    # Try explicit fraction without unit: "1/2 fiskbuljongtärning"
    m = re.match(rf"^({FRAC})\s+(.+)$", text)
    if m:
        amount = parse_amount(m.group(1))
        name = clean_ingredient_name(m.group(2))
        return Ingredient(name=name, amount=amount, unit="st")

    # Try pattern: amount unit name
    # e.g. "400 g kycklingfilé", "2 msk olivolja", "3 st morötter"
    m = re.match(
        rf"^(\d[\d\s,.½⅓⅔¼¾⅛⅜⅝⅞]*?)\s*({UNIT_PATTERN})\s+(.+)$",
        text,
        re.IGNORECASE,
    )
    if m:
        amount = parse_amount(m.group(1))
        unit = normalize_unit(m.group(2))
        name = clean_ingredient_name(m.group(3))
        return Ingredient(name=name, amount=amount, unit=unit)

    # Try pattern: amount name (no unit) -> "2 ägg", "1 citron"
    m = re.match(r"^(\d[\d\s,.½⅓⅔¼¾⅛⅜⅝⅞]*?)\s+(.+)$", text)
    if m:
        amount = parse_amount(m.group(1))
        name = clean_ingredient_name(m.group(2))
        return Ingredient(name=name, amount=amount, unit="st")

    # Last resort: if text still looks like "amount unit name", try aggressive parsing
    m = re.match(r"^([\d/\s½⅓⅔¼¾⅛⅜⅝⅞,.]+)\s+(dl|g|kg|ml|cl|l|msk|tsk|krm|st)\s+(.+)$", text, re.IGNORECASE)
    if m:
        amount = parse_amount(m.group(1).strip())
        unit = normalize_unit(m.group(2))
        name = clean_ingredient_name(m.group(3))
        return Ingredient(name=name, amount=amount, unit=unit)

    # No amount -> "salt", "peppar"
    name = clean_ingredient_name(text)
    return Ingredient(name=name, amount=1, unit="st")


def pick_emoji(title: str, ingredients: list[str], tags: list[str]) -> str:
    """Pick an emoji based on title, ingredients, and tags."""
    search_text = (title + " " + " ".join(ingredients) + " " + " ".join(tags)).lower()
    for keywords, emoji in EMOJI_KEYWORDS:
        for kw in keywords:
            if kw in search_text:
                return emoji
    return "🍽️"


def parse_servings(yields_str: str) -> int:
    """Parse servings from recipe-scrapers yields string."""
    if not yields_str:
        return 4
    m = re.search(r"(\d+)", yields_str)
    if m:
        return int(m.group(1))
    return 4


def derive_tags(title: str, ingredients: list[str], category_tags: list[str]) -> list[str]:
    """Derive tags from title, ingredients, and category."""
    tags = set(category_tags)
    search_text = (title + " " + " ".join(ingredients)).lower()

    tag_keywords = {
        "kyckling": ["kyckling", "kycklingfilé", "kycklingfärs", "kycklinglår"],
        "fisk": ["lax", "torsk", "fisk", "räkor", "sej", "tonfisk", "pangasius"],
        "pasta": ["pasta", "spaghetti", "penne", "fusilli", "linguine", "makaroner"],
        "soppa": ["soppa"],
        "gryta": ["gryta"],
        "husmanskost": ["köttbullar", "pytt i panna", "falukorv", "ärtsoppa"],
        "snabb": [],  # only from category
        "vegetariskt": ["halloumi", "tofu", "kikärt", "linser", "bönor"],
        "barn": ["pannkak", "fiskpinnar", "korv", "falukorv"],
    }

    for tag, keywords in tag_keywords.items():
        for kw in keywords:
            if kw in search_text:
                tags.add(tag)
                break

    # Remove vegetariskt if meat/fish keywords are present
    meat_keywords = ["kyckling", "fläsk", "nöt", "fårkött", "lamm", "bacon",
                     "korv", "färs", "sidfläsk", "lax", "torsk", "tonfisk",
                     "räkor", "fisk", "sej", "ryggbiff", "blandfärs"]
    if "vegetariskt" in tags:
        if any(kw in search_text for kw in meat_keywords):
            tags.discard("vegetariskt")

    return sorted(tags)


def collect_recipe_urls() -> dict[str, list[str]]:
    """Collect recipe URLs from ICA category pages. Returns {url: category_tags}."""
    url_tags: dict[str, list[str]] = {}
    session = requests.Session()
    session.headers.update({"User-Agent": "Mozilla/5.0 (compatible; recipe-collector)"})

    for cat_path, tags in CATEGORIES.items():
        url = f"https://www.ica.se{cat_path}"
        try:
            resp = session.get(url, timeout=15)
            soup = BeautifulSoup(resp.text, "html.parser")
            count = 0
            for a in soup.find_all("a", href=True):
                href = a["href"]
                if not href.startswith("http"):
                    href = "https://www.ica.se" + href
                href = href.rstrip("/") + "/"
                if re.match(r"https://www\.ica\.se/recept/[a-z0-9-]+-\d+/$", href):
                    if href not in url_tags:
                        url_tags[href] = tags
                    else:
                        # Merge tags from multiple categories
                        url_tags[href] = sorted(set(url_tags[href] + tags))
                    count += 1
            print(f"  {cat_path}: {count} links (total unique: {len(url_tags)})")
        except Exception as e:
            print(f"  {cat_path}: ERROR {e}")
        time.sleep(0.3)

    return url_tags


def scrape_recipe(url: str, category_tags: list[str], session: requests.Session) -> Recipe | None:
    """Scrape a single recipe from ICA.se."""
    try:
        resp = session.get(url, timeout=15)
        if resp.status_code != 200:
            return None
        scraper = scrape_html(resp.text, url)

        title = scraper.title()
        if not title:
            return None

        # Skip if matches existing
        if title.lower().strip() in EXISTING_NAMES:
            return None

        servings = parse_servings(scraper.yields())
        raw_ingredients = scraper.ingredients()
        if not raw_ingredients or len(raw_ingredients) < 2:
            return None

        instructions_text = scraper.instructions()
        if not instructions_text:
            return None

        # Split instructions into steps
        instructions = [
            s.strip()
            for s in re.split(r"\n+", instructions_text)
            if s.strip() and len(s.strip()) > 3
        ]
        if not instructions:
            return None

        # Parse ingredients
        ingredients = [parse_ingredient(ing) for ing in raw_ingredients]
        # Filter out empty names and serving suggestions starting with "Gärna"
        ingredients = [i for i in ingredients if i.name and len(i.name) > 1
                       and not re.match(r'^gärna\b', i.name, re.IGNORECASE)]

        if len(ingredients) < 2:
            return None

        emoji = pick_emoji(title, raw_ingredients, category_tags)
        tags = derive_tags(title, raw_ingredients, category_tags)

        return Recipe(
            title=title,
            servings=servings,
            emoji=emoji,
            tags=tags,
            ingredients=ingredients,
            instructions=instructions,
            source_url=url,
        )
    except Exception as e:
        print(f"    ERROR scraping {url}: {e}")
        return None


def escape_sql(s: str) -> str:
    """Escape a string for SQL single quotes."""
    return s.replace("'", "''")


def recipe_to_sql(recipe: Recipe) -> str:
    """Convert a Recipe to an INSERT OR IGNORE SQL statement."""
    rec_id = f"rec_{uuid.uuid4()}"

    tags_json = json.dumps(recipe.tags, ensure_ascii=False)

    ingredients_json = json.dumps(
        [
            {
                "name": i.name,
                "amount": int(i.amount) if i.amount == int(i.amount) else i.amount,
                "unit": i.unit,
            }
            for i in recipe.ingredients
        ],
        ensure_ascii=False,
    )

    instructions_json = json.dumps(recipe.instructions, ensure_ascii=False)

    return (
        f"INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES\n"
        f"('{rec_id}', '{escape_sql(recipe.title)}', {recipe.servings}, '{recipe.emoji}', "
        f"'{escape_sql(tags_json)}',\n"
        f" '{escape_sql(ingredients_json)}',\n"
        f" '{escape_sql(instructions_json)}');"
    )


def main():
    target = 110  # aim for ~110 to have buffer above 100

    print("Step 1: Collecting recipe URLs from ICA.se categories...")
    url_tags = collect_recipe_urls()
    print(f"\nFound {len(url_tags)} unique recipe URLs\n")

    print("Step 2: Scraping individual recipes...")
    session = requests.Session()
    session.headers.update({"User-Agent": "Mozilla/5.0 (compatible; recipe-collector)"})

    recipes: list[Recipe] = []
    seen_titles: set[str] = set(EXISTING_NAMES)
    errors = 0

    for i, (url, tags) in enumerate(url_tags.items()):
        if len(recipes) >= target:
            break

        recipe = scrape_recipe(url, tags, session)
        if recipe:
            title_key = recipe.title.lower().strip()
            if title_key not in seen_titles:
                seen_titles.add(title_key)
                recipes.append(recipe)
                print(f"  [{len(recipes):3d}/{target}] {recipe.emoji} {recipe.title}")
            else:
                print(f"  [skip] Duplicate: {recipe.title}")
        else:
            errors += 1
            if errors % 10 == 0:
                print(f"  ({errors} errors so far)")

        # Be polite
        time.sleep(0.4)

    print(f"\nScraped {len(recipes)} recipes ({errors} errors)\n")

    if len(recipes) < 50:
        print(f"WARNING: Only got {len(recipes)} recipes, expected 100+")
        print("The migration will still be generated.")

    print("Step 3: Generating SQL migration...")
    output_path = "backend/migrations/012_seed_100_recipes.sql"

    lines = [
        "-- Seed data: 100+ Swedish recipes scraped from ICA.se",
        "-- Auto-generated by tools/scrape_ica_recipes.py",
        f"-- Total recipes: {len(recipes)}",
        "",
    ]

    for recipe in recipes:
        lines.append(recipe_to_sql(recipe))
        lines.append("")

    sql = "\n".join(lines)
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(sql)

    print(f"Written {len(recipes)} recipes to {output_path}")
    print(f"File size: {len(sql):,} bytes")

    # Summary stats
    tag_counts: dict[str, int] = {}
    for r in recipes:
        for t in r.tags:
            tag_counts[t] = tag_counts.get(t, 0) + 1

    print("\nTag distribution:")
    for tag, count in sorted(tag_counts.items(), key=lambda x: -x[1]):
        print(f"  {tag}: {count}")

    emoji_counts: dict[str, int] = {}
    for r in recipes:
        emoji_counts[r.emoji] = emoji_counts.get(r.emoji, 0) + 1

    print("\nEmoji distribution:")
    for emoji, count in sorted(emoji_counts.items(), key=lambda x: -x[1]):
        print(f"  {emoji}: {count}")


if __name__ == "__main__":
    main()
