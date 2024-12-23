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


Data files are processed recursively, and the results are saved in the designated result path while maintaining the same directory structure.

If you provide correction files, they will be utilized to evaluate the output's performance. Corrections should be separate files,
appended with the "_correction" suffix in the filename. Additionally, a file with the "_grade" suffix will be generated as part of the output.

File structure example based on the previous task definition:

```
./data/
  |-- catalan_1.pdf
  |-- catalan_1_correction.pdf
  `-- catalan_2.txt                # this file doesn't have a correction
./result/
  |-- catalan_1_openai.txt         # output of the openai LLM
  |-- catalan_1_openai_grade.txt   # grade of the previous result
  `-- catalan_2_openai.txt         # this file won't have a correction, as no correction was provided
```

## Roadmap

- [ ] Support various models for assessment and grading.
- [ ] Refactor code to simplify the architecture.
