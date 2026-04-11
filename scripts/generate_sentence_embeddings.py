"""
generate_sentence_embeddings.py - Generate sentence-transformer embeddings for WTF command database

This script replaces GloVe-based embeddings with sentence-transformers that capture
full command context and semantics.

Usage:
  pip install sentence-transformers
  python scripts/generate_sentence_embeddings.py

Output:
  assets/sentence_cmd_embeddings.bin (~8-10MB for all-MiniLM-L6-v2)
"""

import os
import re
import struct
import math
import hashlib
import yaml
from pathlib import Path

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print("❌ sentence-transformers not installed!")
    print("   Install with: pip install sentence-transformers")
    exit(1)

# Configuration
COMMANDS_YAML = "assets/commands.yml"
OUTPUT_FILE = "assets/sentence_cmd_embeddings.bin"
MODEL_NAME = "all-MiniLM-L6-v2"  # 22MB, fast, good quality
EMBED_MAGIC = b"WTFS"  # WTF Search enhanced
EMBED_VERSION = 3  # New version for sentence-transformers


def load_model(model_name: str):
    """Load sentence-transformer model."""
    print(f"📦 Loading sentence-transformer model: {model_name}...")
    model = SentenceTransformer(model_name)
    print(f"✓ Model loaded (dimension: {model.get_sentence_embedding_dimension()})")
    return model


def load_commands(filepath: str) -> list:
    """Load commands from YAML file."""
    print(f"📖 Loading commands from {filepath}...")

    with open(filepath, 'r', encoding='utf-8') as f:
        data = yaml.safe_load(f)

    # Handle both formats: direct list or dict with 'commands' key
    if isinstance(data, list):
        commands = data
    elif isinstance(data, dict):
        commands = data.get('commands', [])
    else:
        commands = []

    print(f"✓ Loaded {len(commands):,} commands")
    return commands


def normalize_text(value: str) -> str:
    """Normalize text for hashing."""
    return " ".join(str(value).strip().lower().split())


def command_snapshot_hash(commands: list) -> str:
    """Hash the command snapshot for alignment checks."""
    hasher = hashlib.sha256()
    for cmd in commands:
        command = normalize_text(cmd.get("command", ""))
        description = normalize_text(cmd.get("description", ""))
        keywords = cmd.get("keywords") or []
        keyword_text = "\x1e".join(normalize_text(k) for k in keywords)
        tags = cmd.get("tags") or []
        tag_text = "\x1e".join(normalize_text(t) for t in tags)

        hasher.update(command.encode("utf-8"))
        hasher.update(b"\x1f")
        hasher.update(description.encode("utf-8"))
        hasher.update(b"\x1f")
        hasher.update(keyword_text.encode("utf-8"))
        hasher.update(b"\x1f")
        hasher.update(tag_text.encode("utf-8"))
        hasher.update(b"\x1d")

    return hasher.hexdigest()


def prepare_command_text(command: dict) -> str:
    """
    Prepare command text for embedding.
    Combine command, description, keywords, and tags with structure.
    """
    parts = []

    # Command name (most important)
    if command.get('command'):
        parts.append(f"command: {command['command']}")

    # Description
    if command.get('description'):
        parts.append(f"description: {command['description']}")

    # Keywords
    if command.get('keywords'):
        keywords_text = ", ".join(command['keywords'])
        parts.append(f"keywords: {keywords_text}")

    # Tags
    if command.get('tags'):
        tags_text = ", ".join(command['tags'])
        parts.append(f"tags: {tags_text}")

    return " | ".join(parts)


