# Contributing to v2rayA

Usage questions go to [Discussions](https://github.com/v2rayA/v2rayA/discussions). Issues are for a defect or a feature request, pull requests for a change you have built and run. The issue forms and the pull request template come in English and Chinese; pick either.

Before you open an issue or a pull request, check the [latest release](https://github.com/v2rayA/v2rayA/releases/latest): fixes go into the current release only. Then search the existing issues.

## The repository

| Where | What |
|---|---|
| `service/` | the Go service: API, settings database, transparent proxy, the generated core config |
| `core/` | `v2raya_core`, a fork of xray-core with the DNS module, the in-core TUN and the extra protocols |
| `gui/` | the web interface: Vue 3, Vuetify 4 (Material 3), vue-i18n, Vitest |
| `install/` | packaging: systemd and OpenRC units, AUR, Docker, the Windows installer |

`./build.sh`, from the repository root, builds the web interface and both binaries; it needs Go 1.26 and Node 24 with yarn. The service embeds `service/server/router/web`, which only `build.sh` fills: `yarn --cwd gui build` writes to `web/` at the root, and `go build` in `service/` fails until those files are copied across.

## Before a pull request

Run the checks CI runs, from the repository root:

```sh
(cd gui     && yarn lint && yarn typecheck && yarn i18n-check && yarn test && yarn build)
(cd service && go build ./... && go vet ./... && go test ./...)
(cd core    && go build ./... && go vet ./...)
```

Then exercise the change in the running program: a service change by starting `v2raya` and using the feature, a GUI change in a browser, at phone width and in both themes. `TestResolvHijackerConcurrentResetRemove` fails when the tests run as a non-root user; that one is known.

- One change per pull request, and each commit a coherent step of it: no mixed-up commits, no reformatting of untouched code, no lockfile churn.
- Visible text lives in `gui/src/locales/`: update all six locales together. `yarn i18n-check` compares key sets, not whether a translation followed a wording change. The in-app documentation under `gui/src/docs/` has the same six.
- Do not add a model `Co-Authored-By` trailer, an AI signature or a generated-by marker to a commit, pull request text or a comment.
- One topic per commit, imperative subject saying what changes, `fix:` / `feat:` / `docs:` / `ci:` as the history uses.

## AI-assisted contributions

An assistant may write and test the code, and you verify it before you submit it. Point it at [AGENTS.md](AGENTS.md), which states what this repository expects of it.

- Read its output before you push. If you cannot say why a line is there, it does not go in.
- Execute what it wrote. Do not present an untested fix as verified.
- Watch for the wrong root cause, argued well. Ask what else could produce the symptom.
- Keep the diff to the change; do not let the assistant tidy unrelated code or regenerate lockfiles.
- Between review rounds, be able to say in one sentence what changed and why.
- Say which part it wrote. That is not held against you; hiding it is.

Maintainers close an issue or a pull request submitted without human verification, or with a checklist checked without reading. Each checklist carries two statements that are false for an honest submitter; checking either one is the admission.
