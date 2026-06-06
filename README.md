# Quest

A fuzzy shell history search tool written in Go. Like `Ctrl+R` but you describe what the command *does* instead of remembering what it *is*. Built to understand how search ranking algorithms work under the hood — tokenization, BM25 scoring, and TUI design.

---

## Architecture

```
~/.bash_history
       │
       ▼
┌───────────────────┐
│  History Parser   │  filters boring/duplicate commands
│  (history/)       │
└────────┬──────────┘
         │
         ▼
┌───────────────────┐        ┌─────────────────────┐
│   BM25 Index      │        │  ~/.local/share/     │
│   (index/)        │◀──────▶│  quest/quest.db      │
│                   │        └─────────────────────┘
│  IDF per token    │
│  avgLen norm      │  in-memory at search time
└────────┬──────────┘
         │
         ▼
┌───────────────────┐
│   Bubbletea TUI   │  live search on every keystroke
│   (tui/)          │
└───────────────────┘
```

### Packages

- `history/` — parses `~/.bash_history`. Filters empty lines, timestamp lines (`#1234567890`), and boring commands (`ls`, `cd`, `clear`, etc). Deduplicates before indexing.
- `index/` — tokenizer and BM25 scorer. Tokenizer strips shell delimiters (`-`, `/`, `=`, `:`) and splits into searchable terms. BM25 scores each command against a query using IDF (rare tokens score higher) and TF with length normalization.
- `db/` — SQLite store via `go-sqlite3`. Upserts commands — repeated commands bump frequency rather than creating duplicate rows.
- `tui/` — bubbletea TUI. Model/Update/View pattern. Live BM25 search on every keystroke, arrow key navigation, enter to select.

---

## How BM25 Works Here

Standard `Ctrl+R` matches substrings. Quest matches *meaning*.

Every command is tokenized when indexed:
```
ffmpeg -i input.mp4 -vf scale=1280:720 output.mp4
→ ["ffmpeg", "i", "input", "mp4", "vf", "scale", "1280", "720", "output"]
```

At search time, each token in your query is scored against every indexed command using BM25:

- **IDF** — tokens that appear in few commands score higher. `ffmpeg` in 4 commands scores far higher than `git` in 400.
- **TF** — how often the query token appears in this command.
- **Length normalization** — longer commands don't unfairly dominate just by containing more words.

The result: `convert video to gif` finds your `ffmpeg` command even though none of those words appear in it — because `ffmpeg`, `mp4`, and `scale` are rare tokens that correlate with video conversion commands in your history.

---

## Storage

Commands are stored in SQLite at `~/.local/share/quest/quest.db`.

Schema:
```sql
CREATE TABLE commands (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    command   TEXT UNIQUE NOT NULL,
    frequency INTEGER DEFAULT 1,
    last_used DATETIME NOT NULL,
    tokens    TEXT NOT NULL
);
```

Tokens are preprocessed and stored — so search never re-tokenizes, it just splits on spaces.

On conflict (same command run again), frequency is bumped and `last_used` is updated. The BM25 index is built in-memory at search time from all stored commands.

---

## Shell Hook

A `PROMPT_COMMAND` hook captures every command you run and silently indexes it:

```bash
_quest_capture() {
    local last_cmd
    last_cmd=$(history 1 | sed 's/^[[:space:]]*[0-9]*[[:space:]]*//')
    ( quest add "$last_cmd" &>/dev/null & disown )
}
PROMPT_COMMAND="${PROMPT_COMMAND}; _quest_capture"
```

After the initial `quest sync`, the hook keeps the index current automatically.

---

## Usage

```bash
# first time — index your existing history
quest sync

# open interactive TUI
quest ui

# headless search — prints top result to stdout
quest "convert video to gif"

# help
quest help
```

**Commands:**

| Command | Description |
|---|---|
| `quest sync` | index `~/.bash_history` into sqlite |
| `quest ui` | open interactive TUI |
| `quest add <cmd>` | index a single command (used by hook) |
| `quest <query>` | headless BM25 search, prints top result |
| `quest help` | show usage |

**TUI keybinds:**

| Key | Action |
|---|---|
| type | live search |
| `↑ ↓` | navigate results |
| `enter` | select command |
| `esc` | quit |

---

## Install

```bash
git clone https://github.com/kavinbharathii/quest.git
cd quest
go build -o quest .
cp quest ~/.local/bin/quest
```

Add to `~/.bashrc`:
```bash
source ~/Projects/quest/hook/hook.sh
```

Then:
```bash
source ~/.bashrc
quest sync
```

---

## Roadmap

- [ ] Clipboard support — copy selected command instead of printing
- [ ] `q` shell function that runs the selected command in-place
- [ ] Ollama embeddings backend for true semantic search
- [ ] Filter by recency / frequency
- [ ] Exclude patterns (`.questignore`)
