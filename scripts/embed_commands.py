"""
embed_commands.py - Generate embeddings for WTF command database

This script:
1. Loads the glove.bin binary file
2. Reads commands.yml
3. For each command, creates an embedding by averaging word vectors
   from: command + description + keywords
4. Saves embeddings in binary format for Go

Output format (cmd_embeddings.bin):
  [num_commands: uint32]
  [dimension: uint32]
  For each command:
    [embedding: dimension * float32]

Usage:
  python scripts/embed_commands.py

Prerequisites:
  - Run prepare_glove.py first to generate assets/glove.bin

Output:
  assets/cmd_embeddings.bin (~1.5MB)
"""

import os
import re
import struct
import math
import hashlib
import yaml
from pathlib import Path

# Configuration
GLOVE_BIN = "assets/glove.bin"
COMMANDS_YAML = "assets/commands.yml"
OUTPUT_FILE = "assets/cmd_embeddings.bin"
VECTOR_DIM = 100
EMBED_MAGIC = b"WTFE"
EMBED_VERSION = 1

FIELD_WEIGHTS = {
    "command": 3.0,
    "keyword": 2.0,
    "description": 1.0,
}

SIF_A = 1e-3


def load_glove_binary(filepath: str):
    """Load word vectors from binary file."""
    print(f"📖 Loading word vectors from {filepath}...")
    
    vectors = {}
    ranks = {}
    with open(filepath, 'rb') as f:
        vocab_size = struct.unpack('<I', f.read(4))[0]
        print(f"  Vocab size: {vocab_size:,}")
        
        for i in range(vocab_size):
            word_len = struct.unpack('<H', f.read(2))[0]
            word = f.read(word_len).decode('utf-8')
            vector = list(struct.unpack(f'<{VECTOR_DIM}f', f.read(VECTOR_DIM * 4)))
            vectors[word] = vector
            ranks[word] = i
            
            if (i + 1) % 20000 == 0:
                print(f"  Loaded {i + 1:,} / {vocab_size:,} words...")
    
    print(f"✓ Loaded {len(vectors):,} word vectors")
    return vectors, ranks


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


def tokenize(text: str) -> list:
    """Simple tokenization for embedding lookup."""
    if not text:
        return []
    
    # Lowercase, remove special chars, split
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', ' ', text)
    tokens = text.split()
    
    # Filter very short tokens
    tokens = [t for t in tokens if len(t) >= 2]
    
    return tokens


def normalize_text(value: str) -> str:
    return " ".join(str(value).strip().lower().split())


def command_snapshot_hash(commands: list) -> str:
    """Hash the command snapshot used to generate embeddings for runtime alignment checks."""
    hasher = hashlib.sha256()
    for cmd in commands:
        command = normalize_text(cmd.get("command", ""))
        description = normalize_text(cmd.get("description", ""))
        keywords = cmd.get("keywords") or []
        keyword_text = "\x1e".join(normalize_text(k) for k in keywords)

        hasher.update(command.encode("utf-8"))
        hasher.update(b"\x1f")
        hasher.update(description.encode("utf-8"))
        hasher.update(b"\x1f")
        hasher.update(keyword_text.encode("utf-8"))
        hasher.update(b"\x1d")

    return hasher.hexdigest()


def token_variants(token: str) -> list:
    variants = [token]
    if token.endswith("ing") and len(token) > 5:
        variants.append(token[:-3])
    if token.endswith("ed") and len(token) > 4:
        variants.append(token[:-2])
    if token.endswith("es") and len(token) > 4:
        variants.append(token[:-2])
    if token.endswith("s") and len(token) > 3:
        variants.append(token[:-1])

    unique = []
    seen = set()
    for v in variants:
        if v and v not in seen:
            seen.add(v)
            unique.append(v)
    return unique


def token_weight(token: str, ranks: dict, vocab_size: int) -> float:
    weight = 1.0
    rank = ranks.get(token)
    if rank is not None and vocab_size > 0:
        p = (rank + 1) / float(vocab_size)
        weight *= SIF_A / (SIF_A + p)

    if any(ch.isdigit() for ch in token):
        weight *= 1.2
    if len(token) <= 2:
        weight *= 0.85

    return weight


def get_vector(token: str, word_vectors: dict):
    for candidate in token_variants(token):
        vec = word_vectors.get(candidate)
        if vec is not None:
            return candidate, vec
    return None, None


def l2_normalize(vec: list) -> list:
    norm = math.sqrt(sum(x * x for x in vec))
    if norm == 0.0:
        return vec
    return [x / norm for x in vec]


