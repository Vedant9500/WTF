
import os
import requests
import zipfile
import io
import yaml
import re
from tqdm import tqdm

TLDR_ZIP_URL = "https://github.com/tldr-pages/tldr/archive/refs/heads/main.zip"
OUTPUT_FILE = os.path.join("assets", "commands.yml")

# Mapping TLDR platforms to WTF platforms
PLATFORM_MAP = {
    "common": ["linux", "macos", "windows"],
    "linux": ["linux"],
    "osx": ["macos"],
    "windows": ["windows"],
    "android": ["android"],
    "sunos": ["linux"], # Approximate
    "freebsd": ["linux"], # Approximate
    "netbsd": ["linux"], # Approximate
    "openbsd": ["linux"] # Approximate
}

STOP_WORDS = {
    "the", "a", "an", "and", "or", "to", "for", "of", "in", "on", "with",
    "from", "into", "by", "at", "is", "are", "be", "this", "that", "as",
    "using", "use", "via", "over", "under", "all", "any", "more", "information",
    "see", "also", "about", "through", "without", "across"
}

def download_tldr_zip():
    print(f"Downloading TLDR archive from {TLDR_ZIP_URL}...")
    response = requests.get(TLDR_ZIP_URL, stream=True)
    total_size = int(response.headers.get('content-length', 0))
    
    buffer = io.BytesIO()
    with tqdm(total=total_size, unit='B', unit_scale=True, desc="Downloading") as pbar:
        for data in response.iter_content(1024):
            buffer.write(data)
            pbar.update(len(data))
    
    return zipfile.ZipFile(buffer)


def normalize_token(token: str) -> str:
    token = re.sub(r"[^a-zA-Z0-9_-]", "", token).lower()
    return token.strip("-_")


def extract_words(text: str, min_len: int = 3):
    words = set()
    for word in re.split(r"[^a-zA-Z0-9_-]+", text.lower()):
        clean = normalize_token(word)
        if len(clean) >= min_len and clean not in STOP_WORDS:
            words.add(clean)
    return words


def command_family(command_name: str) -> str:
    if not command_name:
        return "general"
    # TLDR command names are usually command identifiers like "git commit" or "docker image rm".
    first = command_name.split()[0]
    first = normalize_token(first.split("/")[-1])
    return first or "general"


def extract_aliases(description: str, command_name: str):
    aliases = set()
    cmd_family = command_family(command_name)
    for alias in re.findall(r"`([^`]+)`", description):
        alias = alias.strip()
        if not alias:
            continue
        base = command_family(alias)
        # Keep close aliases/tool names, avoid long inline examples.
        if base and base != cmd_family and len(alias.split()) <= 3:
            aliases.add(base)
    return aliases


def parse_examples(lines):
    examples = []
    i = 0
    while i < len(lines):
        line = lines[i]
        if line.startswith("- "):
            example_desc = line[2:].strip()
            example_cmd = ""
            j = i + 1
            while j < len(lines):
                candidate = lines[j].strip()
                if candidate.startswith("-"):
                    break
                if "`" in candidate:
                    first = candidate.find("`")
                    last = candidate.rfind("`")
                    if last > first:
                        example_cmd = candidate[first + 1:last].strip()
                        break
                j += 1
            examples.append((example_desc, example_cmd))
            i = j
            continue
        i += 1
    return examples


def extract_example_metadata(examples):
    verbs = set()
    objects = set()
    cmd_tokens = set()

    for desc, cmd in examples:
        desc_words = [w for w in re.split(r"[^a-zA-Z0-9_-]+", desc.lower()) if w]
        if desc_words:
            v = normalize_token(desc_words[0])
            if len(v) >= 3 and v not in STOP_WORDS:
                verbs.add(v)
            for w in desc_words[1:]:
                obj = normalize_token(w)
                if len(obj) >= 3 and obj not in STOP_WORDS:
                    objects.add(obj)

        if cmd:
            for t in re.split(r"[^a-zA-Z0-9_-]+", cmd.lower()):
                clean = normalize_token(t)
                if len(clean) >= 2 and clean not in STOP_WORDS:
                    cmd_tokens.add(clean)

    return verbs, objects, cmd_tokens

