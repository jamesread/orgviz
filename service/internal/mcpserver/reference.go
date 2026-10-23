package mcpserver

const yamlReference = `# orgviz org YAML

One file per chart. The chart ID is the file name without .yml.

title: Example Company          # shown in the chart list

people:                         # in any order; the first person is the root
  - name: Alice Smith           # required; also used to match reports
    title: CEO                  # job title shown on the card
    team: Leadership            # shown on the card when style.features.team is true
    alias: ali                  # optional short name that reports can use
    avatar: alice-smith.jpg     # optional; png, jpg, jpeg, webp or svg uploaded with orgviz_upload_avatar
    attributes:                 # optional free-form key/value pairs (not drawn)
      location: London

  - name: Bob Johnson
    title: Product Manager
    reports: ali                # the manager's name or alias

style:                          # optional; applies to everyone viewing the chart
  theme: ocean                  # optional shared theme from orgviz_save_theme
  backgroundColor: "#f8fafc"    # optional <main /> background color; overrides the theme
  features:
    avatars: true               # profile photos (default true)
    jobTitles: true             # job titles (default true)
    team: false                 # team line (default false)
    expandButtons: true         # expand/collapse buttons (default true; false shows the whole tree)
    compact: true               # compact layout (default true; false is an expanded tree)
  css: |                        # optional custom CSS
    div.person { border-top-color: teal; }
    div.person.level-0 { border-top-width: 6px; }

Notes:
- Each person needs a unique name. reports must match another person's name or alias.
- People with no reports, or an unknown one, are shown under the first person.
- Without avatar, the photo is guessed as /avatars/<first name>.jpg, falling back to a
  default picture.
- Card CSS selectors: div.person, div.person.level-0 (root; level-1 is the next layer),
  img.avatar, .person h3, .person .job-title, .person .team.
- Use orgviz_set_chart_style to change style without rewriting the whole file.
- Use orgviz_export_chart to get a PNG of the whole tree.
- Shared themes are YAML files in the themes directory. Save one with orgviz_save_theme,
  then set style.theme or pass theme to orgviz_set_chart_style. Chart fields override the theme.
`
