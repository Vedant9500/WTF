"""
embedding_server.py - Lightweight sentence-transformer embedding server for WTF CLI

This script provides a simple HTTP-like interface via stdin/stdout for generating
query embeddings using sentence-transformers.

Usage:
  python scripts/embedding_server.py
  
Protocol:
  Input: JSON line with query text
  Output: JSON line with embedding vector
  
Example:
  Input: {"query": "how to compress files"}
  Output: {"embedding": [0.1, 0.2, ...], "dimension": 384}
"""

import sys
import json
import os

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print(json.dumps({"error": "sentence-transformers not installed"}))
    sys.exit(1)

# Load model once
MODEL_NAME = "all-MiniLM-L6-v2"
model = None

def load_model():
    """Load sentence-transformer model."""
    global model
    if model is None:
        model = SentenceTransformer(MODEL_NAME)
    return model

def process_query(query: str):
    """Generate embedding for a query."""
    model = load_model()
    embedding = model.encode(
        [query],
        normalize_embeddings=True,
        convert_to_numpy=True
    )[0]
    return embedding.tolist()

def main():
    """Main server loop - read JSON from stdin, output JSON to stdout."""
    # Send ready signal
    print(json.dumps({"status": "ready", "model": MODEL_NAME}), flush=True)
    
    # Process queries
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        
        try:
            data = json.loads(line)
            query = data.get("query", "")
            
            if not query:
                print(json.dumps({"error": "no query provided"}), flush=True)
                continue
            
            embedding = process_query(query)
            print(json.dumps({
                "status": "ok",
                "embedding": embedding,
                "dimension": len(embedding)
            }), flush=True)
            
        except Exception as e:
            print(json.dumps({"error": str(e)}), flush=True)

if __name__ == "__main__":
    main()
