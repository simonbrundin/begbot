# Standards: API Consolidation

## Applicable Standards

### 1. API Design (`backend/api.md`)

- **RESTful Design**: Follow REST principles with clear resource-based URLs and appropriate HTTP methods (GET, POST, PUT, PATCH, DELETE)
- **Consistent Naming**: Use consistent, lowercase, hyphenated or underscored naming conventions for endpoints across the API
- **Plural Nouns**: Use plural nouns for resource endpoints (e.g., `/users`, `/products`) for consistency
- **Query Parameters**: Use query parameters for filtering, sorting, pagination, and search rather than creating separate endpoints
- **HTTP Status Codes**: Return appropriate, consistent HTTP status codes that accurately reflect the response (200, 201, 400, 404, 500, etc.)
- **Nested Resources**: Limit nesting depth to 2-3 levels maximum to keep URLs readable and maintainable

### 2. Error Handling (`global/error-handling.md`)

- **User-Friendly Messages**: Provide clear, actionable error messages to users without exposing technical details or security information
- **Fail Fast and Explicitly**: Validate input and check preconditions early; fail with clear error messages rather than allowing invalid state
- **Centralized Error Handling**: Handle errors at appropriate boundaries (controllers, API layers) rather than scattering try-catch blocks everywhere

### 3. Validation (`global/validation.md`)

- **Validate on Server Side**: Always validate on the server; never trust client-side validation alone for security or data integrity
- **Fail Early**: Validate input as early as possible and reject invalid data before processing
- **Specific Error Messages**: Provide clear, field-specific error messages that help users correct their input
- **Sanitize Input**: Sanitize user input to prevent injection attacks (SQL, XSS, command injection)

## Compliance Checklist

- [x] RESTful endpoints with proper HTTP methods
- [x] Plural nouns for resources
- [x] Query parameter support (`?mine=true`)
- [x] Appropriate status codes (201 for creation, 400 for bad input)
- [x] Centralized error handler middleware
- [x] Input validation on POST/PUT endpoints
- [x] Field-specific error messages in responses
