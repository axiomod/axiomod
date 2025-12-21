---
trigger: always_on
---

# CLI & Terminal Agent Guidelines

This document establishes the operational boundaries and safety protocols for the AI Coding Agent when interacting with the system shell.

## 1. The Golden Rule of Context
>
> **Never assume the current working directory (CWD). Always verify.**

* **Pre-execution Check:** Before running any command that relies on relative paths (e.g., `cd`, `ls`, `mkdir`, `cat`), the agent must be certain of the current directory.
* **Verification Command:** When in doubt, or at the start of a new session, execute `pwd` to confirm the CWD.

## 2. Directory Navigation & Creation (`cd`, `mkdir`)

### Navigation

* **Verification:** Before executing `cd <path>`, verify `<path>` exists using `ls -d <path>` or by checking the file tree context.
* **Relative vs. Absolute:** Prefer relative paths for readability, but use absolute paths if navigation becomes ambiguous.
* **Chain Breakers:** Do not chain `cd` commands blindly (e.g., `cd .. && cd folder`). If the first fails, the second executes in the wrong location.
  * *Bad:* `cd frontend; npm install`
  * *Good:* `cd frontend && npm install` (Ensures dependencies install only if navigation succeeds).

### Creation

* **Defensive Creation:** Always use the `-p` flag with `mkdir` (e.g., `mkdir -p path/to/dir`). This prevents errors if the parent directory is missing or if the directory already exists.
* **Naming Conventions:** Ensure directory names do not contain whitespace or special characters unless strictly necessary. If they do, strictly use quotes (e.g., `cd "my folder"`).

## 3. Destructive Command Safeguards (`rm`, `mv`)

* **Prohibited Flags:** The agent is strictly prohibited from running `rm -rf /` or `rm -rf ~` under any circumstances.
* **Scope Verification:** Before using `rm` with wildcards (e.g., `rm *.ts`), run `ls *.ts` first to verify exactly which files will be deleted.
* **Directory Removal:** When removing directories, prefer `rmdir` for empty directories to ensure safety. Use `rm -rf` only when explicitly tasked to delete a populated folder and context is confirmed.

## 4. Execution & Process Management

* **Long-running Processes:** Do not start blocking processes (like `npm start` or `python server.py`) without running them in the background or expecting the terminal to lock up.
  * *Preferred:* Instruct the user on how to start the server, or use a separate terminal instance if available.
* **Sudo Usage:** **Avoid `sudo`** unless explicitly requested by the user. If a command fails due to permission errors, report the error to the user rather than blindly attempting `sudo`.

## 5. Environment & Package Managers

* **Installation Checks:** Do not assume tools (Git, Node, Python, Docker) are installed.
  * *Pattern:* Check `node -v` or `which node` before running `npm install`.
* **Lockfiles:** Respect existing lockfiles (`package-lock.json`, `yarn.lock`, `poetry.lock`). Do not switch package managers (e.g., switching from npm to yarn) unless explicitly instructed.

## 6. Output & Error Handling

* **Silence is not Success:** If a command produces no output, verify the result (e.g., after `touch file.txt`, check if it exists).
* **Read the Error:** If a command returns a non-zero exit code, **read the stderr**. Do not blindly retry the exact same command. Analyze the error message to adjust the approach.

## 7. File Editing via CLI

* **Echo/Cat:** For creating simple files, `echo "content" > file.txt` is acceptable.
* **Complex Edits:** For complex code edits, do not use `sed` or `awk` unless you are absolutely certain of the regex. Prefer rewriting the file content fully using the editor's file system API rather than terminal manipulation.
