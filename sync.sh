#!/usr/bin/env bash
set -euo pipefail

# sync.sh - Directory synchronization with:
#   - .syncignore (src/dest), per-side ignores, unignore (!) support
#   - Optional whitelist ("only") mode
#   - Optional import of SOURCE/.gitignore
#   - One-way and two-way operation
#   - Optional config (or pure CLI with --source/--dest)
#   - Default exclusion of .git/ and optional exclusion of all hidden dirs
#
# Requires: rsync >= 3.1 (3.2+ preferred)

CONFIG=""
DRY_RUN=0
#!/usr/bin/env bash
set -euo pipefail
echo "[deprecated] sync.sh has been archived. Please use the Python CLI instead." >&2
echo "Usage: sync-tools sync --source SRC --dest DEST [options]" >&2
echo "See README for single-file builds (PEX, shiv, PyInstaller)." >&2
exit 1
if [[ -n "$CLI_MODE" ]]; then MODE="$CLI_MODE"; fi

# 4) Validate presence of SOURCE/DEST
if [[ -z "${SOURCE:-}" || -z "${DEST:-}" ]]; then
  print_usage
  die "You must provide SOURCE and DEST via config or --source/--dest"
fi

# 5) Default MODE if still unset
MODE="${MODE:-one-way}"

# Normalize/validate MODE
MODE_LOWER="$(echo "$MODE" | tr '[:upper:]' '[:lower:]')"
if [[ "$MODE_LOWER" != "one-way" && "$MODE_LOWER" != "two-way" ]]; then
  die "MODE must be 'one-way' or 'two-way' (got: $MODE)"
fi

# Rsync opts
RSYNC_OPTS=(-a -v --delete --human-readable --itemize-changes --partial)
if rsync --version 2>/dev/null | grep -q 'version 3\.[2-9]'; then
  RSYNC_OPTS+=(--mkpath)
fi
[[ $DRY_RUN -eq 1 ]] && RSYNC_OPTS+=(--dry-run)

ensure_trailing_slash() {
  local p="$1"
  if [[ "$p" != */ ]]; then printf "%s/\n" "$p"; else printf "%s\n" "$p"; fi
}

# Prepare SRC/DST with trailing slashes for rsync
SRC="$(ensure_trailing_slash "$SOURCE")"
DST="$(ensure_trailing_slash "$DEST")"

