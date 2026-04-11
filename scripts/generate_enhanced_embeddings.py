"""
generate_enhanced_embeddings.py - Generate enhanced field-aware embeddings for WTF command database

This script generates improved command embeddings with:
1. Field-aware embeddings (command, description, keywords, tags separately)
2. Contextual pooling with attention-like weighting
3. Subword features for OOV recovery
4. Metadata storage for enhanced ranking

Usage:
  python scripts/generate_enhanced_embeddings.py

Prerequisites:
  - Run prepare_glove.py first to generate assets/glove.bin

Output:
  assets/enhanced_cmd_embeddings.bin (~3-5MB)
"""

import os
import re
import struct
import math
import hashlib
import yaml
from pathlib import Path
from collections import Counter

# Configuration
GLOVE_BIN = "assets/glove.bin"
COMMANDS_YAML = "assets/commands.yml"
OUTPUT_FILE = "assets/enhanced_cmd_embeddings.bin"
VECTOR_DIM = 100
EMBED_MAGIC = b"WTFS"  # WTF Search enhanced
EMBED_VERSION = 2

FIELD_WEIGHTS = {
    "command": 3.0,
    "keyword": 2.0,
    "description": 1.0,
    "tag": 1.2,
}

SIF_A = 1e-3
POOLING_TEMPERATURE = 0.5


def load_glove_binary(filepath: str):
    """Load word vectors from binary file."""
    print(f"📖 Loading word vectors from {filepath}...")

    vectors = {}
    ranks = {}
    freqs = {}
    with open(filepath, 'rb') as f:
        vocab_size = struct.unpack('<I', f.read(4))[0]
        print(f"  Vocab size: {vocab_size:,}")

        for i in range(vocab_size):
            word_len = struct.unpack('<H', f.read(2))[0]
            word = f.read(word_len).decode('utf-8')
            vector = list(struct.unpack(f'<{VECTOR_DIM}f', f.read(VECTOR_DIM * 4)))
            vectors[word] = vector
            ranks[word] = i
            # Approximate frequency from rank (lower rank = higher frequency)
            freqs[word] = 1.0 / (i + 1)

            if (i + 1) % 20000 == 0:
                print(f"  Loaded {i + 1:,} / {vocab_size:,} words...")

    print(f"✓ Loaded {len(vectors):,} word vectors")
    return vectors, ranks, freqs


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


def token_variants(token: str) -> list:
    """Generate token variants for OOV recovery."""
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
    """Compute SIF-like token weight."""
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
    """Get vector for token with OOV recovery."""
    for candidate in token_variants(token):
        vec = word_vectors.get(candidate)
        if vec is not None:
            return candidate, vec
    return None, None


def l2_normalize(vec: list) -> list:
    """L2 normalize a vector."""
    norm = math.sqrt(sum(x * x for x in vec))
    if norm == 0.0:
        return vec
    return [x / norm for x in vec]


def attention_pooling(weighted_tokens: list, word_vectors: dict, word_ranks: dict, vocab_size: int) -> list:
    """
    Compute embedding with attention-like pooling.
    Tokens that are more informative get higher weight.
    """
    if not weighted_tokens:
        return [0.0] * VECTOR_DIM

    # Compute initial weights
    token_weights = []
    for token, field_weight in weighted_tokens:
        matched_token, vec = get_vector(token, word_vectors)
        if vec is None:
            continue

        # SIF weighting
        tw = token_weight(matched_token, word_ranks, vocab_size)
        w = field_weight * tw
        token_weights.append((token, vec, w))

    if not token_weights:
        return [0.0] * VECTOR_DIM

    # Apply temperature scaling for attention
    # More distinctive tokens (lower frequency) get higher attention
    attention_weights = []
    for token, vec, w in token_weights:
        # Use inverse frequency as attention signal
        attention = w ** (1.0 / POOLING_TEMPERATURE)
        attention_weights.append((token, vec, attention))

    # Compute weighted sum
    embedding = [0.0] * VECTOR_DIM
    total_attention = sum(att for _, _, att in attention_weights)

    if total_attention == 0.0:
        return [0.0] * VECTOR_DIM

    for token, vec, att in attention_weights:
        normalized_att = att / total_attention
        for i in range(VECTOR_DIM):
            embedding[i] += vec[i] * normalized_att

    return l2_normalize(embedding)


