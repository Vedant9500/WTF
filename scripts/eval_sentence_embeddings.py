"""
eval_sentence_embeddings.py - Evaluate sentence-transformer search accuracy

This script tests sentence-transformer embeddings on the evaluation queries.

Usage:
  python scripts/eval_sentence_embeddings.py
"""

import os
import sys
import yaml
import math
from pathlib import Path

try:
    from sentence_transformers import SentenceTransformer, util
except ImportError:
    print("❌ sentence-transformers not installed!")
    print("   Install with: pip install sentence-transformers")
    sys.exit(1)

# Configuration
MODEL_NAME = "all-MiniLM-L6-v2"
COMMANDS_YAML = "assets/commands.yml"
EVAL_QUERIES_YAML = "assets/eval_queries.yaml"
SENTENCE_EMBEDDINGS = "assets/sentence_cmd_embeddings.bin"


def load_model():
    """Load sentence-transformer model."""
    print(f"📦 Loading model: {MODEL_NAME}...")
    model = SentenceTransformer(MODEL_NAME)
    print(f"✓ Model loaded (dim={model.get_sentence_embedding_dimension()})")
    return model


def load_commands():
    """Load commands from YAML."""
    print(f"📖 Loading commands...")
    with open(COMMANDS_YAML, 'r', encoding='utf-8') as f:
        data = yaml.safe_load(f)
    
    if isinstance(data, list):
        commands = data
    elif isinstance(data, dict):
        commands = data.get('commands', [])
    else:
        commands = []
    
    print(f"✓ Loaded {len(commands):,} commands")
    return commands


def load_eval_queries():
    """Load evaluation queries."""
    print(f"📋 Loading eval queries...")
    with open(EVAL_QUERIES_YAML, 'r', encoding='utf-8') as f:
        data = yaml.safe_load(f)
    
    queries = data.get('queries', [])
    print(f"✓ Loaded {len(queries)} queries")
    return queries


def prepare_command_text(command: dict) -> str:
    """Prepare command text for embedding (same as generation)."""
    parts = []
    
    if command.get('command'):
        parts.append(f"command: {command['command']}")
    if command.get('description'):
        parts.append(f"description: {command['description']}")
    if command.get('keywords'):
        parts.append(f"keywords: {', '.join(command['keywords'])}")
    if command.get('tags'):
        parts.append(f"tags: {', '.join(command['tags'])}")
    
    return " | ".join(parts)


def find_best_rank(results, relevant):
    """Find the rank of the best relevant command."""
    for i, result in enumerate(results):
        result_cmd = result['command'].lower()
        for rel in relevant:
            rel_lower = rel.lower()
            # Check various matching strategies
            if (rel_lower in result_cmd or 
                result_cmd in rel_lower or
                result_cmd.startswith(rel_lower) or
                rel_lower.startswith(result_cmd)):
                return i + 1
            # Check word overlap
            result_words = set(result_cmd.split())
            rel_words = set(rel_lower.split())
            if result_words & rel_words:  # Any word overlap
                if len(result_words & rel_words) >= len(rel_words) / 2:
                    return i + 1
    return math.inf


def evaluate():
    """Run evaluation."""
    print("=" * 60)
    print("Sentence-Transformer Evaluation")
    print("=" * 60)
    print()
    
    # Load model
    model = load_model()
    print()
    
    # Load commands
    commands = load_commands()
    print()
    
    # Prepare command texts
    command_texts = [prepare_command_text(cmd) for cmd in commands]
    
    # Generate command embeddings
    print("🧮 Generating command embeddings...")
    cmd_embeddings = model.encode(
        command_texts,
        batch_size=64,
        show_progress_bar=True,
        normalize_embeddings=True
    )
    print(f"✓ Generated {len(cmd_embeddings)} embeddings")
    print()
    
    # Load eval queries
    queries = load_eval_queries()
    print()
    
    # Evaluate each query
    print("🔍 Running evaluation...")
    metrics = {
        'total': 0,
        'hit_at_1': 0,
        'hit_at_3': 0,
        'hit_at_5': 0,
        'mrr_sum': 0.0,
    }
    
    for idx, q in enumerate(queries):
        query = q['query']
        relevant = q['relevant']
        
        # Generate query embedding
        query_embedding = model.encode(query, normalize_embeddings=True)
        
        # Compute similarities
        similarities = util.cos_sim(query_embedding, cmd_embeddings)[0]
        
        # Get top 5 results
        top_indices = similarities.argsort(descending=True)[:5]
        
        results = []
        for i in top_indices:
            results.append({
                'command': commands[i]['command'],
                'score': similarities[i].item()
            })
        
        # Find best rank
        rank = find_best_rank(results, relevant)
        
        # Update metrics
        metrics['total'] += 1
        if rank <= 1:
            metrics['hit_at_1'] += 1
        if rank <= 3:
            metrics['hit_at_3'] += 1
        if rank <= 5:
            metrics['hit_at_5'] += 1
        if rank != math.inf:
            metrics['mrr_sum'] += 1.0 / rank
        
        # Print first 5 queries for debugging
        if idx < 5:
            print(f"  Query: {query}")
            print(f"  Relevant: {relevant}")
            print(f"  Results:")
            for i, r in enumerate(results):
                print(f"    [{i+1}] {r['command']} (score={r['score']:.4f})")
            print(f"  Rank: {rank}")
            print()
        else:
            # Print summary
            status = "✓" if rank <= 5 else "✗"
            print(f"  {status} {query:50s} rank={rank}")
    
    # Print final metrics
    print()
    print("=" * 60)
    print("Evaluation Results")
    print("=" * 60)
    print()
    print(f"Total Queries:  {metrics['total']}")
    print(f"Hit@1:          {metrics['hit_at_1']}/{metrics['total']} ({metrics['hit_at_1']/metrics['total']*100:.1f}%)")
    print(f"Hit@3:          {metrics['hit_at_3']}/{metrics['total']} ({metrics['hit_at_3']/metrics['total']*100:.1f}%)")
    print(f"Hit@5:          {metrics['hit_at_5']}/{metrics['total']} ({metrics['hit_at_5']/metrics['total']*100:.1f}%)")
    print(f"MRR:            {metrics['mrr_sum']/metrics['total']:.3f}")
    print()


if __name__ == "__main__":
    evaluate()
