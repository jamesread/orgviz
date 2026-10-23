# AGENTS.md

orgviz draws organisation charts from YAML files. This guide is for AI agents and the people setting them up.

## Overview

- Each chart is one YAML file in the `orgs` directory, such as `orgs/Acme.yml`. The chart ID is the file name without `.yml`, and the chart is shown at `/chart/<chart ID>`.
- Profile photos live in the `avatars` directory and are served at `/avatars/<file>`.
- Shared themes live in the `themes` directory as YAML files. A chart uses one with `style.theme: <theme ID>`.
- The browser UI has a YAML editor pane that previews changes without saving them. Shared themes can be saved from the Theme and Style tab.

An agent creates or edits a chart over MCP. A person then opens the chart URL that each tool returns to review it.

## Discovery

| Path | What |
|---|---|
| `/llms.txt` | Short index of these integration points |
| `/openapi` | OpenAPI 3.1 JSON for the Connect RPC API |
| `/api` | Connect RPC API (read-only: list, get and preview charts) |
| `/mcp` | MCP server (Streamable HTTP) |

## MCP

- **URL:** `<orgviz base URL>/mcp`
- **Transport:** Streamable HTTP, stateless
- **Enabling:** off by default, when `/mcp` returns 404. Set `mcp.enabled: true` in `config.yaml`, or `ORGVIZ_MCP_ENABLED=true`.
- **Authentication:** none. Anyone who can reach `/mcp` can create and change charts and upload images, so only enable it on a trusted network.

Example client configuration:

```json
{
  "mcpServers": {
    "orgviz": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Tools

| Tool | Parameters | Behaviour |
|---|---|---|
| `orgviz_ping` | none | Returns the status and orgviz version. |
| `orgviz_yaml_reference` | none | Describes the org YAML format. Read this before writing YAML. |
| `orgviz_list_charts` | none | Lists charts with `chart_id`, `title`, number of `people` and browser `url`. |
| `orgviz_get_chart` | `chart_id` | Returns the chart's `yaml`, resolved `style`, `url` and `warnings`. |
| `orgviz_export_chart` | `chart_id` | Returns a PNG of the whole tree, using the resolved card features and avatars, plus the browser URL. |
| `orgviz_create_chart` | `chart_id`, `yaml` | Saves a new chart as `<chart_id>.yml`. Fails if the ID is taken or the YAML does not parse. |
| `orgviz_update_chart` | `chart_id`, `yaml` | Replaces an existing chart's YAML. Invalid YAML is rejected and the chart is left unchanged. |
| `orgviz_set_chart_style` | `chart_id`, optional `theme`, `background_color`, `css`, `avatars`, `job_titles`, `team`, `expand_buttons`, `compact` | Changes only the given parts of the chart's `style:` block. An empty `theme`, `background_color` or `css` removes it. `theme` alone clears local overrides so the shared theme applies. |
| `orgviz_list_themes` | none | Lists shared themes with `theme_id` and `title`. |
| `orgviz_get_theme` | `theme_id` | Returns one theme's title and style. |
| `orgviz_save_theme` | `theme_id`, optional `title`, `css`, `background_color`, `avatars`, `job_titles`, `team`, `expand_buttons`, `compact` | Creates or replaces `themes/<theme_id>.yml`. |
| `orgviz_upload_avatar` | `filename`, `image_base64`, optional `overwrite` | Saves a PNG, JPEG, WebP or SVG image (up to 2 MiB) to the avatars directory. Then set `avatar: <filename>` on the person. |
| `orgviz_list_avatars` | none | Lists the image files in the avatars directory. |

Chart IDs use letters, digits, `-` and `_`, up to 64 characters. Avatar file names are lowercase letters, digits, `-` and `_`, ending `.png`, `.jpg`, `.jpeg`, `.webp` or `.svg`, and the image content must match the extension. SVG must be a real `<svg>` document and must not contain `<script>`.

Chart results include `warnings` worth fixing, such as a `reports:` value that matches nobody (that person is drawn under the first person) or an `avatar:` file that has not been uploaded.

### Suggested flow

1. `orgviz_yaml_reference`, to learn the format.
2. `orgviz_upload_avatar` for each photo you have.
3. `orgviz_create_chart` with the YAML, using `avatar:` for uploaded photos.
4. Fix any `warnings` with `orgviz_update_chart`.
5. `orgviz_save_theme` if the look should be reused, then `orgviz_set_chart_style` with `theme`.
6. `orgviz_export_chart` to see the PNG, then give the person the `url` to review the live chart.

## Org YAML

```yaml
title: Example Company
people:
  - name: Alice Smith        # required and unique
    title: CEO
    team: Leadership
    alias: ali               # short name others can report to
    avatar: alice-smith.jpg  # optional uploaded photo
  - name: Bob Johnson
    title: Product Manager
    reports: ali             # manager's name or alias
style:                       # optional
  theme: ocean               # optional shared theme from the themes directory
  backgroundColor: "#f8fafc" # optional <main /> background color; overrides the theme
  features:
    avatars: true            # default true
    jobTitles: true          # default true
    team: false              # default false
    expandButtons: true      # default true; false shows the whole tree
    compact: true            # default true; false is an expanded tree
  css: |
    div.person { border-top-color: teal; }
    div.person.level-0 { border-top-width: 6px; }
```

Card CSS selectors: `div.person`, `div.person.level-0` (root; `level-1` is the next layer), `img.avatar`, `.person h3`, `.person .job-title` and `.person .team`.

## Connect RPC API

Defined in `proto/orgviz/clientapi/v1/clientapi.proto` and described at `/openapi`:

- `getClientInitialSettings`: version, the list of charts and shared themes.
- `getChart`: one chart's people, YAML source and resolved style.
- `previewChart`: parses YAML and returns the people and resolved style without saving.
- `listThemes` / `saveTheme`: list shared themes and create or replace one.

## Development

- `make proto` regenerates the Go, TypeScript and OpenAPI code after changing the proto file.
- `make test` and `make lint` run the Go and frontend checks. Go functions must stay under a cyclomatic complexity of 5.
- The MCP server is `service/internal/mcpserver`. It uses the same chart store (`service/internal/charts`) as the Connect API, which checks and writes files atomically.