def compute_field_embeddings(command: dict, word_vectors: dict, word_ranks: dict) -> tuple:
    """
    Compute field-aware embeddings for a command.
    Returns (pooled_embedding, field_embeddings)
    """
    vocab_size = len(word_vectors)

    # Extract tokens for each field
    command_tokens = []
    if command.get('command'):
        for t in tokenize(str(command['command'])):
            command_tokens.append((t, FIELD_WEIGHTS["command"]))

    desc_tokens = []
    if command.get('description'):
        for t in tokenize(str(command['description'])):
            desc_tokens.append((t, FIELD_WEIGHTS["description"]))

    keyword_tokens = []
    if command.get('keywords'):
        for kw in command['keywords']:
            for t in tokenize(str(kw)):
                keyword_tokens.append((t, FIELD_WEIGHTS["keyword"]))

    tag_tokens = []
    if command.get('tags'):
        for tag in command['tags']:
            for t in tokenize(str(tag)):
                tag_tokens.append((t, FIELD_WEIGHTS["tag"]))

    # Compute field-specific embeddings
    cmd_embed = attention_pooling(command_tokens, word_vectors, word_ranks, vocab_size) if command_tokens else None
    desc_embed = attention_pooling(desc_tokens, word_vectors, word_ranks, vocab_size) if desc_tokens else None
    keyword_embed = attention_pooling(keyword_tokens, word_vectors, word_ranks, vocab_size) if keyword_tokens else None
    tag_embed = attention_pooling(tag_tokens, word_vectors, word_ranks, vocab_size) if tag_tokens else None

    # Compute pooled embedding (weighted average of field embeddings)
    all_tokens = command_tokens + desc_tokens + keyword_tokens + tag_tokens
    pooled_embed = attention_pooling(all_tokens, word_vectors, word_ranks, vocab_size) if all_tokens else [0.0] * VECTOR_DIM

    field_embeddings = {
        'command': cmd_embed,
        'description': desc_embed,
        'keyword': keyword_embed,
        'tag': tag_embed,
    }

    return pooled_embed, field_embeddings


def save_enhanced_embeddings(commands: list, embeddings: list, field_embeddings: list, 
                              command_hash: str, output_path: str):
    """Save enhanced command embeddings in binary format."""
    print(f"💾 Saving enhanced embeddings to {output_path}...")

    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    with open(output_path, 'wb') as f:
        # Write header
        f.write(EMBED_MAGIC)
        f.write(struct.pack('<H', EMBED_VERSION))
        f.write(struct.pack('<I', len(embeddings)))
        f.write(struct.pack('<I', VECTOR_DIM))

        # Write each command's embeddings
        for i, (cmd, pooled, field_embeds) in enumerate(zip(commands, embeddings, field_embeddings)):
            # Write metadata
            cmd_name = str(cmd.get('command', ''))
            f.write(struct.pack('<H', len(cmd_name)))
            f.write(cmd_name.encode('utf-8'))

            # Write platforms
            platforms = cmd.get('platform', [])
            f.write(struct.pack('<B', len(platforms)))
            for platform in platforms:
                f.write(struct.pack('<B', len(platform)))
                f.write(platform.encode('utf-8'))

            # Write niche
            niche = str(cmd.get('niche', ''))
            f.write(struct.pack('<H', len(niche)))
            f.write(niche.encode('utf-8'))

            # Write pipeline flag
            is_pipeline = bool(cmd.get('pipeline', False))
            f.write(struct.pack('<?', is_pipeline))

            # Write pooled embedding
            f.write(struct.pack(f'<{VECTOR_DIM}f', *pooled))

            # Write field embeddings (with presence flag)
            has_field_embeds = any(v is not None for v in field_embeds.values())
            f.write(struct.pack('<B', 1 if has_field_embeds else 0))

            if has_field_embeds:
                # Write command field embedding
                if field_embeds['command']:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *field_embeds['command']))
                else:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *[0.0] * VECTOR_DIM))

                # Write description field embedding
                if field_embeds['description']:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *field_embeds['description']))
                else:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *[0.0] * VECTOR_DIM))

                # Write keyword field embedding
                if field_embeds['keyword']:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *field_embeds['keyword']))
                else:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *[0.0] * VECTOR_DIM))

                # Write tag field embedding
                if field_embeds['tag']:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *field_embeds['tag']))
                else:
                    f.write(struct.pack(f'<{VECTOR_DIM}f', *[0.0] * VECTOR_DIM))

    file_size = os.path.getsize(output_path)
    print(f"✓ Saved {output_path} ({file_size / (1024*1024):.2f} MB)")


