# Marketing screenshots

The PNGs in this directory are the screenshots used in the project README and
docs. They are captured from a throwaway orgviz instance seeded with fictional
organisation files, so every run produces the same images.

orgviz has no database: charts and avatars come from `seed/`, copied into a
temporary config layout before the binary starts.

## Requirements

- Go and Node.js (to build `service/orgviz` and the frontend)
- [`repo-helper`](https://github.com/jamesread/repo-helper) with Chromium and a matching `chromedriver`
- `curl` and `jq`

No containers are needed.

## Regenerate

```bash
make                 # from this directory, or `make screenshots` from the repo root
```

This builds the service and frontend, starts orgviz on `127.0.0.1:18480` with
`-configdir seed` (fictional org YAML and avatars), captures every section of
`screenshots.ini`, writes `screenshot-manifest.yaml`, copies the docs images to
`docs/modules/ROOT/images/marketing/`, and stops everything again.

To browse the seeded instance yourself: `make up seed`, open
<http://127.0.0.1:18480/>, then `make down`.

## Adding a screenshot

Add a section to `screenshots.ini` with `url`, `title` and `description` (and
`theme = dark` for dark variants, which go last), then run `make`.
`screenshot-manifest.yaml` is generated; don't edit it by hand.
