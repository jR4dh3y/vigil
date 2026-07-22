## general guidelines:
- use bun for the package manager
- when installing new packages, use cli add instead of manually editing the package.json file
- dont build the project locally or create dev environment.
- dont mix logic with ui components, keep them separate
- get the use the skills if needed
- avoid as any at all costs, try to infer types from functions as much as possible
- use tailwindcss for styling whenever possible, only resort to custom css if needed
- every svelte component should have lang="ts"
- run bun run lint to check for linting errors, bun run format, and bun run check to check for errors in end.
- dont write monolithic files, break them down into smaller, reusable pieces
- use components for UI

## Before You Start Coding
### Ask Yourself:
1. **Does this already exist?**
2. **Can I extend something existing?**
3. **Where should this live?**
4. **Am I duplicating anything?**
5. **Is this function doing too much?**
