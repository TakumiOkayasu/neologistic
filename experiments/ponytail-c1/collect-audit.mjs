#!/usr/bin/env node

import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import process from "node:process";
import { fileURLToPath, pathToFileURL } from "node:url";

const DEFAULT_MODES = new Set(["off", "lite", "full", "ultra"]);
const ACTIVE_MODES = new Set([...DEFAULT_MODES, "review"]);
const EXPECTED_EVENTS = ["SessionStart", "SubagentStart", "UserPromptSubmit"];
const MAX_FILES = 5_000;
const MAX_FILE_BYTES = 32 * 1024 * 1024;
const MAX_TOTAL_BYTES = 256 * 1024 * 1024;
const MAX_HOOK_OUTPUT_BYTES = 1024 * 1024;
const SCRIPT_PATH = fileURLToPath(import.meta.url);

function sha256(data) {
  return createHash("sha256").update(data).digest("hex");
}

function logicalPath(root, absolute) {
  return path.relative(root, absolute).split(path.sep).join("/");
}

function isInside(root, candidate) {
  const relative = path.relative(root, candidate);
  return relative === "" || (!relative.startsWith(`..${path.sep}`) && relative !== ".." && !path.isAbsolute(relative));
}

function walkArtifact(root, current = root, result = []) {
  for (const item of fs.readdirSync(current, { withFileTypes: true })) {
    if (current === root && item.name === ".git") continue;
    const absolute = path.join(current, item.name);
    if (item.isDirectory()) {
      walkArtifact(root, absolute, result);
    } else if (item.isFile() || item.isSymbolicLink()) {
      result.push(absolute);
      if (result.length > MAX_FILES) throw new Error(`artifact exceeds ${MAX_FILES} entries`);
    } else {
      throw new Error(`unsupported artifact entry type: ${logicalPath(root, absolute)}`);
    }
  }
  return result;
}

function fileEvidence(root, absolute) {
  const stat = fs.lstatSync(absolute);
  const type = stat.isSymbolicLink() ? "symlink" : "file";
  if (type === "file" && stat.size > MAX_FILE_BYTES) {
    throw new Error(`artifact file exceeds ${MAX_FILE_BYTES} bytes: ${logicalPath(root, absolute)}`);
  }
  const content = type === "symlink"
    ? Buffer.from(fs.readlinkSync(absolute), "utf8")
    : fs.readFileSync(absolute);
  return {
    path: logicalPath(root, absolute),
    type,
    bytes: content.byteLength,
    sha256: sha256(content),
  };
}

function artifactFingerprint(root) {
  const entries = [];
  let totalBytes = 0;
  for (const absolute of walkArtifact(root)) {
    const entry = fileEvidence(root, absolute);
    totalBytes += entry.bytes;
    if (totalBytes > MAX_TOTAL_BYTES) throw new Error(`artifact exceeds ${MAX_TOTAL_BYTES} bytes`);
    entries.push(entry);
  }
  entries.sort((a, b) => Buffer.compare(Buffer.from(a.path, "utf8"), Buffer.from(b.path, "utf8")));
  const canonical = entries.map((entry) => JSON.stringify(entry)).join("\n") + "\n";
  return {
    algorithm: "sha256(canonical-jsonl(path,type,bytes,sha256) in UTF-8 path-byte order; excludes root .git only)",
    sha256: sha256(Buffer.from(canonical, "utf8")),
    complete: entries.every((entry) => entry.type === "file"),
    file_count: entries.length,
    total_bytes: totalBytes,
    entries,
  };
}

function nextValue(argv, index, flag) {
  if (index + 1 >= argv.length) throw new Error(`${flag} requires a value`);
  return argv[index + 1];
}

function parseArgs(argv) {
  const result = { instructions: [] };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--help") result.help = true;
    else if (arg === "--subject") result.subject = nextValue(argv, index++, arg);
    else if (arg === "--plugin-root") result.pluginRoot = nextValue(argv, index++, arg);
    else if (arg === "--state-dir") result.stateDir = nextValue(argv, index++, arg);
    else if (arg === "--config") result.config = nextValue(argv, index++, arg);
    else if (arg === "--instruction") result.instructions.push(nextValue(argv, index++, arg));
    else if (arg === "--probe-isolated") result.probeIsolated = true;
    else if (arg === "--expect-artifact-sha256") result.expectArtifactSha256 = nextValue(argv, index++, arg);
    else if (arg === "--out") result.out = nextValue(argv, index++, arg);
    else throw new Error(`unknown argument: ${arg}`);
  }
  return result;
}

