# Consul KV Sync Tool

A simple, safe, and flexible CLI tool to backup, export, and import Consul KV data between environments such as staging → local.

Perfect for developers who need to replicate environment configurations locally without manually copying keys.

## Features

### 🛡 1. Local Backup

This tool automatically creates a timestamped file:  
```
local_backup_YYYY-MM-DD_HHMMSS.json
```

### 📤 2. Export Staging Consul KV

Exports all staging KV entries into:
```
consul_export.json
```

### 📥 3. Import With Prefix Rewrite

Import supports:  
- 'Filtering keys using --from-prefix  
- Rewriting imported keys using --to-prefix  
- Importing into root. 
- Moving an entire folder to another folder. 

Examples:
```
serviceA/config/db → localA/config/db
serviceA/feature/x → feature/x   (import into root)
```

### ✔ 4. Key Safety Rules

The tool ensures:  
- No key will ever begin with / 
- No key will ever become empty. 
-Invalid/empty keys are skipped safely

### 📦 5. Simple CLI Flags

| Flag            | Description                              |
| --------------- | ---------------------------------------- |
| `--backup`      | Only backup local Consul KV              |
| `--export`      | Export staging KV → `consul_export.json` |
| `--import`      | Backup local → load export → import      |
| `--from-prefix` | Only import keys under this prefix       |
| `--to-prefix`   | Rewrite imported keys to this prefix     |

## 🚀 Getting Started

### ⚙ Environment Variables

| Variable               | Description           | Default                      |
| ---------------------- | --------------------- | ---------------------------- |
| `LOCAL_CONSUL_ADDR`    | Local Consul URL      | `http://localhost:8500`      |
| `LOCAL_CONSUL_TOKEN`   | ACL token (if needed) | empty                        |
| `STAGING_CONSUL_ADDR`  | Staging Consul URL    | `http://staging-consul:8500` |
| `STAGING_CONSUL_TOKEN` | ACL token for staging | empty                        |

Example:  
```
export STAGING_CONSUL_ADDR="<<replace_with_your_consul_url>>"
export STAGING_CONSUL_TOKEN="<<replace_with_your_token>>"
```

### 🛡 Backup Mode

Backs up local Consul into a timestamped JSON file:  
```
go run main.go --backup
```
Output:  
```
local_backup_2025-01-01_143233.json
```

You may restore by running:  
```
mv local_backup_2025-01-01_143233.json consul_export.json
go run main.go --import
```

### 
📤 Export From Staging

Exports all keys into consul_export.json:  
```
go run main.go --export
```

### 📥 Import Into Local (Safe Mode)

Import automatically:  
1. Backs up local Consul. 
2. Loads consul_export.json. 
3. Filters using from-prefix. 
4. Rewrites using to-prefix. 
5. Writes into local Consul

Basic import (all keys). 
```
go run main.go --import
```

### 🎯 Prefix Filtering & Rewriting

These flags make the tool extremely flexible.

#### 1️⃣ Import only a folder into root

```
go run main.go --import --from-prefix="serviceA/" --to-prefix=""
```

Result:
```
serviceA/config/db  →  config/db
serviceA/feature/x  →  feature/x
```

#### 2️⃣ Import folder → different folder

```
go run main.go --import \
  --from-prefix="serviceA/" \
  --to-prefix="localA/"
```

Result:
```
serviceA/config/db → localA/config/db
serviceA/feature/x → localA/feature/x
```

#### 3️⃣ Import everything into a new namespace

```
go run main.go --import --to-prefix="staging_copy/"
```

Result:
```
config/db → staging_copy/config/db
```

4️⃣ Import everything except root prefixes

If --from-prefix is empty, all keys are processed.

## 📁 File Architecture

```
main.go
README.md
consul_export.json            # created by --export
local_backup_*.json           # backups created automatically
```

## 🔒 Safety Behavior

The import logic guarantees:  
1. No invalid keys.Keys cannot begin with /.  
    - Keys cannot be empty.  
    - Empty or invalid rewritten keys are skipped. 
2. Backup happens before modification. 
Even if import fails midway, your previous state is safe.

## 🧪 Example Workflow

Copy staging → local, but only serviceA/, flatten into root:  
```
go run main.go --import --from-prefix="serviceA/" --to-prefix=""
```

Copy everything staging → local under imported/ 
```
go run main.go --import --to-prefix="imported/"
```

Export staging only:  
```
go run main.go --export
```

Backup then import:  
```
go run main.go --import
```

## 🔧 Future Enhancements (Optional)

If you want, I can add any of these:  
- --dry-run (see what would change without modifying Consul). 
- --exclude-prefix or multiple --from-prefix support. 
- Progress bar. 
- Parallel import (for large 10k+ keybases). 
- Export/import using Consul snapshots instead of JSON. 
- YAML support. 

Just tell me!