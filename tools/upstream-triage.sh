#!/usr/bin/env bash
# upstream-triage — 上游高价值提交甄别账本
#
# 背景：本 fork 已 clean-room 重写并切换 MIT，上游为 LGPL。
# ┌─────────────────────────────────────────────────────────────┐
# │ 红线：永远不许 git cherry-pick 上游提交（哪怕 pick 后再改）。   │
# │ 正确姿势：读 commit message 与行为差异 → 关掉上游代码 →        │
# │ 在本仓库架构里独立重实现，提交时带 trailer: Upstream-Ref: <sha> │
# └─────────────────────────────────────────────────────────────┘
#
# 用法：
#   tools/upstream-triage.sh scan [--all]        列出水位线之上、未入台账的高价值提交
#   tools/upstream-triage.sh show <sha>          看某条上游提交的 message + 改动面（不看代码细节）
#   tools/upstream-triage.sh mark <sha> <ported|skipped|deferred> [备注] [本地SHA]
#   tools/upstream-triage.sh seal [--force]      水位线推进到 upstream/main（有未甄别项时需 --force）
#   tools/upstream-triage.sh status              摘要
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)
WM_FILE="$ROOT/docs/upstream/WATERMARK"
LEDGER="$ROOT/docs/upstream/LEDGER.md"
REMOTE=${UPSTREAM_REMOTE:-upstream}
BRANCH=${UPSTREAM_BRANCH:-main}

# 高价值启发式：安全/正确性/稳定性；UI 与杂务一律无关（本 fork 前端已完全重写）
HI_VALUE='fix|secur|cve|vuln|leak|race|panic|crash|deadlock|overflow|inject|corrupt'
EXCLUDES=(':(exclude)frontend' ':(exclude)docs' ':(exclude).github' ':(exclude)*.md')

watermark() { cat "$WM_FILE"; }

in_ledger() { # $1 = short or full sha
  local s; s=$(git rev-parse --short=8 "$1")
  grep -q "| \`$s\`" "$LEDGER" 2>/dev/null
}

cmd_scan() {
  local all=${1:-}
  git fetch -q "$REMOTE"
  local wm range
  wm=$(watermark)
  range="$wm..$REMOTE/$BRANCH"
  local args=(log "$range" --no-merges --date=short --pretty='%h|%ad|%s')
  if [[ "$all" != "--all" ]]; then
    args+=(--extended-regexp --regexp-ignore-case --grep="$HI_VALUE")
  fi
  local n=0
  while IFS='|' read -r sha date subject; do
    [[ -z "$sha" ]] && continue
    in_ledger "$sha" && continue
    # 纯排除路径内的改动跳过（改动面与排除集求差为空 → 无关）
    if ! git show --stat --oneline "$sha" -- . "${EXCLUDES[@]}" | tail -n +2 | grep -q .; then
      continue
    fi
    n=$((n+1))
    printf '%2d. %s  %s  %s\n' "$n" "$sha" "$date" "$subject"
  done < <(git "${args[@]}")
  if [[ $n -eq 0 ]]; then
    echo "✓ 水位线 $wm 之上没有待甄别的高价值提交"
  else
    echo
    echo "→ 逐条处理：tools/upstream-triage.sh show <sha>，然后 mark <sha> ported|skipped|deferred [备注]"
  fi
}

cmd_show() {
  local sha=$1
  git show -s --format='%h %ad%n%n%B' --date=short "$sha"
  echo '── 改动面（stat，不含代码）──'
  git show --stat --format='' "$sha"
}

cmd_mark() {
  local sha verdict note local_sha
  sha=$(git rev-parse --short=8 "$1"); verdict=$2; note=${3:-}; local_sha=${4:--}
  case "$verdict" in ported|skipped|deferred) ;; *) echo "verdict 必须是 ported|skipped|deferred"; exit 1;; esac
  if in_ledger "$sha"; then echo "⚠ $sha 已在台账中"; exit 1; fi
  local subject; subject=$(git show -s --format='%s' "$sha")
  printf '| `%s` | %s | %s | %s | %s |\n' "$sha" "$(date +%F)" "$verdict" "$local_sha" "${note:-$subject}" >> "$LEDGER"
  echo "✓ 已记账：$sha → $verdict"
  if [[ "$verdict" == "ported" && "$local_sha" == "-" ]]; then
    echo "  提示：重实现的本地提交记得带 trailer  Upstream-Ref: $sha"
  fi
}

cmd_seal() {
  local force=${1:-}
  local pending; pending=$(cmd_scan | grep -c '^\s*[0-9]' || true)
  if [[ "$pending" -gt 0 && "$force" != "--force" ]]; then
    echo "✗ 还有 $pending 条高价值提交未甄别，先 mark 完或用 --force"
    exit 1
  fi
  git rev-parse "$REMOTE/$BRANCH" > "$WM_FILE"
  echo "✓ 水位线已推进到 $(cat "$WM_FILE")"
}

cmd_status() {
  echo "水位线: $(watermark)"
  echo "  落后上游: $(git rev-list --count "$(watermark)..$REMOTE/$BRANCH" 2>/dev/null || echo '?（先 fetch）') 个提交"
  for v in ported skipped deferred; do
    printf '  %-8s %s\n' "$v" "$(grep -c "| $v |" "$LEDGER" 2>/dev/null || echo 0)"
  done
}

case "${1:-scan}" in
  scan)   cmd_scan "${2:-}" ;;
  show)   cmd_show "$2" ;;
  mark)   shift; cmd_mark "$@" ;;
  seal)   cmd_seal "${2:-}" ;;
  status) cmd_status ;;
  *) sed -n '2,20p' "$0"; exit 1 ;;
esac
