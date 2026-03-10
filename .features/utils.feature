Feature: Hjälpfunktioner
  Som en frontendapplikation
  Vill jag ha hjälpfunktioner för formatering och beräkning
  Så att jag kan visa data på ett konsekvent sätt

  Scenario: Formatera öre till kronor korrekt
    Given beloppet 10000 öre
    When jag formaterar det som valuta
    Then ska resultatet vara "100.00 kr"

  Scenario: Formatera litet belopp till kronor
    Given beloppet 500 öre
    When jag formaterar det som valuta
    Then ska resultatet vara "5.00 kr"

  Scenario: Formatera nollt belopp
    Given beloppet 0 öre
    When jag formaterar det som valuta
    Then ska resultatet vara "0.00 kr"

  Scenario: Formatera negativt belopp
    Given beloppet -5000 öre
    When jag formaterar det som valuta
    Then ska resultatet vara "-50.00 kr"

  Scenario: Formatera ISO-datum till svenskt format
    Given datumet "2024-01-15"
    When jag formaterar det som ett datum
    Then ska resultatet vara "2024-01-15"

  Scenario: Beräkna vinst korrekt
    Given ett handelsobjekt med:
      | field                  | value |
      | buy_price              | 5000  |
      | buy_shipping_cost      | 500   |
      | sell_price             | 8000  |
      | sell_packaging_cost    | 200   |
      | sell_postage_cost      | 100   |
      | sell_shipping_collected| 0     |
    When jag beräknar vinsten
    Then ska vinsten vara 2200

  Scenario: Beräkna vinst med saknade säljvärden
    Given ett handelsobjekt med köppris 5000 och inget säljpris
    When jag beräknar vinsten
    Then ska vinsten vara -5000
