# LLM Review

## Overview

This tool is designed to evaluate Large Language Models (LLMs) across various scenarios, such as:

- Underrepresented languages
- Diverse linguistic and knowledge contexts

## How to use it

Define a list of tasks with the following format:

```yaml
tasks:
  - name: No prompt
    models:
      - openai
    prompt: 
    data_path: ./data/
    result_path: ./result/
```

Data files will be recursively parsed, and results will be stored in the result path following the same directory structure.