def verify_enhanced_embeddings(filepath: str, num_samples: int = 5):
    """Verify the enhanced binary file can be read correctly."""
    print(f"🔍 Verifying enhanced embeddings file...")

    with open(filepath, 'rb') as f:
        magic = f.read(4)
        if magic != EMBED_MAGIC:
            print(f"  ❌ Invalid magic: {magic}")
            return

        version = struct.unpack('<H', f.read(2))[0]
        num_commands = struct.unpack('<I', f.read(4))[0]
        dimension = struct.unpack('<I', f.read(4))[0]

        print(f"  Format: enhanced v{version}, Commands: {num_commands:,}, Dimension: {dimension}")

        # Read first few embeddings
        for i in range(min(num_samples, num_commands)):
            # Skip metadata (variable length)
            cmd_len = struct.unpack('<H', f.read(2))[0]
            cmd_name = f.read(cmd_len).decode('utf-8')

            num_platforms = struct.unpack('<B', f.read(1))[0]
            for _ in range(num_platforms):
                plat_len = struct.unpack('<B', f.read(1))[0]
                f.read(plat_len)

            niche_len = struct.unpack('<H', f.read(2))[0]
            f.read(niche_len)

            f.read(1)  # pipeline flag

            # Read pooled embedding
            embedding = struct.unpack(f'<{dimension}f', f.read(dimension * 4))
            norm = sum(x*x for x in embedding) ** 0.5
            print(f"  Command {i} '{cmd_name}': norm={norm:.4f}, values[0:3]={embedding[0]:.4f}, {embedding[1]:.4f}, {embedding[2]:.4f}")

            # Skip field embeddings
            has_fields = struct.unpack('<B', f.read(1))[0]
            if has_fields:
                f.read(VECTOR_DIM * 4 * 4)  # 4 field embeddings

    print("✓ Enhanced embeddings verified")


def main():
    print("=" * 60)
    print("Enhanced Command Embedding Generator for WTF CLI")
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
    word_vectors, word_ranks, word_freqs = load_glove_binary(GLOVE_BIN)
    print()

    # Step 2: Load commands
    commands = load_commands(COMMANDS_YAML)
    snapshot_hash = command_snapshot_hash(commands)
    print()

    # Step 3: Compute enhanced embeddings
    print("🧮 Computing enhanced command embeddings...")
    embeddings = []
    field_embeddings_list = []
    zero_count = 0

    for i, cmd in enumerate(commands):
        pooled_embed, field_embeds = compute_field_embeddings(cmd, word_vectors, word_ranks)
        embeddings.append(pooled_embed)
        field_embeddings_list.append(field_embeds)

        # Check if zero vector (no matching words)
        if all(x == 0.0 for x in pooled_embed):
            zero_count += 1

        if (i + 1) % 500 == 0:
            print(f"  Processed {i + 1:,} / {len(commands):,} commands...")

    print(f"✓ Computed {len(embeddings):,} enhanced embeddings")
    print(f"  (Note: {zero_count} commands had no matching words in vocabulary)")
    print()

    # Step 4: Save embeddings
    save_enhanced_embeddings(commands, embeddings, field_embeddings_list, snapshot_hash, OUTPUT_FILE)
    print()

    # Step 5: Verify
    verify_enhanced_embeddings(OUTPUT_FILE)
    print()

    print("=" * 60)
    print("✅ Done! Enhanced command embeddings ready for WTF CLI")
    print(f"   Output: {OUTPUT_FILE}")
    print("=" * 60)


if __name__ == "__main__":
    main()