# Collect config-level excludes into arrays
CONFIG_EXCLUDE_FILES=()
CONFIG_EXCLUDE_PATS=()
[[ -n "$EXCLUDES_FILE" ]] && CONFIG_EXCLUDE_FILES+=("$EXCLUDES_FILE")
if [[ -n "$EXCLUDE" ]]; then
  if declare -p EXCLUDE 2>/dev/null | grep -q 'declare \-a'; then
    CONFIG_EXCLUDE_PATS+=("${EXCLUDE[@]}")
  else
    # shellcheck disable=SC2206
    TMP_SPLIT=(${EXCLUDE//,/ })
    CONFIG_EXCLUDE_PATS+=("${TMP_SPLIT[@]}")
  fi
fi

# Helpers for filter building
to_filter_rule() {
  local pat="$1"
  if [[ "$pat" == !* ]]; then
    local p="${pat:1}"
    # Strip trailing /** if present to get base path
    local base="$p"
    if [[ "$base" == */** ]]; then
      base="${base%/**}"
    fi

    # Was the pattern anchored (leading slash)?
    local has_lead=0
    if [[ "$base" == /* ]]; then has_lead=1; fi

    # Build parent includes (no leading slash for splitting)
    local rel="${base#/}"
    IFS='/' read -r -a parts <<<"$rel"
    local accum=""
    for ((i=0;i<${#parts[@]}-1;i++)); do
      accum+="${parts[i]}/"
      if [[ $has_lead -eq 1 ]]; then
        printf -- "+ /%s\n" "$accum"
      else
        printf -- "+ %s\n" "$accum"
      fi
    done

    # Include the base (as a directory) and its recursive children
    if [[ $has_lead -eq 1 ]]; then
      printf -- "+ %s/\n" "${base%/}"
      printf -- "+ %s/**\n" "${base%/}"
    else
      printf -- "+ %s/\n" "${base%/}"
      printf -- "+ %s/**\n" "${base%/}"
    fi
  else
    printf "- %s\n" "$pat"
  fi
}

clean_to_filter_file() {
  local in="$1"
  [[ -f "$in" ]] || die "Ignore file not found: $in"
  local out; out="$(mktemp)"
  while IFS= read -r line || [[ -n "$line" ]]; do
    # Trim
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" ]] && continue
    [[ "$line" =~ ^# ]] && continue
    to_filter_rule "$line" >>"$out"
  done <"$in"
    if [ "${SYNC_DEBUG:-}" = "1" ]; then
      # preserve a copy for debugging (name includes pid and timestamp)
      cp "$out" "/tmp/sync_debug_filter_$$.$(date +%s%3N)" 2>/dev/null || true
    fi
  echo "$out"
}

patterns_to_filter_file() {
  local out; out="$(mktemp)"
  for pat in "$@"; do
    [[ -n "$pat" ]] || continue
    to_filter_rule "$pat" >>"$out"
  done
  if [ "${SYNC_DEBUG:-}" = "1" ]; then
    cp "$out" "/tmp/sync_debug_filter_$$.$(date +%s%3N)" 2>/dev/null || true
  fi
  echo "$out"
}

# Whitelist support
collect_only_items() {
  local -n out_arr="$1"
  out_arr=()
  if [[ -n "$ONLY_LIST_FILE" ]]; then
    [[ -f "$ONLY_LIST_FILE" ]] || die "ONLY_LIST_FILE not found: $ONLY_LIST_FILE"
    while IFS= read -r line || [[ -n "$line" ]]; do
      line="${line#"${line%%[![:space:]]*}"}"
      line="${line%"${line##*[![:space:]]}"}"
      [[ -z "$line" ]] && continue
      [[ "$line" =~ ^# ]] && continue
      out_arr+=("$line")
    done <"$ONLY_LIST_FILE"
  fi
  if [[ ${#ONLY_ITEMS[@]} -gt 0 ]]; then
    out_arr+=("${ONLY_ITEMS[@]}")
  fi
}

build_only_filter_file() {
  local -a only_list=()
  collect_only_items only_list
  [[ ${#only_list[@]} -gt 0 ]] || return 1

  local out; out="$(mktemp)"
  # Write include rules first, then exclude everything else. Including recursive
  # patterns (p/**) ensures directories and their contents are allowed.
  for path in "${only_list[@]}"; do
    local p="${path#./}"
    [[ -z "$p" ]] && continue
    # Include parent directories so rsync can create them on DEST
    IFS='/' read -r -a parts <<<"$p"
    local accum=""
    for ((i=0;i<${#parts[@]}-1;i++)); do
      accum+="${parts[i]}/"
      printf "+ %s\n" "$accum" >>"$out"
    done
    # Allow the path itself and any children beneath it
    printf "+ %s\n" "$p" >>"$out"
    printf "+ %s/**\n" "$p" >>"$out"
  done
  # Exclude everything else
  echo "- *" >>"$out"
  echo "$out"
}

TMPS_TO_CLEAN=()
cleanup() {
  for t in "${TMPS_TO_CLEAN[@]:-}"; do
    [[ -n "$t" && -f "$t" ]] && rm -f "$t"
  done
}
trap cleanup EXIT

# Build per-side filters with precedence:
#   whitelist, default filters (.git exclusion, optional hidden dirs exclusion),
#   .syncignore, .gitignore (src), config files, config patterns,
#   CLI files, CLI patterns
build_side_filters() {
  local side="$1"       # "src" or "dest"
  local -n out_arr="$2" # output array

  local root=""
  local use_sync=1
  if [[ "$side" == "src" ]]; then
    root="${SOURCE%/}"
    use_sync=$USE_SRC_SYNCIGNORE
  else
    root="${DEST%/}"
    use_sync=$USE_DST_SYNCIGNORE
  fi

  # Whitelist (applies to both sides, same list)
  local only_file=""
  # Apply whitelist-only mode only to the SOURCE side. Applying it to DEST
  # can prevent rsync from creating files because the DEST side would be
  # excluded by the initial "- *" rule. Only use the whitelist for the
  # source side to limit what is sent.
  if [[ "$side" == "src" ]]; then
    if only_file="$(build_only_filter_file)"; then
      TMPS_TO_CLEAN+=("$only_file")
      out_arr+=(--filter ". $only_file")
      info "Using whitelist-only mode"
    fi
  fi

  # NOTE: default filters (like excluding .git/ and hidden dirs) are appended
  # at the end of this function so that user-provided include/unignore rules
  # (from .syncignore, .gitignore, config, or CLI) can appear before the
  # final exclusions and therefore override them when necessary.

  # .syncignore
  if [[ $use_sync -eq 1 && -f "$root/.syncignore" ]]; then
    local cfi; cfi="$(clean_to_filter_file "$root/.syncignore")"
    TMPS_TO_CLEAN+=("$cfi")
    out_arr+=(--filter ". $cfi")
    info "Using $side .syncignore"
  fi

  # SOURCE .gitignore (opt-in)
  if [[ "$side" == "src" && $USE_SOURCE_GITIGNORE -eq 1 && -f "${SOURCE%/}/.gitignore" ]]; then
    local gfi; gfi="$(clean_to_filter_file "${SOURCE%/}/.gitignore")"
    TMPS_TO_CLEAN+=("$gfi")
    out_arr+=(--filter ". $gfi")
    info "Using SOURCE .gitignore"
  fi

  if [[ $ONLY_SYNCIGNORE -eq 0 ]]; then
    for f in "${CONFIG_EXCLUDE_FILES[@]}"; do
      local cfi; cfi="$(clean_to_filter_file "$f")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    done
    if [[ ${#CONFIG_EXCLUDE_PATS[@]} -gt 0 ]]; then
      local cfi; cfi="$(patterns_to_filter_file "${CONFIG_EXCLUDE_PATS[@]}")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    fi
  fi

  if [[ "$side" == "src" ]]; then
    for f in "${CLI_IGNORE_SRC_FILES[@]}"; do
      local cfi; cfi="$(clean_to_filter_file "$f")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    done
    if [[ ${#CLI_IGNORE_SRC_PATTERNS[@]} -gt 0 ]]; then
      local cfi; cfi="$(patterns_to_filter_file "${CLI_IGNORE_SRC_PATTERNS[@]}")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    fi
  else
    for f in "${CLI_IGNORE_DST_FILES[@]}"; do
      local cfi; cfi="$(clean_to_filter_file "$f")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    done
    if [[ ${#CLI_IGNORE_DST_PATTERNS[@]} -gt 0 ]]; then
      local cfi; cfi="$(patterns_to_filter_file "${CLI_IGNORE_DST_PATTERNS[@]}")"
      TMPS_TO_CLEAN+=("$cfi")
      out_arr+=(--filter ". $cfi")
    fi
  fi

  # Append default filters last (lowest precedence), so that explicit include
  # rules generated from user patterns can override the defaults.
  # - Always exclude top-level .git/ (both sides)
  out_arr+=(--filter "- /.git/")
  # - Optionally exclude all hidden directories (both sides)
  if [[ $EXCLUDE_HIDDEN_DIRS -eq 1 ]]; then
    out_arr+=(--filter "- .*/")
  fi
}

FILTER_ARGS_SRC=()
FILTER_ARGS_DST=()
build_side_filters "src" FILTER_ARGS_SRC
build_side_filters "dest" FILTER_ARGS_DST

# One-way (SOURCE -> DEST)
run_one_way() {
  info "One-way sync: $SOURCE -> $DEST"
  if [[ "${SYNC_DEBUG:-0}" == "1" ]]; then
    echo "[sync-debug] RSYNC_OPTS=${RSYNC_OPTS[*]}" >>/tmp/sync_debug.log
    echo "[sync-debug] FILTER_ARGS_SRC=${FILTER_ARGS_SRC[*]}" >>/tmp/sync_debug.log
    echo "[sync-debug] FILTER_ARGS_DST=${FILTER_ARGS_DST[*]}" >>/tmp/sync_debug.log
    echo "[sync-debug] SRC=$SRC DST=$DST" >>/tmp/sync_debug.log
    echo "[sync-debug] RSYNC_CMD: rsync ${RSYNC_OPTS[*]} ${FILTER_ARGS_SRC[*]} ${FILTER_ARGS_DST[*]} $SRC $DST" >>/tmp/sync_debug.log
  fi
  # Pass source filters first, then destination filters. Putting source-side
  # include/unignore rules early in the combined filter list ensures they
  # can override later default destination exclusions (e.g. - /.git/).
  rsync "${RSYNC_OPTS[@]}" \
    "${FILTER_ARGS_SRC[@]}" \
    "${FILTER_ARGS_DST[@]}" \
    "$SRC" "$DST"
}

# Two-way
twoway_pass() {
  local from="$1"
  local to="$2"
  local side_from="$3" # "src" or "dest"
  local side_to="$4"   # "src" or "dest"

  local -n filter_from="FILTER_ARGS_${side_from^^}"
  local -n filter_to="FILTER_ARGS_${side_to^^}"

  rsync -a --update --inplace --no-owner --no-group \
    --times --omit-dir-times \
    --human-readable --itemize-changes \
    "${filter_from[@]}" \
    "${filter_to[@]}" \
    ${DRY_RUN:+--dry-run} \
    --delete-delay \
    --copy-dirlinks \
    "$(ensure_trailing_slash "$from")" \
    "$(ensure_trailing_slash "$to")"
}

mark_conflict() {
  local path="$1"
  local ts; ts="$(date +%Y%m%d-%H%M%S)"
  echo "$path.conflict-$ts"
}

handle_two_way() {
  info "Two-way sync between: $SOURCE <-> $DEST"
  twoway_pass "$SOURCE" "$DEST" "src" "dst"
  twoway_pass "$DEST" "$SOURCE" "dst" "src"

  if [[ $DRY_RUN -eq 1 ]]; then
    warn "Conflict detection skipped in dry-run."
    return
  fi

  # Simple conflict preservation
  TMP_A=$(mktemp); TMP_B=$(mktemp)
  TMPS_TO_CLEAN+=("$TMP_A" "$TMP_B")
  (cd "$SOURCE" && find . -type f | sort) >"$TMP_A"
  (cd "$DEST" && find . -type f | sort) >"$TMP_B"

  comm -12 "$TMP_A" "$TMP_B" | while IFS= read -r rel; do
    A_FILE="$SOURCE/${rel#./}"
    B_FILE="$DEST/${rel#./}"
    [[ -f "$A_FILE" && -f "$B_FILE" ]] || continue
    A_SUM="$(cksum < "$A_FILE" | awk '{print $1":"$2}')"
    B_SUM="$(cksum < "$B_FILE" | awk '{print $1":"$2}')"
    if [[ "$A_SUM" != "$B_SUM" ]]; then
      CONFLICT_PATH="$(mark_conflict "$A_FILE")"
      warn "Conflict: $rel -> preserving DEST as $(basename "$CONFLICT_PATH")"
      mkdir -p "$(dirname "$CONFLICT_PATH")"
      cp -p "$B_FILE" "$CONFLICT_PATH"
    fi
  done
}

main() {
  case "$MODE_LOWER" in
    one-way) run_one_way ;;
    two-way) handle_two_way ;;
    *) die "Unsupported MODE: $MODE" ;;
  esac
  info "Done."
}

main
exit 0