function usage() {
  return `Usage:
  node collect-audit.mjs --subject local-installation --plugin-root <dir> [options]

Options:
  --subject <kind>          local-installation or upstream-reference (required)
  --plugin-root <dir>       Ponytail artifact root (required)
  --state-dir <dir>         Live Codex PLUGIN_DATA directory, when known
  --config <file>           Effective Ponytail config path; otherwise inferred
  --instruction <label=path>  Instruction source to hash; repeatable
  --probe-isolated          Run manifest hooks from a verified temporary snapshot
  --expect-artifact-sha256 <hex>  Required hash gate for --probe-isolated
  --out <file>              Write JSON without replacing an existing file
  --help                    Show this help
`;
}

function contentEvidence(content) {
  return { bytes: content.byteLength, sha256: sha256(content) };
}

function readJsonFile(absolute) {
  let content;
  try {
    const stat = fs.statSync(absolute);
    if (!stat.isFile() || stat.size > MAX_FILE_BYTES) {
      return { report: { status: "invalid", bytes: stat.size }, value: null };
    }
    content = fs.readFileSync(absolute);
  } catch (error) {
    return {
      report: { status: error.code === "ENOENT" ? "missing" : "unreadable" },
      value: null,
    };
  }
  const evidence = contentEvidence(content);
  try {
    const value = JSON.parse(content.toString("utf8").replace(/^\uFEFF/, ""));
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      return { report: { status: "invalid", ...evidence }, value: null };
    }
    return { report: { status: "valid", ...evidence }, value };
  } catch {
    return { report: { status: "invalid", ...evidence }, value: null };
  }
}

function readModeFile(absolute) {
  let content;
  try {
    const stat = fs.statSync(absolute);
    if (!stat.isFile() || stat.size > 4_096) {
      return { status: "invalid", value: null, bytes: stat.size };
    }
    content = fs.readFileSync(absolute);
  } catch (error) {
    return { status: error.code === "ENOENT" ? "missing" : "unreadable", value: null };
  }
  const evidence = contentEvidence(content);
  const value = content.toString("utf8").trim();
  if (!ACTIVE_MODES.has(value)) return { status: "invalid", value: null, ...evidence };
  return { status: "valid", value, ...evidence };
}

function exactDefaultMode(value) {
  if (typeof value !== "string" || value.length === 0) return null;
  const normalized = value.toLowerCase();
  return DEFAULT_MODES.has(normalized) ? normalized : null;
}

function resolveDefaultMode(envMode, config) {
  const fromEnvironment = exactDefaultMode(envMode);
  if (fromEnvironment) return { value: fromEnvironment, source: "PONYTAIL_DEFAULT_MODE" };
  const fromConfig = exactDefaultMode(config?.defaultMode);
  if (fromConfig) return { value: fromConfig, source: "config.defaultMode" };
  return { value: "full", source: "built-in default" };
}

function hashedValue(value, validSet = null) {
  if (value === undefined) return { status: "missing" };
  const content = Buffer.from(String(value), "utf8");
  if (validSet === null) return { status: "present", ...contentEvidence(content) };
  const normalized = typeof value === "string" ? value.toLowerCase() : null;
  if (validSet?.has(normalized)) return { status: "valid", value: normalized, ...contentEvidence(content) };
  return { status: "invalid", ...contentEvidence(content) };
}

function matcherEvidence(value) {
  const evidence = hashedValue(value);
  if (value === undefined) return evidence;
  try {
    new RegExp(value, "i");
    return { ...evidence, regex_valid: true };
  } catch {
    return { ...evidence, regex_valid: false };
  }
}

function configLocation(env, homedir) {
  if (env.XDG_CONFIG_HOME) return path.join(env.XDG_CONFIG_HOME, "ponytail", "config.json");
  if (process.platform === "win32") {
    return path.join(env.APPDATA || path.join(homedir, "AppData", "Roaming"), "ponytail", "config.json");
  }
  return path.join(homedir, ".config", "ponytail", "config.json");
}

