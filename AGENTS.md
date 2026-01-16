# AIDU Agent Context

## Project Overview

This is the default AIDU agent context. Agents working on tasks should follow these guidelines
to ensure high-quality, maintainable outputs.

## General Principles

- Write clean, readable code that follows language-specific best practices
- Prioritize simplicity and clarity over cleverness
- Handle errors appropriately and gracefully
- Test your implementations thoroughly

## Coding Standards

### Variable and Function Naming

- Use clear, descriptive names that convey intent
- Follow language conventions (camelCase, snake_case, etc.)
- Avoid abbreviations unless widely understood

### Error Handling

- Always handle errors explicitly
- Provide helpful error messages
- Don't silently ignore failures

### Code Comments

- Add comments where the code's intent isn't obvious
- Avoid comments that just restate what the code does
- Document gotchas and non-obvious behavior
- Explain WHY, not WHAT

### Testing

- Write tests for all new functionality
- Ensure existing tests still pass
- Test edge cases and error conditions
- Aim for meaningful test coverage

## Language-Specific Guidelines

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Handle errors explicitly, never ignore them
- Use meaningful receiver names
- Keep functions focused and concise

### Python

- Follow PEP 8 style guide
- Use type hints where appropriate
- Write docstrings for functions and classes
- Use virtual environments for dependencies

### JavaScript/TypeScript

- Use modern ES6+ features
- Prefer const/let over var
- Use TypeScript for type safety when available
- Handle promises and async operations properly

## Architecture Considerations

- Keep components loosely coupled
- Follow SOLID principles
- Design for testability
- Consider scalability and maintainability

## Security

- Validate and sanitize all user input
- Never commit secrets or credentials
- Follow OWASP security guidelines
- Use parameterized queries for databases
- Implement proper authentication and authorization

## Documentation

- Write clear README files for projects
- Document API endpoints and their parameters
- Include usage examples
- Note any dependencies or prerequisites

## Success Criteria

Your work will be considered successful when:

1. **Functionality**: Code works as specified
2. **Quality**: Code is clean, well-structured, and follows best practices
3. **Tests**: Tests exist and pass
4. **Documentation**: Code is appropriately documented
5. **Errors**: Error handling is robust and informative

## When You Get Stuck

If you encounter issues:

1. Analyze error messages carefully
2. Check the logs for details
3. Verify your assumptions
4. Try a simpler approach
5. Report what you've tried and what isn't working

Remember: Your goal is to produce working, maintainable code that solves the task at hand.
