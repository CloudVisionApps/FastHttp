# Config File Includes

FastHTTP supports including other JSON config files in the main configuration, similar to nginx's `include` directive.

## Usage

### Single File Include

```json
{
  "include": "vhosts.json",
  "user": "apache",
  "group": "apache",
  "listen": ["80"]
}
```

### Multiple Files Include (Array)

```json
{
  "include": ["vhosts.json", "mime-types.json", "admin-config.json"],
  "user": "apache",
  "group": "apache"
}
```

### Multiple Files Include (Using "includes" field)

```json
{
  "include": "mime-types.json",
  "includes": ["vhosts.json", "admin-config.json"],
  "user": "apache",
  "group": "apache"
}
```

**Note**: Both `include` and `includes` fields are supported. You can use:
- `include` with a single string: `"include": "file.json"`
- `include` with an array: `"include": ["file1.json", "file2.json"]`
- `includes` with an array: `"includes": ["file1.json", "file2.json"]`
- Mix both: `"include": "file1.json"` and `"includes": ["file2.json", "file3.json"]`

## How It Works

1. **Path Resolution**: 
   - Relative paths are resolved relative to the main config file's directory
   - Absolute paths are used as-is

2. **Merging Rules**:
   - **Arrays** (VirtualHosts, MimeTypes, Listen ports): Appended, duplicates avoided
   - **Scalar fields**: Included config values override base config if set
   - **Include order matters**: Later includes override earlier ones

3. **Circular Include Prevention**:
   - Maximum depth: 10 levels
   - Circular includes are detected and rejected

## Example Structure

**fasthttp.json** (main config):
```json
{
  "include": "vhosts.json",
  "user": "apache",
  "group": "apache",
  "listen": ["80"],
  "adminEnabled": true,
  "adminPort": "8080"
}
```

**vhosts.json** (included):
```json
{
  "virtualHosts": [
    {
      "serverName": "example.com",
      "documentRoot": "/var/www/example",
      "locations": []
    }
  ]
}
```

## Best Practices

1. **Separate concerns**: Put virtual hosts in separate files
2. **Use relative paths**: Makes configs portable
3. **Avoid deep nesting**: Keep include depth reasonable
4. **Document includes**: Comment what each included file contains
