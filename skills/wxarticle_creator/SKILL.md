---
name: wxarticle_creator
description: Draft Weixin official account articles from user draft.
metadata:
  short-description: Weixin draft via markdown-to-html-cli and wxcli
---

# WX Article Creator

Use this skill when the user asks to draft or create a Weixin official
account article (plain text or Markdown) and submit it as a draft.

---

## Inputs

- `draft`: User-provided draft text (plain or Markdown)
- `title`: Article title
- `thumb_media_id`: Weixin thumbnail media id for draft
- `format`: Optional enum (`auto`, `markdown`, `plain`)

## Outputs

- `markdown`: Final markdown source
- `html`: Rendered HTML
- `wxcli_result`: Command output

---

## Rules

- Do not invent wxcli commands; use README-documented format.
- Always ask for user confirmation before running `wxcli`.
- Render HTML using exactly:
  `npx markdown-to-html-cli --source <SOURCE.md> --style=./style.css`
- CSS must be loaded from `skills/wxarticle_creator/style.css`.
- Preserve meaning; no new facts. Keep length within +/-10% when polishing.

---

## Workflow

1. Detect input format.
   - If `format=auto`, detect Markdown by headings (`#`), lists (`-`/`1.`), links,
     or code fences.
   - If Markdown is provided, keep it.
   - If plain text, convert it to Markdown with clear hierarchy.

2. Convert plain text to structured Markdown (if needed).
   - Use the provided `title` as an H1.
   - Derive H2 sections from topic shifts or leading phrases.
   - Use H3 for subtopics, and bullet/numbered lists for enumerations.
   - Keep original ordering and meaning.

3. Polish Markdown lightly.
   - Fix typos and improve readability.
   - Keep structure; do not add new facts.

4. Write Markdown source to a temporary file in a temp directory.
   - Create a temp directory (e.g., `/tmp/wxarticle_creator-<YYYYMMDD-HHMMSS>`).
   - Save as `<TMP_DIR>/article-<YYYYMMDD-HHMMSS>.md` (or another `<SOURCE.md>`).
   - Remove the temp directory after HTML is generated, unless the user requests it.

5. Convert Markdown to HTML with the CLI.
   - Copy `skills/wxarticle_creator/style.css` into the temp directory as `style.css`.
   - Run the command from the temp directory so `./style.css` resolves.
   - Command:
     `npx markdown-to-html-cli --source <SOURCE.md> --style=./style.css`
   - Capture stdout as HTML output.
   - Ensure the content is wrapped in `<div id="nice">...</div>` if it is not
     already, so the stylesheet applies.

6. User confirmation.
   - Show polished markdown and a rendered HTML snippet.
   - Ask: "Proceed to create Weixin draft with wxcli?"

7. Submit via wxcli.
   - Direct content:
     `wxcli draft add --title "<title>" --content "<html>" --thumb-media-id <thumb_media_id>`
   - From pipeline:
     `npx markdown-to-html-cli --source <SOURCE.md> --style=./style.css | wxcli draft add --title "<title>" --content - --thumb-media-id <thumb_media_id>`
   - Return output to the user.

---

## Failure Handling

- If `thumb_media_id` is missing, request it explicitly.
- If the CLI conversion fails, show the error and ask to retry.
- If `wxcli` fails, show the error output and ask whether to retry or adjust inputs.
