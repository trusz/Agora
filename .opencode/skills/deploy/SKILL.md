---
name: deploy
description: Deploy the Agora app to the server via SSH. Builds the executable, uploads it with a versioned filename (agora_<date>-<release>), and updates the symlink.
---

## What I do

Build and deploy the Agora executable to the server, versioned by date and release number.

## When to use me

Use when asked to deploy, release, or push Agora to the server.

## Workflow

### 1. Build the executable

```bash
cd /Users/trusz/Work/089_Agora
make build-linux
```

The output binary will be at `build/agora_<TODAY>`.

### 2. Determine the versioned filename

Today's date in `YYYY-MM-DD` format. Check existing files on server to find next release number:

```bash
ssh agora "ls /home/tamas/agora/agora_<TODAY>* 2>/dev/null"
```

- If none exist → use `agora_<TODAY>-1`
- If `agora_<TODAY>-1` exists → use `agora_<TODAY>-2`, etc.

### 3. Upload the executable

```bash
scp build/agora agora:/home/tamas/agora/<VERSIONED_NAME>
ssh agora "chmod +x /home/tamas/agora/<VERSIONED_NAME>"
```

### 4. Update the symlink

```bash
ssh agora "ln -sf /home/tamas/agora/<VERSIONED_NAME> /home/tamas/agora/agora"
```

### 5. Verify

```bash
ssh agora "ls -la /home/tamas/agora/agora"
```

Output should show symlink pointing to the new versioned file.

## Notes

- Upload target: `agora:/home/tamas/agora/`
- Server SSH alias: `agora`
- Remote deploy path: `/home/tamas/agora/`
- Symlink name: `agora` (no extension)
- Versioned format: `agora_YYYY-MM-DD-<N>` (N starts at 1)
