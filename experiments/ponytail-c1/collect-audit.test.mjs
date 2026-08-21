import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { artifactFingerprint, collect, resolveDefaultMode } from "./collect-audit.mjs";

function write(root, relative, content) {
  const absolute = path.join(root, relative);
  fs.mkdirSync(path.dirname(absolute), { recursive: true });
  fs.writeFileSync(absolute, content);
}

function fakePlugin(root) {
  write(root, ".codex-plugin/plugin.json", JSON.stringify({
    name: "ponytail",
    version: "test-1",
    skills: "./instruction-set/",
    hooks: "./lifecycle/map.json",
  }));
  write(root, "package.json", JSON.stringify({ name: "ponytail", version: "test-1" }));
  write(root, "instruction-set/ponytail/SKILL.md", "---\nname: ponytail\n---\nTEST RULES\n");
  write(root, "lifecycle/map.json", JSON.stringify({
    hooks: {
      SessionStart: [{
        matcher: "startup|resume|clear|compact",
        hooks: [{
          type: "command",
          command: "node \"${CLAUDE_PLUGIN_ROOT}/lifecycle/start.js\"",
          timeout: 5,
        }],
      }],
      SubagentStart: [{
        hooks: [{
          type: "command",
          command: "node \"${CLAUDE_PLUGIN_ROOT}/lifecycle/subagent.js\"",
          timeout: 5,
        }],
      }],
      UserPromptSubmit: [{
        hooks: [{
          type: "command",
          command: "node \"${CLAUDE_PLUGIN_ROOT}/lifecycle/prompt.js\"",
          timeout: 5,
        }],
      }],
    },
  }));
  write(root, "lifecycle/start.js", `
const fs = require("node:fs");
const path = require("node:path");
const state = path.join(process.env.PLUGIN_DATA, ".ponytail-active");
fs.mkdirSync(path.dirname(state), { recursive: true });
fs.writeFileSync(state, "full");
process.stdout.write(JSON.stringify({
  systemMessage: "PONYTAIL:FULL",
  hookSpecificOutput: { hookEventName: "SessionStart", additionalContext: "TEST RULES" },
}));
`);
  write(root, "lifecycle/subagent.js", `
const fs = require("node:fs");
const path = require("node:path");
let mode;
try { mode = fs.readFileSync(path.join(process.env.PLUGIN_DATA, ".ponytail-active"), "utf8").trim(); }
catch { process.exit(0); }
function emit() {
  process.stdout.write(JSON.stringify({
    systemMessage: "PONYTAIL:" + mode.toUpperCase(),
    hookSpecificOutput: { hookEventName: "SubagentStart", additionalContext: "TEST RULES" },
  }));
}
if (!process.env.PONYTAIL_SUBAGENT_MATCHER) {
  emit();
} else {
  let done = false;
  function finish() { if (!done) { done = true; emit(); } }
  process.stdin.on("data", () => {});
  process.stdin.on("end", finish);
  setTimeout(() => { finish(); process.exit(0); }, 1000).unref();
}
`);
  write(root, "lifecycle/prompt.js", `
const fs = require("node:fs");
const path = require("node:path");
const state = path.join(process.env.PLUGIN_DATA, ".ponytail-active");
let input = "";
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  const prompt = JSON.parse(input).prompt;
  const mode = prompt.includes("ultra") ? "ultra" : "off";
  if (mode === "off") { try { fs.unlinkSync(state); } catch {} }
  else fs.writeFileSync(state, mode);
  process.stdout.write(JSON.stringify({
    systemMessage: "PONYTAIL:" + mode.toUpperCase(),
    hookSpecificOutput: { hookEventName: "UserPromptSubmit", additionalContext: "MODE " + mode },
  }));
});
`);
}

test("collector keeps local, isolated, and live evidence boundaries separate", async () => {
  const parent = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-test-"));
  const pluginRoot = path.join(parent, "installed", "ponytail");
  const stateDir = path.join(parent, "live-plugin-data");
  const configPath = path.join(parent, "invalid-config.json");
  try {
    const git = spawnSync("git", ["init", "--quiet", parent]);
    assert.equal(git.status, 0);
    fakePlugin(pluginRoot);
    fs.mkdirSync(stateDir, { recursive: true });
    fs.writeFileSync(path.join(stateDir, ".ponytail-active"), "review");
    fs.writeFileSync(configPath, "{not-json\n");

    const staticReport = await collect({
      subject: "local-installation",
      pluginRoot,
      stateDir,
      config: configPath,
      instructions: [],
    }, {});
    assert.equal(staticReport.isolated_hook_probes.status, "not-run");

    const mismatched = await collect({
      subject: "local-installation",
      pluginRoot,
      stateDir,
      config: configPath,
      instructions: [],
      probeIsolated: true,
      expectArtifactSha256: "0".repeat(64),
    }, {});
    assert.equal(mismatched.isolated_hook_probes.status, "probe-ineligible");
    assert.equal(fs.readFileSync(path.join(stateDir, ".ponytail-active"), "utf8"), "review");

    const report = await collect({
      subject: "local-installation",
      pluginRoot,
      stateDir,
      config: configPath,
      instructions: [],
      probeIsolated: true,
      expectArtifactSha256: staticReport.installation.artifact.sha256,
    }, {});

    assert.equal(report.installation.git.status, "not-git-root");
    assert.equal(report.installation.git.commit, null);
    assert.equal(report.installation.codex_manifest.skills_ref, "./instruction-set/");
    assert.equal(report.installation.codex_manifest.hooks_ref, "./lifecycle/map.json");
    assert.equal(report.installation.hook_map.path, "lifecycle/map.json");
    assert.equal(report.installation.skills.length, 1);
    assert.equal(report.mode_sources.config.status, "invalid");
    assert.equal(report.mode_sources.active_state.status, "valid");
    assert.equal(report.mode_sources.active_state.value, "review");
    assert.equal(report.isolated_hook_probes.status, "completed");
    assert.equal(report.isolated_hook_probes.observations_valid, true);
    assert.equal(report.isolated_hook_probes.probes.length, 7);
    assert.equal(report.isolated_hook_probes.observations.session_and_subagent_instruction_match, true);
    assert.equal(report.isolated_hook_probes.observations.open_stdin_completed_without_eof, true);
    assert.equal(report.isolated_hook_probes.observations.mode_persisted_across_compact, false);
    assert.equal(report.isolated_hook_probes.observations.disable_removed_state, true);
    assert.equal(report.evidence_boundary.current_codex_app_activation, "unknown");
    assert.equal(report.evidence_boundary.live_subagent_propagation, "unknown");
    assert.equal(Object.hasOwn(report, "collector_checks"), false);
  } finally {
    fs.rmSync(parent, { recursive: true, force: true });
  }
});

