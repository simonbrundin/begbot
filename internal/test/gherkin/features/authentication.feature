Feature: Authentication
  As an API user
  I want all protected endpoints to require a valid JWT
  So that only authenticated users can access the API

  Background:
    Given an authentication middleware is configured

  Scenario: Request without Authorization header is rejected
    When a GET request is made to "/api/listings" without an Authorization header
    Then the response status should be 401
    And the response should contain error code "UNAUTHORIZED"

  Scenario: Request with token but missing Bearer prefix is rejected
    When a GET request is made to "/api/listings" with Authorization header "some-token"
    Then the response status should be 401
    And the response should contain error code "UNAUTHORIZED"

  Scenario: Request with invalid Bearer token is rejected
    When a GET request is made to "/api/listings" with Authorization header "Bearer invalid-token"
    Then the response status should be 401
    And the response should contain error code "UNAUTHORIZED"

  Scenario: GET /api/cron-jobs/status is accessible without authentication
    When a GET request is made to "/api/cron-jobs/status" without an Authorization header
    Then the response status should be 200

  Scenario: Valid JWT token grants access and sets user ID in context
    Given a valid JWT token signed with a test RSA key
    When a GET request is made to "/api/listings" with the valid Bearer token
    Then the response status should be 200
    And the user ID should be set in the request context
