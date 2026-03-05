Feature: Authentication Middleware
  As an API endpoint
  I want to verify that requests have valid authentication tokens
  So that unauthorized users cannot access protected resources

  Scenario: Request without Authorization header is rejected
    Given an auth middleware protecting an endpoint
    When a request is made without an Authorization header
    Then the response status should be 401 Unauthorized

  Scenario: Request with invalid Bearer token is rejected
    Given an auth middleware protecting an endpoint
    When a request is made with an invalid Bearer token
    Then the response status should be 401 Unauthorized

  Scenario: Request without Bearer prefix is rejected
    Given an auth middleware protecting an endpoint
    When a request is made with Authorization header "some-token" without Bearer prefix
    Then the response status should be 401 Unauthorized
