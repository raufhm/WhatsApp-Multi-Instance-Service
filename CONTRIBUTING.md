# Contributing

Thanks for wanting to help. This project is still evolving, so the most useful contributions are the ones that are small, clear, and grounded in the current codebase.

## Before You Start

- Read [README.md](README.md) for the project positioning and local setup.
- Read [AGENT.md](AGENT.md) for practical repo guidance.
- Check whether someone already opened an issue for the thing you want to work on.

## Local Setup

The quickest local path is Docker Compose:

```sh
cp .env.example .env
docker compose up --build -d
docker compose logs -f app
```

For frontend-only work:

```sh
cd frontend
npm install
npm run build
npm test -- --run
```

For backend work:

```sh
cd backend
go test ./...
go run .
```

## What Makes A Good Contribution

Good contributions tend to be:

- a bug fix with a clear failure mode
- a small usability improvement
- a test that protects existing behavior
- a documentation fix that makes setup or usage clearer
- a narrow refactor that reduces complexity without changing behavior

Please avoid large drive-by changes that mix unrelated work together.

## Code Style

- Match the style already used in the file you are editing.
- Keep diffs focused.
- Prefer small, reviewable changes over broad rewrites.
- Add tests when you change behavior or fix a bug.
- Be careful with wording in docs. If something is not fully verified, say that it "supports" or "appears to" instead of claiming more than the code proves.

## Testing

Before opening a pull request, please run the most relevant checks you can:

- backend changes: `go test ./...`
- frontend changes: `npm test -- --run`
- build-related changes: `npm run build`

If you only touched a small area, targeted tests are still valuable.

## Pull Request Checklist

- The change is scoped and easy to review.
- The code works locally.
- Relevant tests pass.
- Any new behavior is documented.
- Screenshots or short recordings are included for visible UI changes.
- The PR description explains the problem and the fix clearly.

## Reporting Bugs

When opening a bug report, include:

- what you expected to happen
- what actually happened
- steps to reproduce
- your OS, browser, or runtime if relevant
- any error messages or logs that help explain the issue

## Requesting Features

When opening a feature request, try to include:

- the problem you are trying to solve
- the workflow you want to improve
- whether this is blocking you today
- any rough ideas for implementation, if you have them

## Good First Issues

If you are new to the project, look for:

- documentation cleanup
- small UI polish
- test coverage gaps
- error message improvements
- minor bug fixes with a clear reproduction path

## Need Help?

If something in the setup or workflow is unclear, open an issue and describe where you got stuck. That usually helps improve the repo for everyone.