function redactPath(absolute, roots) {
  const resolved = path.resolve(absolute);
  for (const [label, root] of roots) {
    if (!root) continue;
    const candidate = path.resolve(root);
    if (resolved === candidate) return label;
    if (isInside(candidate, resolved)) return `${label}/${logicalPath(candidate, resolved)}`;
  }
  return `<EXTERNAL>/${path.basename(resolved)}`;
}

function runGit(root, args) {
  const result = spawnSync("git", ["-C", root, ...args], {
    encoding: "utf8",
    maxBuffer: 64 * 1024,
  });
  if (result.error) return { status: "unavailable", value: null };
  if (result.status !== 0) return { status: "not-git-worktree", value: null };
  return { status: "known", value: result.stdout.trim() };
}

function gitEvidence(root) {
  const top = runGit(root, ["rev-parse", "--show-toplevel"]);
  if (top.status !== "known") return { status: top.status, commit: null };
  let exactRoot = false;
  try {
    exactRoot = fs.realpathSync(top.value) === root;
  } catch {
    return { status: "unreadable-git-root", commit: null };
  }
  if (!exactRoot) return { status: "not-git-root", commit: null };
  const commit = runGit(root, ["rev-parse", "HEAD"]);
  return commit.status === "known"
    ? { status: "known", commit: commit.value }
    : { status: commit.status, commit: null };
}

function resolveInside(root, reference, expectedType) {
  if (typeof reference !== "string" || !reference.startsWith("./")) {
    throw new Error(`manifest reference must start with ./: ${String(reference)}`);
  }
  const lexical = path.resolve(root, reference);
  if (!isInside(root, lexical)) throw new Error(`manifest reference escapes plugin root: ${reference}`);
  let resolved;
  try {
    resolved = fs.realpathSync(lexical);
  } catch {
    throw new Error(`manifest reference is missing: ${reference}`);
  }
  if (!isInside(root, resolved)) throw new Error(`manifest reference resolves outside plugin root: ${reference}`);
  const stat = fs.statSync(resolved);
  if (expectedType === "file" && !stat.isFile()) throw new Error(`manifest reference is not a file: ${reference}`);
  if (expectedType === "directory" && !stat.isDirectory()) throw new Error(`manifest reference is not a directory: ${reference}`);
  return resolved;
}

function parseNodeCommand(root, command) {
  if (typeof command !== "string") return null;
  const match = command.match(/^node\s+(?:"\$\{CLAUDE_PLUGIN_ROOT\}\/([^"\r\n]+\.js)"|'\$\{CLAUDE_PLUGIN_ROOT\}\/([^'\r\n]+\.js)'|\$\{CLAUDE_PLUGIN_ROOT\}\/(\S+\.js))$/);
  const reference = match?.[1] || match?.[2] || match?.[3];
  if (!reference) return null;
  try {
    return {
      reference: `./${reference.replaceAll("\\", "/")}`,
      absolute: resolveInside(root, `./${reference}`, "file"),
    };
  } catch {
    return null;
  }
}

