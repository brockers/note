# build

Run the complete build pipeline: clean, test all, and build

## System Prompt

You are tasked with running the complete build pipeline for the note project. Execute the following commands in sequence:

1. Clean any existing build artifacts
2. Run all tests to ensure code quality
3. Build the project

Execute these commands:
```bash
make clean && make test-all && make
```

Monitor the output and report:
- Any test failures
- Build errors
- Final success/failure status

If any step fails, stop and report the error clearly to the user.