test("default resolution matches Ponytail's non-trimming precedence", () => {
  assert.deepEqual(resolveDefaultMode("ultra", { defaultMode: "lite" }), {
    value: "ultra",
    source: "PONYTAIL_DEFAULT_MODE",
  });
  assert.deepEqual(resolveDefaultMode(" full ", { defaultMode: "lite" }), {
    value: "lite",
    source: "config.defaultMode",
  });
  assert.deepEqual(resolveDefaultMode(undefined, { defaultMode: " lite " }), {
    value: "full",
    source: "built-in default",
  });
});

test("static collection reports manifest references outside the plugin root", async () => {
  const parent = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-escape-test-"));
  const pluginRoot = path.join(parent, "ponytail");
  try {
    write(pluginRoot, ".codex-plugin/plugin.json", JSON.stringify({
      name: "ponytail",
      version: "test-escape",
      skills: "./skills/",
      hooks: "./../outside-hooks.json",
    }));
    write(pluginRoot, "skills/ponytail/SKILL.md", "TEST\n");
    write(parent, "outside-hooks.json", JSON.stringify({ hooks: {} }));

    const report = await collect({
      subject: "local-installation",
      pluginRoot,
      instructions: [],
    }, {});
    assert.equal(report.installation.inspection_status, "probe-ineligible");
    assert.equal(report.installation.inspection_error_code, "reference-outside-plugin-root");
    assert.equal(report.isolated_hook_probes.status, "not-run");
  } finally {
    fs.rmSync(parent, { recursive: true, force: true });
  }
});

test("artifact scope excludes only root git metadata and marks symlinks incomplete", (t) => {
  if (process.platform === "win32") {
    t.skip("symlink creation is privilege-dependent on Windows");
    return;
  }
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-artifact-test-"));
  try {
    write(root, ".git/HEAD", "root-one\n");
    write(root, "vendor/.git/config", "nested-one\n");
    write(root, "rules.txt", "rules\n");
    fs.symlinkSync("rules.txt", path.join(root, "rules-link"));

    const first = artifactFingerprint(root);
    assert.equal(first.complete, false);
    assert.deepEqual(first.entries.map((entry) => entry.path), [
      "rules-link",
      "rules.txt",
      "vendor/.git/config",
    ]);

    write(root, ".git/HEAD", "root-two\n");
    assert.equal(artifactFingerprint(root).sha256, first.sha256);
    write(root, "vendor/.git/config", "nested-two\n");
    assert.notEqual(artifactFingerprint(root).sha256, first.sha256);
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
});

test("static collection reports a manifest symlink outside the root", async (t) => {
  if (process.platform === "win32") {
    t.skip("symlink creation is privilege-dependent on Windows");
    return;
  }
  const parent = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-symlink-test-"));
  const pluginRoot = path.join(parent, "ponytail");
  try {
    write(pluginRoot, ".codex-plugin/plugin.json", JSON.stringify({
      name: "ponytail",
      version: "test-symlink",
      skills: "./skills/",
      hooks: "./hooks-link.json",
    }));
    write(pluginRoot, "skills/ponytail/SKILL.md", "TEST\n");
    write(parent, "outside-hooks.json", JSON.stringify({ hooks: {} }));
    fs.symlinkSync(path.join(parent, "outside-hooks.json"), path.join(pluginRoot, "hooks-link.json"));

    const report = await collect({
      subject: "local-installation",
      pluginRoot,
      instructions: [],
    }, {});
    assert.equal(report.installation.inspection_status, "probe-ineligible");
    assert.equal(report.installation.inspection_error_code, "reference-outside-plugin-root");
    assert.equal(report.installation.artifact.complete, false);
  } finally {
    fs.rmSync(parent, { recursive: true, force: true });
  }
});
