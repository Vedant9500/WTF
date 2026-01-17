"""
prepare_glove.py - Download GloVe 100d and convert to binary format for WTF CLI

This script:
1. Downloads GloVe 6B 100d vectors (if not present)
2. Filters to top 100k most common words
3. Saves in a compact binary format for fast Go loading

Output format (glove.bin):
  [vocab_size: uint32]
  For each word:
    [word_len: uint16][word: bytes][vector: 100 * float32]

Usage:
  python scripts/prepare_glove.py

Output:
  assets/glove.bin (~40MB)
"""

import os
import struct
import zipfile
import urllib.request
from pathlib import Path

# Configuration
GLOVE_URL = "https://nlp.stanford.edu/data/glove.6B.zip"
GLOVE_ZIP = "glove.6B.zip"
GLOVE_TXT = "glove.6B.100d.txt"
OUTPUT_FILE = "assets/glove.bin"
TOP_K_WORDS = 100000  # Keep top 100k most common words
VECTOR_DIM = 100


def download_glove():
    """Download GloVe vectors if not present."""
    if os.path.exists(GLOVE_TXT):
        print(f"✓ Found existing {GLOVE_TXT}")
        return
    
    if not os.path.exists(GLOVE_ZIP):
        print(f"⬇ Downloading GloVe from {GLOVE_URL}...")
        print("  (This is ~862MB, may take a few minutes)")
        urllib.request.urlretrieve(GLOVE_URL, GLOVE_ZIP, reporthook=download_progress)
        print()
    
    print(f"📦 Extracting {GLOVE_TXT}...")
    with zipfile.ZipFile(GLOVE_ZIP, 'r') as zf:
        zf.extract(GLOVE_TXT)
    print(f"✓ Extracted {GLOVE_TXT}")


def download_progress(block_num, block_size, total_size):
    """Progress callback for urllib."""
    downloaded = block_num * block_size
    percent = min(100, downloaded * 100 / total_size)
    print(f"\r  Progress: {percent:.1f}% ({downloaded // (1024*1024)}MB / {total_size // (1024*1024)}MB)", end="")


def load_glove_vectors(filepath: str, top_k: int) -> dict:
    """Load GloVe vectors from text file, keeping top_k words."""
    print(f"📖 Loading GloVe vectors (keeping top {top_k:,} words)...")
    
    vectors = {}
    with open(filepath, 'r', encoding='utf-8') as f:
        for i, line in enumerate(f):
            if i >= top_k:
                break
            
            parts = line.strip().split()
            word = parts[0]
            vector = [float(x) for x in parts[1:]]
            
            if len(vector) != VECTOR_DIM:
                continue  # Skip malformed lines
            
            vectors[word] = vector
            
            if (i + 1) % 10000 == 0:
                print(f"  Loaded {i + 1:,} words...")
    
    print(f"✓ Loaded {len(vectors):,} word vectors")
    return vectors


def save_binary_format(vectors: dict, output_path: str):
    """Save vectors in compact binary format."""
    print(f"💾 Saving binary format to {output_path}...")
    
    # Ensure output directory exists
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, 'wb') as f:
        # Write header: vocab size (uint32)
        f.write(struct.pack('<I', len(vectors)))
        
        # Write each word and vector
        for word, vector in vectors.items():
            # Word: length (uint16) + bytes
            word_bytes = word.encode('utf-8')
            f.write(struct.pack('<H', len(word_bytes)))
            f.write(word_bytes)
            
            # Vector: 100 float32 values
            f.write(struct.pack(f'<{VECTOR_DIM}f', *vector))
    
    file_size = os.path.getsize(output_path)
    print(f"✓ Saved {output_path} ({file_size / (1024*1024):.1f} MB)")


def verify_binary_file(filepath: str, sample_words: list):
    """Verify the binary file can be read correctly."""
    print(f"🔍 Verifying binary file...")
    
    with open(filepath, 'rb') as f:
        vocab_size = struct.unpack('<I', f.read(4))[0]
        print(f"  Vocab size: {vocab_size:,}")
        
        # Read first few words to verify
        for i in range(min(5, vocab_size)):
            word_len = struct.unpack('<H', f.read(2))[0]
            word = f.read(word_len).decode('utf-8')
            vector = struct.unpack(f'<{VECTOR_DIM}f', f.read(VECTOR_DIM * 4))
            print(f"  Sample word '{word}': vector[0:3] = {vector[0]:.4f}, {vector[1]:.4f}, {vector[2]:.4f}")
    
    print("✓ Binary file verified")


def main():
    print("=" * 60)
    print("GloVe Binary Converter for WTF CLI")
    print("=" * 60)
    print()
    
    # Step 1: Download GloVe
    download_glove()
    print()
    
    # Step 2: Load vectors
    vectors = load_glove_vectors(GLOVE_TXT, TOP_K_WORDS)
    print()
    
    # Step 3: Save binary format
    save_binary_format(vectors, OUTPUT_FILE)
    print()
    
    # Step 4: Verify
    verify_binary_file(OUTPUT_FILE, list(vectors.keys())[:5])
    print()
    
    print("=" * 60)
    print("✅ Done! GloVe binary file ready for WTF CLI")
    print(f"   Output: {OUTPUT_FILE}")
    print("=" * 60)


if __name__ == "__main__":
    main()