function inspectInstallation(root, artifact) {
  const manifestPath = resolveInside(root, "./.codex-plugin/plugin.json", "file");
  const manifest = readJsonFile(manifestPath);
  if (manifest.report.status !== "valid") throw new Error(`Codex manifest is ${manifest.report.status}`);
  if (manifest.value.name !== "ponytail" || typeof manifest.value.version !== "string") {
    throw new Error("Codex manifest must identify a versioned ponytail plugin");
  }
  const skillsRoot = resolveInside(root, manifest.value.skills, "directory");
  const hookMapPath = resolveInside(root, manifest.value.hooks, "file");
  const hookMap = readJsonFile(hookMapPath);
  if (hookMap.report.status !== "valid" || !hookMap.value.hooks || typeof hookMap.value.hooks !== "object") {
    throw new Error("Codex hook map is invalid");
  }

  const declarations = [];
  const runnable = new Map();
  const eventCounts = new Map(EXPECTED_EVENTS.map((event) => [event, 0]));
  for (const [event, groups] of Object.entries(hookMap.value.hooks)) {
    for (const group of Array.isArray(groups) ? groups : []) {
      for (const hook of Array.isArray(group?.hooks) ? group.hooks : []) {
        if (eventCounts.has(event)) eventCounts.set(event, eventCounts.get(event) + 1);
        const command = String(hook?.command ?? "");
        const parsedCommand = parseNodeCommand(root, command);
        const timeout = Number(hook?.timeout);
        declarations.push({
          event,
          matcher: group?.matcher === undefined ? { status: "missing" } : hashedValue(group.matcher),
          type: hook?.type ?? null,
          command: { ...contentEvidence(Buffer.from(command, "utf8")), shape_valid: parsedCommand !== null },
          script: parsedCommand ? {
            path: parsedCommand.reference,
            ...fileEvidence(root, parsedCommand.absolute),
          } : null,
          timeout_seconds: Number.isFinite(timeout) ? timeout : null,
          status_message: hook?.statusMessage === undefined ? { status: "missing" } : hashedValue(hook.statusMessage),
        });
        if (EXPECTED_EVENTS.includes(event) && hook?.type === "command" && parsedCommand) {
          if (runnable.has(event)) throw new Error(`multiple runnable ${event} hooks`);
          if (!Number.isFinite(timeout) || timeout <= 0 || timeout > 30) {
            throw new Error(`${event} timeout must be between 0 and 30 seconds`);
          }
          runnable.set(event, { ...parsedCommand, timeout_ms: Math.round(timeout * 1_000) });
        }
      }
    }
  }
  for (const event of EXPECTED_EVENTS) {
    if (eventCounts.get(event) !== 1) throw new Error(`expected exactly one ${event} hook`);
    if (!runnable.has(event)) throw new Error(`missing safe runnable ${event} hook`);
  }

  const skillsPrefix = logicalPath(root, skillsRoot);
  const skills = artifact.entries.filter((entry) => {
    const insideSkills = entry.path === "SKILL.md" || entry.path.startsWith(`${skillsPrefix}/`);
    return insideSkills && entry.path.endsWith("/SKILL.md");
  });
  const packagePath = path.join(root, "package.json");
  const packageJson = readJsonFile(packagePath);
  const packageVersion = packageJson.value && typeof packageJson.value.version === "string"
    ? packageJson.value.version
    : null;

  return {
    report: {
      inspection_status: "probe-eligible",
      version: manifest.value.version,
      package_version: packageVersion,
      versions_match: packageVersion === null ? null : packageVersion === manifest.value.version,
      codex_manifest: {
        path: ".codex-plugin/plugin.json",
        ...manifest.report,
        name: manifest.value.name,
        version: manifest.value.version,
        skills_ref: manifest.value.skills,
        hooks_ref: manifest.value.hooks,
      },
      package_json: { path: "package.json", ...packageJson.report, version: packageVersion },
      skills,
      hook_map: { path: logicalPath(root, hookMapPath), ...hookMap.report },
      hook_declarations: declarations,
    },
    runnable,
  };
}

function fieldShapeEvidence(value) {
  if (value === undefined) return { status: "missing" };
  const encoded = Buffer.from(JSON.stringify(value), "utf8");
  return {
    status: "present",
    type: Array.isArray(value) ? "array" : value === null ? "null" : typeof value,
    item_count: Array.isArray(value) ? value.length : null,
    ...contentEvidence(encoded),
  };
}

function inspectionErrorCode(error) {
  const message = String(error?.message || "");
  if (/escapes plugin root|outside plugin root/.test(message)) return "reference-outside-plugin-root";
  if (/missing/.test(message)) return "required-entry-missing";
  if (/multiple|exactly one/.test(message)) return "hook-count-drift";
  if (/timeout/.test(message)) return "unsafe-timeout";
  return "unsupported-manifest-or-hook-wiring";
}

