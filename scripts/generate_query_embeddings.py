"""
generate_query_embeddings.py - Generate sentence-transformer embeddings for eval queries

Usage:
  python scripts/generate_query_embeddings.py
"""

import os
import sys
import json
import yaml
from pathlib import Path

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print("❌ sentence-transformers not installed!")
    print("   Install with: pip install sentence-transformers")
    sys.exit(1)

MODEL_NAME = "all-MiniLM-L6-v2"
EVAL_QUERIES_SHORT = "assets/eval_queries.yaml"
EVAL_QUERIES_LONG = "assets/eval_queries_long.yaml"
OUTPUT_FILE = "assets/query_embeddings.bin"


def load_all_queries():
    """Load all eval queries from both YAML files."""
    queries = []
    for path in [EVAL_QUERIES_SHORT, EVAL_QUERIES_LONG]:
        with open(path, 'r', encoding='utf-8') as f:
            data = yaml.safe_load(f)
        queries.extend([q['query'] for q in data.get('queries', [])])
    # Deduplicate while preserving order
    seen = set()
    unique = []
    for q in queries:
        if q not in seen:
            seen.add(q)
            unique.append(q)
    return unique


def main():
    print("=" * 60)
    print("Query Embedding Generator")
    print("=" * 60)
    print()

    # Load model
    print(f"📦 Loading model: {MODEL_NAME}...")
    model = SentenceTransformer(MODEL_NAME)
    dim = model.get_sentence_embedding_dimension()
    print(f"✓ Model loaded (dim={dim})")
    print()

    # Load queries
    all_queries = load_all_queries()
    
    print(f"📖 Loaded {len(all_queries)} unique queries")
    print()

    # Generate embeddings
    print("🧮 Generating query embeddings...")
    embeddings = model.encode(
        all_queries,
        batch_size=32,
        show_progress_bar=True,
        normalize_embeddings=True
    )
    print(f"✓ Generated {len(embeddings)} embeddings")
    print()

    # Save embeddings
    # Format: [num_queries:u32][dimension:u32][query_len:u16][query_bytes][embedding:f32*dim]...
    print(f"💾 Saving embeddings to {OUTPUT_FILE}...")
    os.makedirs(os.path.dirname(OUTPUT_FILE), exist_ok=True)
    
    with open(OUTPUT_FILE, 'wb') as f:
        # Magic header
        f.write(b'WTQE')  # WTF Query Embeddings
        # Version
        f.write((1).to_bytes(2, 'little'))
        # Num queries
        f.write(len(all_queries).to_bytes(4, 'little'))
        # Dimension
        f.write(dim.to_bytes(4, 'little'))
        
        for i, (query, emb) in enumerate(zip(all_queries, embeddings)):
            # Query text length
            query_bytes = query.encode('utf-8')
            f.write(len(query_bytes).to_bytes(2, 'little'))
            # Query text
            f.write(query_bytes)
            # Embedding
            for val in emb:
                f.write(val.tobytes())
    
    file_size = os.path.getsize(OUTPUT_FILE)
    print(f"✓ Saved {OUTPUT_FILE} ({file_size / 1024:.2f} KB)")
    print()

    # Verify
    print("🔍 Verifying embeddings...")
    with open(OUTPUT_FILE, 'rb') as f:
        magic = f.read(4)
        assert magic == b'WTQE', f"Invalid magic: {magic}"
        version = int.from_bytes(f.read(2), 'little')
        assert version == 1, f"Invalid version: {version}"
        num_q = int.from_bytes(f.read(4), 'little')
        dim_q = int.from_bytes(f.read(4), 'little')
        assert num_q == len(all_queries), f"Num mismatch: {num_q} != {len(all_queries)}"
        assert dim_q == dim, f"Dim mismatch: {dim_q} != {dim}"
    print("✓ Embeddings verified")
    print()

    print("=" * 60)
    print("✅ Done! Query embeddings ready")
    print("=" * 60)


if __name__ == "__main__":
    main()