def compute_embedding(command: dict, word_vectors: dict, word_ranks: dict) -> list:
    """Compute embedding via weighted pooling of command/keyword/description tokens."""
    weighted_tokens = []

    if command.get('command'):
        for t in tokenize(str(command['command'])):
            weighted_tokens.append((t, FIELD_WEIGHTS["command"]))

    if command.get('keywords'):
        for kw in command['keywords']:
            for t in tokenize(str(kw)):
                weighted_tokens.append((t, FIELD_WEIGHTS["keyword"]))

    if command.get('description'):
        for t in tokenize(str(command['description'])):
            weighted_tokens.append((t, FIELD_WEIGHTS["description"]))

    if not weighted_tokens:
        return [0.0] * VECTOR_DIM

    token_counts = {}
    for token, _ in weighted_tokens:
        token_counts[token] = token_counts.get(token, 0) + 1

    embedding = [0.0] * VECTOR_DIM
    total_weight = 0.0
    vocab_size = len(word_vectors)

    for token, field_weight in weighted_tokens:
        matched_token, vec = get_vector(token, word_vectors)
        if vec is None:
            continue

        repeat_discount = 1.0 / math.sqrt(token_counts[token])
        tw = token_weight(matched_token, word_ranks, vocab_size)
        w = field_weight * repeat_discount * tw

        for i in range(VECTOR_DIM):
            embedding[i] += vec[i] * w
        total_weight += w

    if total_weight == 0.0:
        return [0.0] * VECTOR_DIM

    embedding = [x / total_weight for x in embedding]
    return l2_normalize(embedding)


def save_embeddings(embeddings: list, command_hash: str, output_path: str):
    """Save command embeddings in binary format."""
    print(f"💾 Saving embeddings to {output_path}...")
    
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, 'wb') as f:
        # Header:
        #   magic[4] + version[u16] + reserved[u16] + num_commands[u32] + dimension[u32]
        #   hash_len[u16] + command_hash[bytes]
        f.write(EMBED_MAGIC)
        f.write(struct.pack('<H', EMBED_VERSION))
        f.write(struct.pack('<H', 0))
        f.write(struct.pack('<I', len(embeddings)))
        f.write(struct.pack('<I', VECTOR_DIM))
        hash_bytes = command_hash.encode('utf-8')
        f.write(struct.pack('<H', len(hash_bytes)))
        f.write(hash_bytes)
        
        # Write each embedding
        for embedding in embeddings:
            f.write(struct.pack(f'<{VECTOR_DIM}f', *embedding))
    
    file_size = os.path.getsize(output_path)
    print(f"✓ Saved {output_path} ({file_size / (1024*1024):.2f} MB)")


def verify_embeddings(filepath: str, num_samples: int = 5):
    """Verify the binary file can be read correctly."""
    print(f"🔍 Verifying embeddings file...")
    
    with open(filepath, 'rb') as f:
        prefix = f.read(4)
        if prefix == EMBED_MAGIC:
            version = struct.unpack('<H', f.read(2))[0]
            _reserved = struct.unpack('<H', f.read(2))[0]
            num_commands = struct.unpack('<I', f.read(4))[0]
            dimension = struct.unpack('<I', f.read(4))[0]
            hash_len = struct.unpack('<H', f.read(2))[0]
            command_hash = f.read(hash_len).decode('utf-8') if hash_len > 0 else ""
            print(f"  Format: metadata v{version}, Commands: {num_commands:,}, Dimension: {dimension}")
            if command_hash:
                print(f"  Command hash: {command_hash[:16]}...")
        else:
            # Legacy format support
            num_commands = struct.unpack('<I', prefix)[0]
            dimension = struct.unpack('<I', f.read(4))[0]
            print(f"  Format: legacy, Commands: {num_commands:,}, Dimension: {dimension}")
        
        # Read first few embeddings
        for i in range(min(num_samples, num_commands)):
            embedding = struct.unpack(f'<{dimension}f', f.read(dimension * 4))
            norm = sum(x*x for x in embedding) ** 0.5
            print(f"  Embedding {i}: norm={norm:.4f}, values[0:3]={embedding[0]:.4f}, {embedding[1]:.4f}, {embedding[2]:.4f}")
    
    print("✓ Embeddings verified")


def main():
    print("=" * 60)
    print("Command Embedding Generator for WTF CLI")
    print("=" * 60)
    print()
    
    # Check prerequisites
    if not os.path.exists(GLOVE_BIN):
        print(f"❌ Error: {GLOVE_BIN} not found!")
        print("   Please run prepare_glove.py first.")
        return
    
    if not os.path.exists(COMMANDS_YAML):
        print(f"❌ Error: {COMMANDS_YAML} not found!")
        return
    
    # Step 1: Load word vectors
    word_vectors, word_ranks = load_glove_binary(GLOVE_BIN)
    print()
    
    # Step 2: Load commands
    commands = load_commands(COMMANDS_YAML)
    snapshot_hash = command_snapshot_hash(commands)
    print()
    
    # Step 3: Compute embeddings
    print("🧮 Computing command embeddings...")
    embeddings = []
    zero_count = 0
    
    for i, cmd in enumerate(commands):
        embedding = compute_embedding(cmd, word_vectors, word_ranks)
        embeddings.append(embedding)
        
        # Check if zero vector (no matching words)
        if all(x == 0.0 for x in embedding):
            zero_count += 1
        
        if (i + 1) % 500 == 0:
            print(f"  Processed {i + 1:,} / {len(commands):,} commands...")
    
    print(f"✓ Computed {len(embeddings):,} embeddings")
    print(f"  (Note: {zero_count} commands had no matching words in vocabulary)")
    print()
    
    # Step 4: Save embeddings
    save_embeddings(embeddings, snapshot_hash, OUTPUT_FILE)
    print()
    
    # Step 5: Verify
    verify_embeddings(OUTPUT_FILE)
    print()
    
    print("=" * 60)
    print("✅ Done! Command embeddings ready for WTF CLI")
    print(f"   Output: {OUTPUT_FILE}")
    print("=" * 60)


if __name__ == "__main__":
    main()
