# Quest

A BM25-powered shell history search tool written in Go. Like `Ctrl+R` but ranked — search across your entire bash history by token relevance, not just recency. No external services, no cloud, no LLM. Just your history, a sqlite DB, and a bubbletea TUI.

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
- `index/` — tokenizer and BM25 scorer. Strips shell delimiters (`-`, `/`, `=`, `:`) and splits into searchable terms. Scores each command using IDF (rare tokens rank higher) and TF with length normalization.
- `db/` — SQLite store via `go-sqlite3`. Upserts commands — repeated commands bump frequency rather than creating duplicate rows.
- `tui/` — bubbletea TUI. Model/Update/View pattern. Live BM25 search on every keystroke, arrow key navigation, enter to select.

---

## How BM25 Works Here

Standard `Ctrl+R` matches substrings in order. Quest scores every command in your history against your query simultaneously, ranking by relevance.

Every command is tokenized when indexed:
```
ffmpeg -i input.mp4 -vf scale=1280:720 output.mp4
→ ["ffmpeg", "i", "input", "mp4", "vf", "scale", "1280", "720", "output"]
```

At search time, BM25 scores each command using:

- **IDF** — tokens that appear in few commands score higher. `ffmpeg` in 4 commands scores far higher than `git` in 400. Rare tokens carry more signal.
- **TF** — how often the query token appears in this command, with diminishing returns.
- **Length normalization** — longer commands don't unfairly dominate just by containing more words.

Searching `ffmpeg scale` surfaces your video conversion command even if you don't remember the exact flags. Searching `follower port leader` surfaces your replication run commands. The more specific your tokens, the better the results.

---

## Limitations

BM25 matches tokens, not meaning. Searching `convert video to gif` will not find `ffmpeg -i input.mp4 output.gif` because none of those words overlap. Search using words that actually appear in the command — `ffmpeg`, `mp4`, `scale`, etc.

This is the honest ceiling of the approach. Quest is a ranked `Ctrl+R`, not a semantic search engine. It has zero external dependencies and that is intentional.

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

Tokens are preprocessed and stored — search never re-tokenizes, it just splits on spaces. The BM25 index is built in-memory at search time from all stored commands.

On conflict (same command run again), frequency is bumped and `last_used` is updated.

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

After the initial `quest sync`, the hook keeps the index current automatically. If quest isn't installed or the binary isn't in PATH, the hook fails silently — no impact on your shell.

---

## Usage

```bash
# first time — index your existing history
quest sync

# open interactive TUI
quest ui

# headless search — prints top result to stdout
quest "ffmpeg scale mp4"

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

**Dependencies:** Go 1.22+, gcc (for go-sqlite3 CGO). No external services.

---

## Roadmap

- [ ] Clipboard support — copy selected command instead of running it
- [ ] `q` shell function that runs the selected command in-place
- [ ] Filter by recency / frequency
- [ ] Exclude patterns (`.questignore`)
