## Description

Briefly describe the motivation behind this change and the problem it resolves.

Fixes #(issue number)

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Safety / Security enhancement
- [ ] Documentation update
- [ ] CI/CD / Infrastructure update

## Architectural & Safety Matrix Check

- [ ] Does this change modify API mutation logic?
- [ ] Has client-side safety middleware (`readonly`, `readwrite-mine`, etc.) been preserved and tested?
- [ ] Have credential redaction rules been verified for any output changes?

## How Has This Been Tested?

Please describe the tests that you ran to verify your changes:

- [ ] Unit tests (`go test -v ./...`)
- [ ] Integration tests against Zabbix 7.0+ API
- [ ] Manual CLI verification

## Checklist

- [ ] My code follows the style guidelines of this project (`gofmt -s -w .`).
- [ ] I have performed a self-review of my own code.
- [ ] I have added tests that prove my fix is effective or that my feature works.
- [ ] New and existing unit tests pass locally with my changes.
