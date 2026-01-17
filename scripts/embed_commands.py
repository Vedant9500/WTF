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
import yaml
from pathlib import Path

# Configuration
GLOVE_BIN = "assets/glove.bin"
COMMANDS_YAML = "assets/commands.yml"
OUTPUT_FILE = "assets/cmd_embeddings.bin"
VECTOR_DIM = 100


def load_glove_binary(filepath: str) -> dict:
    """Load word vectors from binary file."""
    print(f"📖 Loading word vectors from {filepath}...")
    
    vectors = {}
    with open(filepath, 'rb') as f:
        vocab_size = struct.unpack('<I', f.read(4))[0]
        print(f"  Vocab size: {vocab_size:,}")
        
        for i in range(vocab_size):
            word_len = struct.unpack('<H', f.read(2))[0]
            word = f.read(word_len).decode('utf-8')
            vector = list(struct.unpack(f'<{VECTOR_DIM}f', f.read(VECTOR_DIM * 4)))
            vectors[word] = vector
            
            if (i + 1) % 20000 == 0:
                print(f"  Loaded {i + 1:,} / {vocab_size:,} words...")
    
    print(f"✓ Loaded {len(vectors):,} word vectors")
    return vectors


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


def compute_embedding(command: dict, word_vectors: dict) -> list:
    """Compute embedding for a command by averaging word vectors."""
    # Combine text from command, description, and keywords
    texts = []
    
    if command.get('command'):
        texts.append(command['command'])
    
    if command.get('description'):
        texts.append(command['description'])
    
    if command.get('keywords'):
        texts.extend(command['keywords'])
    
    # Tokenize all text
    all_tokens = []
    for text in texts:
        all_tokens.extend(tokenize(str(text)))
    
    # Get vectors for tokens that exist in vocabulary
    vectors = []
    for token in all_tokens:
        if token in word_vectors:
            vectors.append(word_vectors[token])
    
    # Average the vectors
    if vectors:
        embedding = [0.0] * VECTOR_DIM
        for vec in vectors:
            for i in range(VECTOR_DIM):
                embedding[i] += vec[i]
        for i in range(VECTOR_DIM):
            embedding[i] /= len(vectors)
        return embedding
    else:
        # No matching words - return zero vector
        return [0.0] * VECTOR_DIM


def save_embeddings(embeddings: list, output_path: str):
    """Save command embeddings in binary format."""
    print(f"💾 Saving embeddings to {output_path}...")
    
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, 'wb') as f:
        # Header: num_commands, dimension
        f.write(struct.pack('<I', len(embeddings)))
        f.write(struct.pack('<I', VECTOR_DIM))
        
        # Write each embedding
        for embedding in embeddings:
            f.write(struct.pack(f'<{VECTOR_DIM}f', *embedding))
    
    file_size = os.path.getsize(output_path)
    print(f"✓ Saved {output_path} ({file_size / (1024*1024):.2f} MB)")


def verify_embeddings(filepath: str, num_samples: int = 5):
    """Verify the binary file can be read correctly."""
    print(f"🔍 Verifying embeddings file...")
    
    with open(filepath, 'rb') as f:
        num_commands = struct.unpack('<I', f.read(4))[0]
        dimension = struct.unpack('<I', f.read(4))[0]
        print(f"  Commands: {num_commands:,}, Dimension: {dimension}")
        
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
    word_vectors = load_glove_binary(GLOVE_BIN)
    print()
    
    # Step 2: Load commands
    commands = load_commands(COMMANDS_YAML)
    print()
    
    # Step 3: Compute embeddings
    print("🧮 Computing command embeddings...")
    embeddings = []
    zero_count = 0
    
    for i, cmd in enumerate(commands):
        embedding = compute_embedding(cmd, word_vectors)
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
    save_embeddings(embeddings, OUTPUT_FILE)
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
