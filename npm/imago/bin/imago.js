#!/usr/bin/env node
// Launcher for the imago CLI. The actual program is a small Go binary; this
// script finds the prebuilt binary that matches the current OS/CPU (installed
// as an optional, platform-specific dependency) and runs it transparently.
"use strict";

const { spawnSync } = require("child_process");

function resolveBinary() {
  const platform = process.platform; // 'darwin' | 'linux' | 'win32' | ...
  const arch = process.arch; //         'x64' | 'arm64' | ...
  const pkg = `@singhvibhanshu/imago-${platform}-${arch}`;
  const exe = platform === "win32" ? "imago.exe" : "imago";
  try {
    return require.resolve(`${pkg}/bin/${exe}`);
  } catch (_) {
    return null;
  }
}

const binary = resolveBinary();

if (!binary) {
  console.error(
    `imago: no prebuilt binary available for ${process.platform}-${process.arch}.\n` +
      `If your platform should be supported, please file an issue with this info.`
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(`imago: failed to run binary: ${result.error.message}`);
  process.exit(1);
}

// Forward the child's exit code (or signal) so scripts behave correctly.
if (result.signal) {
  process.kill(process.pid, result.signal);
} else {
  process.exit(result.status === null ? 1 : result.status);
}
