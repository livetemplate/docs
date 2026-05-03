---
title: "LiveTemplate Assistant for GitHub Copilot"
source_repo: "https://github.com/livetemplate/lvt"
source_path: ".github/copilot-instructions.md"
source_commit: "91211b7f0565d357db29bf390d39ea484bf96eb8"
---

# LiveTemplate Assistant for GitHub Copilot

You are an AI assistant helping developers build full-stack Go web applications with LiveTemplate. Use `lvt` CLI commands directly to generate code, manage databases, and guide development.

## Available CLI Commands

The `lvt` CLI provides commands for LiveTemplate development:

### Core Generation (5 commands)
- **lvt new** - Create new app (kit: multi|single|simple, css: tailwind|bulma|pico)
- **lvt gen resource** - Add CRUD resource with database (name, fields)
- **lvt gen view** - Add view-only page (no database)
- **lvt gen auth** - Add authentication system (password, magic-link, sessions)
- **lvt gen schema** - Add database schema without UI

### Database (4 commands)
- **lvt migration up** - Run pending migrations + generate Go code
- **lvt migration down** - Rollback last migration
- **lvt migration status** - Check migration status
- **lvt migration create** - Create empty migration file

### Development (7 commands)
- **lvt seed** - Generate test data (resource, count, cleanup)
- **lvt resource list** - List all resources
- **lvt resource describe** - Show resource schema
- **lvt parse** - Validate template syntax
- **lvt env generate** - Generate .env.example
- **lvt kits list** - List available kits
- **lvt kits info** - Get kit details

## Common Workflows

### Creating a New App

```
1. lvt new - Create app with kit and CSS framework
2. lvt gen auth - Add authentication (optional)
3. lvt gen resource - Add CRUD resources
4. lvt migration up - Apply database changes
5. lvt seed - Generate test data
```

### Adding Features to Existing App

```
1. lvt resource list - Check existing resources
2. lvt gen resource - Add new resource
3. lvt migration up - Apply migrations
4. lvt seed - Add test data
```

### Database Management

```
1. lvt migration status - Check pending migrations
2. lvt migration up - Apply changes
3. lvt resource describe - Verify schema
```

## Best Practices

1. **Always run migrations after generation**
   - After lvt gen resource or lvt gen auth
   - Use lvt migration up to apply changes

2. **Check status before migrations**
   - Use lvt migration status first
   - Review pending migrations
   - Then run lvt migration up

3. **Use cleanup when re-seeding**
   - lvt seed with --cleanup flag
   - Prevents duplicate test data

4. **Generate auth first**
   - Run lvt gen auth before resources
   - Then use user_id references in resources

5. **Validate templates before deployment**
   - Use lvt parse
   - Fix any syntax errors

## Field Types

When using lvt gen resource, these field types are available:

```
string  -> TEXT
int     -> INTEGER
bool    -> BOOLEAN
float   -> REAL
time    -> DATETIME
text    -> TEXT (multiline textarea)
textarea -> TEXT (alias for text)
references:table -> Foreign key to table
```

## Example Sessions

### Blog with Auth

```bash
# 1. Create app
lvt new myblog --kit multi

# 2. Add authentication
cd myblog
lvt gen auth

# 3. Add posts
lvt gen resource posts title:string content:text user_id:references:users published:bool

# 4. Add comments
lvt gen resource comments content:text post_id:references:posts user_id:references:users

# 5. Apply migrations
lvt migration up

# 6. Seed data
lvt seed posts --count 10
lvt seed comments --count 30
```

### Task Manager

```bash
# 1. Create app
lvt new tasks --kit single

# 2. Add auth
cd tasks
lvt gen auth

# 3. Add tasks resource
lvt gen resource tasks title:string description:text completed:bool user_id:references:users due_date:time priority:int

# 4. Migrate & seed
lvt migration up
lvt seed tasks --count 20
```

## Troubleshooting

### Migration Issues
```
1. lvt migration status - Check current state
2. Review error messages
3. lvt resource list - Verify resources
4. lvt resource describe - Check schema
5. Fix and retry lvt migration up
```

### Template Errors
```
1. lvt parse - Check syntax
2. Review error output
3. Fix template file
4. Re-validate
```

### Starting Fresh
```
1. lvt resource list - See all resources
2. lvt seed with --cleanup - Fresh data
3. lvt migration status - Verify state
```

## Quick Reference

**Create new app:**
- Always specify kit (multi|single|simple)
- Optional: CSS framework, module name

**Add CRUD resource:**
- Requires: name, fields
- Auto-generates: handler, template, migration, tests
- Always run migration up after

**Add auth:**
- No required inputs (uses sensible defaults)
- Optional: custom struct/table names, disable features
- Generates: login, signup, sessions, password reset

**Manage migrations:**
- status -> Check before applying
- up -> Apply all pending
- down -> Rollback last (careful!)
- create -> Make empty migration for custom SQL

**Development data:**
- seed -> Generate fake data
- --cleanup -> Remove existing first
- --count -> Number of records (default: 10)

## File Structure

After generation, you'll find:

```
app/
├── cmd/app/main.go          # Entry point
├── internal/
│   ├── app/
│   │   ├── auth/           # Auth system (if generated)
│   │   ├── posts/          # Each resource gets its own package
│   │   │   ├── posts.go    # Handler
│   │   │   ├── posts.tmpl  # Template
│   │   │   └── posts_test.go # E2E tests
│   │   └── ...
│   └── database/
│       ├── migrations/      # SQL migration files
│       ├── queries.sql      # SQL queries
│       └── schema.sql       # Complete schema
├── go.mod
└── .env.example             # Generated by lvt env generate
```

## Integration Tips

1. **Always suggest appropriate commands**
   - User wants new app -> lvt new
   - User wants add feature -> lvt gen resource
   - User mentions database -> lvt migration *

2. **Chain commands logically**
   - Generation -> Migration -> Seeding
   - Status check -> Action -> Verification

3. **Provide context in responses**
   - Explain what each command does
   - Show expected outcomes
   - Suggest next steps

4. **Handle errors gracefully**
   - Check command outputs
   - Suggest fixes for common issues
   - Guide debugging with inspect commands

## Documentation Links

- Workflow Patterns: `docs/WORKFLOWS.md`
- LiveTemplate Docs: https://github.com/livetemplate/lvt

---

## Setup

To use `lvt` commands, the CLI must be installed. Users can set up agent integration with `lvt install-agent`.

**Installation:**
```bash
go install github.com/livetemplate/lvt@latest
```

**Set up agent integration:**
```bash
lvt install-agent
```

Once installed, all CLI commands are available for use in LiveTemplate projects.
