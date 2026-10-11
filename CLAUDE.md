# Project Guardrails

## Models

- Use `opusplan`: Claude Opus 5.5 for planning, Claude Sonnet 5.5 for writing code.

## Planning first

- Plan before touching anything. No file edits, installs, commands with side effects, or commits until the plan is approved.
- Ask any relevant clarifying questions before planning, and ask them as a batch, not one at a time.
- Make the plan specific: list the exact files to change, the steps in order, the commands to run, and how the result will be verified.

## Git and commits

- Never list Claude, Claude Code, or Anthropic as a co-author, co-worker, or contributor in any commit.
- Do not add `Co-Authored-By` trailers that name Claude or Anthropic.
- Do not add "Generated with Claude Code" or similar footers to commit messages, PR titles, or PR descriptions.
- This applies in git generally: commit messages, branch names, tags, notes, and PR text.
- Commits are authored under the user's own git identity only.

## Branch naming

Use the prefix that matches the type of work:

- Feature: `feat/[feature-name]`
- Fix: `fix/[fix-name]`
- Deployment: `deploy/[name]`
- Plan: `plan/[name]`
- Documentation: `docs/[name]`

Use lowercase kebab-case for the name part, for example `feat/user-login-otp`.

## Agents and skills

- For frontend work, use the `design-taste-frontend` skill.
- For coding work delegated to a sub-agent, use the `ponytail:ponytail` agent.
- For technical conversations, use the `caveman` skill for compressed, no-padding responses.
