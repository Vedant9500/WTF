
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

def parse_tldr_page(content):
    lines = content.splitlines()
    command_name = ""
    description = ""
    examples = []
    
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
    
    # We only need metadata for the main entry, but extracting keywords from examples helps
    keywords = set()
    keywords.add(command_name)
    keywords.update(re.split(r'[-\s_.]+', command_name))
    
    # Add words from description (simple stop word filtering could be added here)
    for word in description.split():
        clean_word = re.sub(r'[^\w]', '', word).lower()
        if len(clean_word) > 2:
            keywords.add(clean_word)

    return {
        "command": command_name,
        "description": description,
        "keywords": list(keywords)
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
                        "niche": "general", # Default niche
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
            kw_list = ", ".join(f'"{k}"' for k in entry["keywords"][:8])  # Limit to 8 keywords
            f.write(f'  keywords: [{kw_list}]\n')
            
            f.write(f'  niche: "{entry["niche"]}"\n')
            
            # Platform as inline list without quotes
            plat_list = ", ".join(entry["platform"])
            f.write(f'  platform: [{plat_list}]\n')
            
            f.write(f'  pipeline: {str(entry["pipeline"]).lower()}\n')

if __name__ == "__main__":
    main()