def generate_embeddings(model, commands: list) -> list:
    """Generate sentence embeddings for all commands."""
    print(f"🧮 Generating sentence embeddings...")

    # Prepare all command texts
    command_texts = []
    for cmd in commands:
        text = prepare_command_text(cmd)
        command_texts.append(text)

    # Generate embeddings in batches
    embeddings = model.encode(
        command_texts,
        batch_size=64,
        show_progress_bar=True,
        normalize_embeddings=True,  # L2 normalize for cosine similarity
        convert_to_numpy=True
    )

    print(f"✓ Generated {len(embeddings):,} embeddings (dim={embeddings.shape[1]})")
    return embeddings


def save_embeddings(embeddings, command_hash: str, output_path: str, commands: list):
    """Save sentence embeddings in binary format."""
    print(f"💾 Saving embeddings to {output_path}...")

    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    num_commands, dimension = embeddings.shape

    with open(output_path, 'wb') as f:
        # Write header
        f.write(EMBED_MAGIC)
        f.write(struct.pack('<H', EMBED_VERSION))
        f.write(struct.pack('<I', num_commands))
        f.write(struct.pack('<I', dimension))

        # Write command hash
        hash_bytes = command_hash.encode('utf-8')
        f.write(struct.pack('<H', len(hash_bytes)))
        f.write(hash_bytes)

        # Write embeddings
        for i in range(num_commands):
            f.write(struct.pack(f'<{dimension}f', *embeddings[i]))

    file_size = os.path.getsize(output_path)
    print(f"✓ Saved {output_path} ({file_size / (1024*1024):.2f} MB)")


def verify_embeddings(filepath: str, dimension: int, num_samples: int = 5):
    """Verify the binary file can be read correctly."""
    print(f"🔍 Verifying embeddings file...")

    with open(filepath, 'rb') as f:
        magic = f.read(4)
        if magic != EMBED_MAGIC:
            print(f"  ❌ Invalid magic: {magic}")
            return

        version = struct.unpack('<H', f.read(2))[0]
        num_commands = struct.unpack('<I', f.read(4))[0]
        dim = struct.unpack('<I', f.read(4))[0]
        hash_len = struct.unpack('<H', f.read(2))[0]
        command_hash = f.read(hash_len).decode('utf-8') if hash_len > 0 else ""

        print(f"  Format: sentence v{version}")
        print(f"  Commands: {num_commands:,}, Dimension: {dim}")
        if command_hash:
            print(f"  Command hash: {command_hash[:16]}...")

        # Read first few embeddings
        for i in range(min(num_samples, num_commands)):
            embedding = struct.unpack(f'<{dim}f', f.read(dimension * 4))
            norm = sum(x*x for x in embedding) ** 0.5
            print(f"  Embedding {i}: norm={norm:.4f}, values[0:3]={embedding[0]:.4f}, {embedding[1]:.4f}, {embedding[2]:.4f}")

    print("✓ Embeddings verified")


def main():
    print("=" * 60)
    print("Sentence-Transformer Embedding Generator for WTF CLI")
    print("=" * 60)
    print()

    # Check prerequisites
    if not os.path.exists(COMMANDS_YAML):
        print(f"❌ Error: {COMMANDS_YAML} not found!")
        return

    # Step 1: Load model
    model = load_model(MODEL_NAME)
    dimension = model.get_sentence_embedding_dimension()
    print()

    # Step 2: Load commands
    commands = load_commands(COMMANDS_YAML)
    snapshot_hash = command_snapshot_hash(commands)
    print()

    # Step 3: Generate embeddings
    embeddings = generate_embeddings(model, commands)
    print()

    # Step 4: Save embeddings
    save_embeddings(embeddings, snapshot_hash, OUTPUT_FILE, commands)
    print()

    # Step 5: Verify
    verify_embeddings(OUTPUT_FILE, dimension)
    print()

    print("=" * 60)
    print("✅ Done! Sentence-transformer embeddings ready for WTF CLI")
    print(f"   Model: {MODEL_NAME}")
    print(f"   Output: {OUTPUT_FILE}")
    print(f"   Dimension: {dimension}")
    print("=" * 60)


if __name__ == "__main__":
    main()