def parse_tldr_page(content):
    lines = content.splitlines()
    command_name = ""
    description = ""
    
    # Simple parser for TLDR format
    # > Command Name
    # 
    # > Description
    # 
    # - Example description
    # 
    #   `command example`
    
    if len(lines) > 0:
        command_name = lines[0].replace("# ", "").strip()
        
    desc_lines = []
    for line in lines[1:]:
        if line.startswith("> "):
            desc_lines.append(line[2:].strip())
        elif line.startswith("-"):
            break # Start of examples
            
    description = " ".join(desc_lines)
    
    examples = parse_examples(lines)

    # Build searchable metadata: keywords + richer tags (aliases, family, TLDR example verbs/objects).
    keywords = set()
    tags = set()

    keywords.add(command_name.lower())
    keywords.update(re.split(r'[-\s_.]+', command_name))

    fam = command_family(command_name)
    tags.add(f"family:{fam}")
    keywords.add(fam)
    
    keywords.update(extract_words(description))

    aliases = extract_aliases(description, command_name)
    for alias in aliases:
        keywords.add(alias)
        tags.add(f"alias:{alias}")

    verbs, objects, cmd_tokens = extract_example_metadata(examples)
    keywords.update(verbs)
    keywords.update(objects)
    keywords.update(cmd_tokens)
    for v in verbs:
        tags.add(f"verb:{v}")
    # keep object tag fan-out bounded
    for obj in sorted(objects)[:8]:
        tags.add(f"object:{obj}")

    return {
        "command": command_name,
        "description": description,
        "keywords": list(keywords),
        "tags": list(tags),
        "family": fam,
        "aliases": list(aliases),
    }

def main():
    if not os.path.exists("assets"):
        os.makedirs("assets")

    try:
        zip_ref = download_tldr_zip()
    except Exception as e:
        print(f"Failed to download TLDR archive: {e}")
        return

    commands_db = {}
    
    print("Processing TLDR pages...")
    # Walk through zip file
    for file_info in tqdm(zip_ref.infolist(), desc="Parsing"):
        if not file_info.filename.endswith(".md"):
            continue
            
        # Check if it's in a pages directory
        # tldr-main/pages/common/git.md
        parts = file_info.filename.split("/")
        if "pages" not in parts:
            continue
            
        try:
            pages_index = parts.index("pages")
            if pages_index + 2 >= len(parts):
                continue
                
            platform_dir = parts[pages_index + 1]
            filename = parts[pages_index + 2]
            
            if platform_dir not in PLATFORM_MAP:
                continue
                
            with zip_ref.open(file_info) as f:
                content = f.read().decode('utf-8')
                data = parse_tldr_page(content)
                
                cmd_name = data["command"]
                
                if cmd_name not in commands_db:
                    commands_db[cmd_name] = {
                        "command": cmd_name,
                        "description": data["description"],
                        "keywords": data["keywords"],
                        "tags": data.get("tags", []),
                        "niche": data.get("family", "general"),
                        "platform": set(),
                        "pipeline": False
                    }
                
                # Merge platforms
                target_platforms = PLATFORM_MAP[platform_dir]
                commands_db[cmd_name]["platform"].update(target_platforms)
                
                # Merge keywords
                existing_keywords = set(commands_db[cmd_name]["keywords"])
                existing_keywords.update(data["keywords"])
                commands_db[cmd_name]["keywords"] = list(existing_keywords)

                existing_tags = set(commands_db[cmd_name].get("tags", []))
                existing_tags.update(data.get("tags", []))
                commands_db[cmd_name]["tags"] = list(existing_tags)
                
        except Exception as e:
            # print(f"Error parsing {file_info.filename}: {e}")
            continue

    # Convert to list and cleanup
    final_list = []
    print("Finalizing database...")
    for cmd_name in sorted(commands_db.keys()):
        entry = commands_db[cmd_name]
        entry["platform"] = sorted(list(entry["platform"]))
        entry["keywords"] = sorted(list(entry["keywords"]))
        entry["tags"] = sorted(list(entry.get("tags", [])))
        final_list.append(entry)

    print(f"Found {len(final_list)} unique commands.")
    
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
        # Write in the exact format of the backup file:
        # - command: "..."
        #   description: "..."
        #   keywords: [...]
        #   niche: "..."
        #   platform: [...]
        #   pipeline: false
        
        for entry in final_list:
            # Quote command if it has special chars
            cmd = entry["command"]
            desc = entry["description"]
            
            # Escape backslashes FIRST, then quotes
            cmd_escaped = cmd.replace('\\', '\\\\').replace('"', '\\"')
            desc_escaped = desc.replace('\\', '\\\\').replace('"', '\\"')
            
            f.write(f'- command: "{cmd_escaped}"\n')
            f.write(f'  description: "{desc_escaped}"\n')
            
            # Keywords as inline list
            kw_list = ", ".join(f'"{k}"' for k in entry["keywords"])
            f.write(f'  keywords: [{kw_list}]\n')

            if entry.get("tags"):
                tag_list = ", ".join(f'"{t}"' for t in entry["tags"])
                f.write(f'  tags: [{tag_list}]\n')
            
            f.write(f'  niche: "{entry["niche"]}"\n')
            
            # Platform as inline list without quotes
            plat_list = ", ".join(entry["platform"])
            f.write(f'  platform: [{plat_list}]\n')
            
            f.write(f'  pipeline: {str(entry["pipeline"]).lower()}\n')

if __name__ == "__main__":
    main()
