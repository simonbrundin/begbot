Feature: Delete Listing
  As a user of the ads page
  I want to be able to delete listings
  So that I can remove unwanted or incorrect listings

  Background:
    Given I have a listing database

  Scenario: Successfully delete an existing listing
    Given a listing with id "1" exists in the database
    When I send a DELETE request to "/api/listings/1"
    Then the response status should be 204
    And the listing with id "1" should no longer exist in the database

  Scenario: Try to delete a non-existent listing
    Given no listing with id "999" exists in the database
    When I send a DELETE request to "/api/listings/999"
    Then the response status should be 404

  Scenario: Delete listing with valuations
    Given a listing with id "2" exists in the database
    And the listing has valuations
    When I send a DELETE request to "/api/listings/2"
    Then the response status should be 204
    And the listing with id "2" should no longer exist in the database
    And the related valuations should also be deleted

  Scenario: Delete listing with traded items
    Given a listing with id "3" exists in the database
    And the listing has traded items
    When I send a DELETE request to "/api/listings/3"
    Then the response status should be 204
    And the listing with id "3" should no longer exist in the database
    And the related traded items should also be deleted

  Scenario: Invalid listing ID format
    Given I have a listing database
    When I send a DELETE request to "/api/listings/invalid"
    Then the response status should be 400

  Scenario: Delete listing with image links
    Given a listing with id "4" exists in the database
    And the listing has image links
    When I send a DELETE request to "/api/listings/4"
    Then the response status should be 204
    And the listing with id "4" should no longer exist in the database
    And the related image links should also be deleted
