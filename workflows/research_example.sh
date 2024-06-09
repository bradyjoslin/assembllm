#!/bin/bash

TOPIC="ten bullets summarizing extism plug-in systems with wasm"
RESEARCHER="you are a technical research assistant"
ANALYSIS="analyze these capabilities against the broader backdrop of webassembly."
WRITER="you are a technical writer specializing in trends, skilled at helping developers understand the practical use of new technology described from first principles"
BLOG_POST="write a blog post on the provided research, avoid bullets, use prose and include section headers"

assembllm -p perplexity "$TOPIC" \
| assembllm -r "$RESEARCHER" "$ANALYSIS" \
| assembllm --raw -r "$WRITER" "$BLOG_POST" \
| tee research_example_output.md