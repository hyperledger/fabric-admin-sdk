


## Configuration justification
### tsconfig.json
- compilerOptions
  - `moduleResolution`: let typescript use native node modules (not for commonjs)

## Dependency justification
- `google-protobuf`: used for message `Timestamp`, which is excluded from fabric-protos
### Dev
 