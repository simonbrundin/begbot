Feature: Autentiseringsmiddleware
  Som en API-endpoint
  Vill jag verifiera att förfrågningar har giltiga autentiseringstoken
  Så att obehöriga användare inte kan komma åt skyddade resurser

  Scenario: Förfrågan utan Authorization-header avvisas
    Given en autentiseringsmiddleware som skyddar en endpoint
    When en förfrågan görs utan Authorization-header
    Then ska svarsstatusen vara 401 Obehörig

  Scenario: Förfrågan med ogiltig Bearer-token avvisas
    Given en autentiseringsmiddleware som skyddar en endpoint
    When en förfrågan görs med en ogiltig Bearer-token
    Then ska svarsstatusen vara 401 Obehörig

  Scenario: Förfrågan utan Bearer-prefix avvisas
    Given en autentiseringsmiddleware som skyddar en endpoint
    When en förfrågan görs med Authorization-headern "some-token" utan Bearer-prefix
    Then ska svarsstatusen vara 401 Obehörig