function staticManifestSummary(root, error) {
  const manifestPath = path.join(root, ".codex-plugin", "plugin.json");
  let manifest;
  try {
    if (!fs.lstatSync(manifestPath).isFile()) throw new Error("not a regular file");
    manifest = readJsonFile(manifestPath);
  } catch {
    manifest = { report: { status: "missing-or-nonregular" }, value: null };
  }
  return {
    inspection_status: "probe-ineligible",
    inspection_error_code: inspectionErrorCode(error),
    version: typeof manifest.value?.version === "string" ? manifest.value.version : null,
    codex_manifest: {
      path: ".codex-plugin/plugin.json",
      ...manifest.report,
      name: typeof manifest.value?.name === "string" ? manifest.value.name : null,
      version: typeof manifest.value?.version === "string" ? manifest.value.version : null,
      skills_field: fieldShapeEvidence(manifest.value?.skills),
      hooks_field: fieldShapeEvidence(manifest.value?.hooks),
    },
  };
}

function parseHookOutput(stdout, expectedEvent, expectedMode) {
  if (stdout.byteLength === 0) {
    return { json_object: false, system_message_matches: null, hook_event_matches: null, instruction: null };
  }
  try {
    const parsed = JSON.parse(stdout.toString("utf8"));
    const isObject = Boolean(parsed && typeof parsed === "object" && !Array.isArray(parsed));
    const context = parsed?.hookSpecificOutput?.additionalContext;
    return {
      json_object: isObject,
      system_message_matches: parsed?.systemMessage === `PONYTAIL:${expectedMode.toUpperCase()}`,
      hook_event_matches: context === undefined
        ? null
        : parsed?.hookSpecificOutput?.hookEventName === expectedEvent,
      instruction: typeof context === "string"
        ? contentEvidence(Buffer.from(context, "utf8"))
        : null,
    };
  } catch {
    return { json_object: false, system_message_matches: false, hook_event_matches: false, instruction: null };
  }
}

function summarizeHook(name, expectedEvent, expectedMode, result, durationMs) {
  const stdout = Buffer.isBuffer(result.stdout) ? result.stdout : Buffer.from(result.stdout || "");
  const stderr = Buffer.isBuffer(result.stderr) ? result.stderr : Buffer.from(result.stderr || "");
  return {
    name,
    expected_event: expectedEvent,
    expected_mode: expectedMode,
    execution: result.execution ?? (result.error?.code === "ETIMEDOUT" ? "timed-out" : result.error ? "spawn-error" : "completed"),
    error_code: result.error?.code ?? result.errorCode ?? null,
    exit_code: result.status ?? result.exitCode ?? null,
    signal: result.signal ?? null,
    duration_ms: Math.round(durationMs),
    stdout: contentEvidence(stdout),
    stderr: contentEvidence(stderr),
    contract: parseHookOutput(stdout, expectedEvent, expectedMode),
  };
}

function runHook(script, env, cwd, input, name, event, mode) {
  const started = process.hrtime.bigint();
  const result = spawnSync(process.execPath, [script.absolute], {
    cwd,
    env,
    input: Buffer.from(input, "utf8"),
    timeout: script.timeout_ms,
    maxBuffer: MAX_HOOK_OUTPUT_BYTES,
  });
  const durationMs = Number(process.hrtime.bigint() - started) / 1e6;
  return summarizeHook(name, event, mode, result, durationMs);
}

function runOpenStdinHook(script, env, cwd, name, event, mode) {
  return new Promise((resolve) => {
    const stdout = [];
    const stderr = [];
    let stdoutBytes = 0;
    let stderrBytes = 0;
    let execution = "completed";
    let errorCode = null;
    const started = process.hrtime.bigint();
    const child = spawn(process.execPath, [script.absolute], { cwd, env, stdio: ["pipe", "pipe", "pipe"] });
    const capture = (target, chunk, isStdout) => {
      const next = (isStdout ? stdoutBytes : stderrBytes) + chunk.byteLength;
      if (next > MAX_HOOK_OUTPUT_BYTES) {
        execution = "output-limit";
        child.kill("SIGKILL");
        return;
      }
      target.push(chunk);
      if (isStdout) stdoutBytes = next;
      else stderrBytes = next;
    };
    child.stdout.on("data", (chunk) => capture(stdout, chunk, true));
    child.stderr.on("data", (chunk) => capture(stderr, chunk, false));
    child.on("error", (error) => {
      execution = "spawn-error";
      errorCode = error.code ?? "unknown";
    });
    const timer = setTimeout(() => {
      execution = "timed-out";
      child.kill("SIGKILL");
    }, script.timeout_ms);
    child.on("close", (exitCode, signal) => {
      clearTimeout(timer);
      const durationMs = Number(process.hrtime.bigint() - started) / 1e6;
      resolve(summarizeHook(name, event, mode, {
        stdout: Buffer.concat(stdout),
        stderr: Buffer.concat(stderr),
        exitCode,
        signal,
        execution,
        errorCode,
      }, durationMs));
    });
  });
}

