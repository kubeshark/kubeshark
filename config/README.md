# Configuration schema versioning

The top-level `version` value identifies the CLI configuration schema. It is
independent of the Kubeshark release version: compatible application releases
must continue to use the same schema version.

## Compatibility rules

- Files without `version` predate schema versioning and are treated as schema
  version 1.
- The CLI loads only the schema version it explicitly supports. An incompatible
  file produces an error that points to `kubeshark config -r`.
- `kubeshark config -r` does not load the file it replaces, so regeneration also
  works for malformed or incompatible configurations.
- Writers always stamp the current version and never mutate the caller's config.
- The loader decodes once into a temporary copy and replaces the runtime config
  only after validation succeeds. Invalid files cannot leave partial changes.

## Changing the schema

Do not increment `currentConfigVersion` for additive or otherwise compatible
changes. Increment it only when loading the previous schema would be unsafe or
ambiguous. Keep the initial version unchanged so unversioned legacy files retain
their original meaning.

Keep compatibility checks at the file-loading boundary instead of scattering
version conditions across command code. Add table-driven tests for the current,
legacy, malformed, older, and newer formats whenever the policy changes.