function isolatedEnvironment(tempRoot, pluginRoot) {
  const pluginData = path.join(tempRoot, "plugin-data");
  const home = path.join(tempRoot, "home");
  const config = path.join(tempRoot, "config");
  fs.mkdirSync(pluginData, { recursive: true });
  fs.mkdirSync(home, { recursive: true });
  fs.mkdirSync(config, { recursive: true });
  const env = {
    PATH: path.dirname(process.execPath),
    HOME: home,
    USERPROFILE: home,
    TMPDIR: tempRoot,
    XDG_CONFIG_HOME: config,
    CLAUDE_CONFIG_DIR: path.join(tempRoot, "claude"),
    CLAUDE_PLUGIN_ROOT: pluginRoot,
    PLUGIN_DATA: pluginData,
    PONYTAIL_DEFAULT_MODE: "full",
  };
  if (process.env.SystemRoot) env.SystemRoot = process.env.SystemRoot;
  return { env, pluginData };
}

function matchingInstructions(left, right) {
  if (!left?.contract?.instruction || !right?.contract?.instruction) return null;
  return left.contract.instruction.sha256 === right.contract.instruction.sha256;
}

async function isolatedProbes(pluginRoot, runnable) {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-"));
  const { env, pluginData } = isolatedEnvironment(tempRoot, pluginRoot);
  const stateFile = path.join(pluginData, ".ponytail-active");
  const probes = [];
  try {
    const start = runHook(
      runnable.get("SessionStart"), env, tempRoot,
      JSON.stringify({ source: "startup", session_id: "isolated-c1" }),
      "session-start-full", "SessionStart", "full",
    );
    probes.push(start);
    const afterStart = readModeFile(stateFile);

    const subagent = runHook(
      runnable.get("SubagentStart"), env, tempRoot,
      JSON.stringify({ agent_type: "general-purpose", session_id: "isolated-c1" }),
      "subagent-start-full", "SubagentStart", "full",
    );
    probes.push(subagent);

    const matcherEnv = { ...env, PONYTAIL_SUBAGENT_MATCHER: "^general-purpose$" };
    const openStdin = await runOpenStdinHook(
      runnable.get("SubagentStart"), matcherEnv, tempRoot,
      "subagent-open-stdin", "SubagentStart", "full",
    );
    probes.push(openStdin);

    probes.push(runHook(
      runnable.get("UserPromptSubmit"), env, tempRoot,
      JSON.stringify({ prompt: "@ponytail ultra", session_id: "isolated-c1" }),
      "prompt-switch-ultra", "UserPromptSubmit", "ultra",
    ));
    const afterUltra = readModeFile(stateFile);

    const compact = runHook(
      runnable.get("SessionStart"), env, tempRoot,
      JSON.stringify({ source: "compact", session_id: "isolated-c1" }),
      "session-start-compact", "SessionStart", "full",
    );
    probes.push(compact);
    const afterCompact = readModeFile(stateFile);

    probes.push(runHook(
      runnable.get("UserPromptSubmit"), env, tempRoot,
      JSON.stringify({ prompt: "stop ponytail", session_id: "isolated-c1" }),
      "prompt-disable", "UserPromptSubmit", "off",
    ));
    const afterDisable = readModeFile(stateFile);

    const disabledSubagent = runHook(
      runnable.get("SubagentStart"), env, tempRoot,
      JSON.stringify({ agent_type: "general-purpose", session_id: "isolated-c1" }),
      "subagent-after-disable", "SubagentStart", "off",
    );
    probes.push(disabledSubagent);

    return {
      classification: "isolated code-path evidence; not Codex host activation or propagation evidence",
      state_observations: {
        after_session_start: afterStart,
        after_ultra_switch: afterUltra,
        after_compact_session_start: afterCompact,
        after_disable: afterDisable,
      },
      observations: {
        session_and_subagent_instruction_match: matchingInstructions(start, subagent),
        open_stdin_completed_without_eof: openStdin.execution === "completed" && openStdin.exit_code === 0,
        open_stdin_instruction_matches_session: matchingInstructions(start, openStdin),
        mode_persisted_across_compact:
          afterUltra.status === "valid" && afterCompact.status === "valid"
            ? afterUltra.value === afterCompact.value
            : null,
        compact_instruction_matches_initial_session: matchingInstructions(start, compact),
        disable_removed_state: afterDisable.status === "missing",
        disabled_subagent_stdout_empty: disabledSubagent.stdout.bytes === 0,
      },
      probes,
    };
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
}

async function probeVerifiedSnapshot(pluginRoot, expectedSha256) {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ponytail-c1-snapshot-"));
  const snapshotRoot = path.join(tempRoot, "artifact");
  const rootGit = path.join(pluginRoot, ".git");
  try {
    fs.cpSync(pluginRoot, snapshotRoot, {
      recursive: true,
      filter: (source) => path.resolve(source) !== rootGit,
    });
    const before = artifactFingerprint(snapshotRoot);
    if (!before.complete || before.sha256 !== expectedSha256) {
      return {
        status: "probe-ineligible",
        reason: "temporary snapshot hash or completeness mismatch",
        expected_artifact_sha256: expectedSha256,
        snapshot_artifact_sha256_before: before.sha256,
      };
    }

    let inspected;
    try {
      inspected = inspectInstallation(snapshotRoot, before);
    } catch (error) {
      return {
        status: "probe-ineligible",
        reason: inspectionErrorCode(error),
        expected_artifact_sha256: expectedSha256,
        snapshot_artifact_sha256_before: before.sha256,
      };
    }
    const observations = await isolatedProbes(snapshotRoot, inspected.runnable);
    let after = null;
    try {
      after = artifactFingerprint(snapshotRoot);
    } catch {
      // A hook introduced an unsupported entry. Treat every observation as invalid.
    }
    const unchanged = after?.complete === true && after.sha256 === expectedSha256;
    return {
      status: unchanged ? "completed" : "invalid-artifact-mutated-during-probe",
      expected_artifact_sha256: expectedSha256,
      snapshot_artifact_sha256_before: before.sha256,
      snapshot_artifact_sha256_after: after?.sha256 ?? null,
      observations_valid: unchanged,
      ...observations,
    };
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
}

function instructionEvidence(specs, roots) {
  return specs.map((spec) => {
    const separator = spec.indexOf("=");
    if (separator < 1) throw new Error(`instruction must be label=path: ${spec}`);
    const label = spec.slice(0, separator);
    const absolute = path.resolve(spec.slice(separator + 1));
    const stat = fs.statSync(absolute);
    if (!stat.isFile() || stat.size > MAX_FILE_BYTES) throw new Error(`instruction is not a bounded file: ${label}`);
    const content = fs.readFileSync(absolute);
    return { label, path: redactPath(absolute, roots), ...contentEvidence(content) };
  });
}

async function collect(args, environment = process.env) {
  if (!args.subject || !["local-installation", "upstream-reference"].includes(args.subject)) {
    throw new Error("--subject must be local-installation or upstream-reference");
  }
  if (!args.pluginRoot) throw new Error("--plugin-root is required");
  const pluginRoot = fs.realpathSync(args.pluginRoot);
  if (!fs.statSync(pluginRoot).isDirectory()) throw new Error("--plugin-root must be a directory");

  const artifact = artifactFingerprint(pluginRoot);
  let inspected;
  try {
    inspected = inspectInstallation(pluginRoot, artifact);
  } catch (error) {
    inspected = { report: staticManifestSummary(pluginRoot, error), runnable: null };
  }
  const stateDir = args.stateDir ? path.resolve(args.stateDir) : null;
  const homedir = os.homedir();
  const configPath = path.resolve(args.config || configLocation(environment, homedir));
  const roots = [
    ["<PONYTAIL_PLUGIN_ROOT>", pluginRoot],
    ["<PLUGIN_DATA>", stateDir],
    ["~", homedir],
  ];
  const config = readJsonFile(configPath);
  const defaultMode = resolveDefaultMode(environment.PONYTAIL_DEFAULT_MODE, config.value);
  const configDefault = config.value?.defaultMode;
  const stateSnapshot = stateDir
    ? readModeFile(path.join(stateDir, ".ponytail-active"))
    : { status: "unknown", value: null, reason: "--state-dir not supplied" };
  let probes = {
    status: "not-run",
    reason: "static collection is the default; use --probe-isolated with an expected artifact SHA-256",
  };
  if (args.probeIsolated) {
    const expected = String(args.expectArtifactSha256 || "").toLowerCase();
    if (!/^[0-9a-f]{64}$/.test(expected)) {
      probes = { status: "probe-ineligible", reason: "missing or invalid --expect-artifact-sha256" };
    } else if (expected !== artifact.sha256) {
      probes = {
        status: "probe-ineligible",
        reason: "expected artifact SHA-256 does not match static collection",
        expected_artifact_sha256: expected,
        observed_artifact_sha256: artifact.sha256,
      };
    } else if (!artifact.complete) {
      probes = { status: "probe-ineligible", reason: "artifact contains symlinks" };
    } else if (!inspected.runnable) {
      probes = { status: "probe-ineligible", reason: "manifest or hook wiring is not probe-eligible" };
    } else {
      probes = await probeVerifiedSnapshot(pluginRoot, expected);
    }
  }

  return {
    schema: "ponytail-c1-audit-v1",
    collected_at_utc: new Date().toISOString(),
    subject: args.subject,
    evidence_boundary: {
      installation_role: args.subject === "local-installation" ? "primary" : "reference-only",
      isolated_hook_probes: "code-path evidence only",
      current_codex_app_activation: "unknown",
      live_subagent_propagation: "unknown",
      current_codex_hook_runtime: "unknown",
    },
    collector: {
      path: "experiments/ponytail-c1/collect-audit.mjs",
      sha256: sha256(fs.readFileSync(SCRIPT_PATH)),
    },
    installation: {
      root: "<PONYTAIL_PLUGIN_ROOT>",
      git: gitEvidence(pluginRoot),
      artifact: {
        algorithm: artifact.algorithm,
        scope_excludes: ["<PONYTAIL_PLUGIN_ROOT>/.git"],
        sha256: artifact.sha256,
        complete: artifact.complete,
        file_count: artifact.file_count,
        total_bytes: artifact.total_bytes,
      },
      ...inspected.report,
    },
    isolated_probe_runtime: {
      status: probes.status,
      node: process.version,
      platform: process.platform,
      arch: process.arch,
      child_environment: "bounded allowlist; temporary cwd/HOME/config/PLUGIN_DATA",
    },
    mode_sources: {
      environment: hashedValue(environment.PONYTAIL_DEFAULT_MODE, DEFAULT_MODES),
      subagent_matcher: matcherEvidence(environment.PONYTAIL_SUBAGENT_MATCHER),
      config: {
        path: redactPath(configPath, roots),
        ...config.report,
        default_mode: config.report.status === "valid"
          ? hashedValue(configDefault, DEFAULT_MODES)
          : { status: "unknown" },
      },
      resolved_default: defaultMode,
      active_state: {
        path: "<PLUGIN_DATA>/.ponytail-active",
        ...stateSnapshot,
      },
    },
    instructions: instructionEvidence(args.instructions, roots),
    isolated_hook_probes: probes,
    unknowns: [
      "current_codex_app_activation",
      "live_subagent_propagation",
      "current_codex_hook_runtime",
      "live_hook_exit_and_timeout_behavior",
    ],
  };
}

async function main() {
  try {
    const args = parseArgs(process.argv.slice(2));
    if (args.help) {
      process.stdout.write(usage());
      return;
    }
    const report = await collect(args);
    const serialized = JSON.stringify(report, null, 2) + "\n";
    if (args.out) fs.writeFileSync(args.out, serialized, { encoding: "utf8", flag: "wx" });
    else process.stdout.write(serialized);
  } catch (error) {
    process.stderr.write(`collect-audit: ${error.message}\n`);
    process.exitCode = 2;
  }
}

if (process.argv[1] && pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url) {
  await main();
}

export { artifactFingerprint, collect, parseArgs, redactPath, resolveDefaultMode };
